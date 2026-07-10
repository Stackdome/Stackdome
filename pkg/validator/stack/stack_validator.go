package stack

import (
	"context"
	"fmt"
	"regexp"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stackdeploy"
	"github.com/Stackdome/stackdome/pkg/validator"
	"github.com/Stackdome/stackdome/pkg/validator/stackresource"
)

const (
	fieldName         = "name"
	fieldSpecSettings = "spec.settings"
)

//go:generate mockgen -source=stack_validator.go -destination=../../mocks/mock_stack_validator_dependencies.go -package=mocks
type secretService interface {
	ValidateSecretHasKeys(ctx context.Context, secretID string, requiredKeys []string) (bool, []string, *errors.ServiceError)
	ValidateSecretExists(ctx context.Context, secretID string) (bool, *errors.ServiceError)
	InternalGetByID(ctx context.Context, ID string) (*models.Secret, *errors.ServiceError)
}

type postgresAddonService interface {
	GetPostgresAddon(ctx context.Context, id string) (*models.PostgresAddon, *errors.ServiceError)
}

// stackValidator is network-free by construction: it runs only cheap,
// in-memory checks for whole-stack create/update — uniqueness, settings,
// connections, interpolations, and (via resourceValidator) every per-resource
// rule. Expensive checks (image pull, push access, git clone) run later, in
// the release worker's validation reconciler.
type stackValidator struct {
	secretService        secretService
	postgresAddonService postgresAddonService
	resourceValidator    stackresource.Validator
}

type StackValidatorSpec struct {
	SecretService        secretService
	PostgresAddonService postgresAddonService
	ResourceValidator    stackresource.Validator
}

func NewStackValidator(
	spec StackValidatorSpec,
) validator.StackValidator {
	if spec.ResourceValidator == nil {
		panic("stack.NewStackValidator: ResourceValidator is required")
	}
	return &stackValidator{
		secretService:        spec.SecretService,
		postgresAddonService: spec.PostgresAddonService,
		resourceValidator:    spec.ResourceValidator,
	}
}

func (v *stackValidator) ValidateForCreate(ctx context.Context, spec *models.Stack) *errors.ServiceError {
	var ferrs []errors.FieldError

	ferrs = append(ferrs, validateStackName(spec)...)
	ferrs = append(ferrs, v.validateUniqueResourceNames(spec)...)

	resourceErrs, serr := v.validateResources(ctx, spec)
	if serr != nil {
		return serr
	}
	ferrs = append(ferrs, resourceErrs...)

	ferrs = append(ferrs, validateStackSettings(spec)...)
	ferrs = append(ferrs, v.validateConnections(ctx, nil, spec)...)

	ferrs = dedupeFieldErrors(ferrs)
	if len(ferrs) > 0 {
		return errors.ValidationFailed(ferrs)
	}
	return nil
}

func (v *stackValidator) ValidateForUpdate(ctx context.Context, existing *models.Stack, spec *models.Stack) *errors.ServiceError {
	// Validate immutable fields
	if spec.Name != existing.Name {
		return errors.BadRequest("stack name cannot be updated")
	}
	if spec.UserID != existing.UserID {
		return errors.BadRequest("stack user cannot be updated")
	}
	if spec.OrganisationID != existing.OrganisationID {
		return errors.BadRequest("stack organisation cannot be updated")
	}

	var ferrs []errors.FieldError

	ferrs = append(ferrs, v.validateUniqueResourceNames(spec)...)

	resourceErrs, serr := v.validateResources(ctx, spec)
	if serr != nil {
		return serr
	}
	ferrs = append(ferrs, resourceErrs...)

	ferrs = append(ferrs, validateStackSettings(spec)...)
	ferrs = append(ferrs, v.validateConnections(ctx, existing, spec)...)

	ferrs = dedupeFieldErrors(ferrs)
	if len(ferrs) > 0 {
		return errors.ValidationFailed(ferrs)
	}
	return nil
}

// ValidateConnections runs only the connection-scoped rules over the full
// stack context (resource/volume/secret/postgres-addon lookups needed to
// validate a connection's endpoints), skipping validateResources,
// validateUniqueResourceNames, and validateStackSettings entirely. It backs
// connection-only mutations (create/update/delete a single connection) so a
// pre-existing, unrelated invalidity elsewhere in the stack can't block an
// edit the user has no way to fix from the connection form.
func (v *stackValidator) ValidateConnections(ctx context.Context, spec *models.Stack) *errors.ServiceError {
	ferrs := dedupeFieldErrors(v.validateConnections(ctx, nil, spec))
	if len(ferrs) > 0 {
		return errors.ValidationFailed(ferrs)
	}
	return nil
}

// ValidateShell runs only the rules scoped to the stack's own columns
// (validateStackSettings), skipping validateResources,
// validateUniqueResourceNames, and connection validation entirely. It backs
// thin shell update (PUT /stacks/{id}), which never carries children. Name
// rules are create-only (ValidateForCreate): the name is immutable on every
// update path, so re-validating it here would only brick updates of stacks
// created before the current name rules.
func (v *stackValidator) ValidateShell(_ context.Context, spec *models.Stack) *errors.ServiceError {
	ferrs := validateStackSettings(spec)
	ferrs = dedupeFieldErrors(ferrs)
	if len(ferrs) > 0 {
		return errors.ValidationFailed(ferrs)
	}
	return nil
}

// validateResources runs the shared per-resource validator (input shape,
// referential existence, sibling rules) over every resource in the stack,
// prefixing each field error with its resource's index. A non-nil
// ServiceError means a lookup failed for a reason other than not-found and
// validation was aborted.
func (v *stackValidator) validateResources(ctx context.Context, spec *models.Stack) ([]errors.FieldError, *errors.ServiceError) {
	var ferrs []errors.FieldError
	for i, resource := range spec.StackResources {
		siblings := make([]*models.StackResource, 0, len(spec.StackResources)-1)
		for j, other := range spec.StackResources {
			if i != j {
				siblings = append(siblings, other)
			}
		}
		resErrs, serr := v.resourceValidator.Validate(ctx, spec, resource, siblings)
		if serr != nil {
			return nil, serr
		}
		for _, fe := range resErrs {
			fe.Field = fmt.Sprintf("spec.stack_resources[%d].%s", i, fe.Field)
			ferrs = append(ferrs, fe)
		}
	}
	return ferrs, nil
}

// stackNamePattern is the RFC 1123 DNS-label charset: lowercase
// alphanumerics and '-', starting and ending with an alphanumeric. The stack
// name is embedded verbatim in the generated Kubernetes namespace name
// ("<stack-name>-<uuid>", truncated to the DNS-label cap), so anything
// outside this charset would make the namespace invalid and stall
// reconciliation at apply time instead of failing the request here, and
// anything longer than models.MaxStackNameLength would leave fewer than
// models.MinNamespaceUUIDSuffixLength UUID characters after truncation.
var stackNamePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

func validateStackName(spec *models.Stack) []errors.FieldError {
	if spec.Name == "" {
		return []errors.FieldError{{
			Field:   fieldName,
			Code:    errors.VErrStackNameInvalid,
			Message: "stack name is required",
		}}
	}
	if len(spec.Name) > models.MaxStackNameLength {
		return []errors.FieldError{{
			Field: fieldName,
			Code:  errors.VErrStackNameInvalid,
			Message: fmt.Sprintf(
				"stack name must be at most %d characters", models.MaxStackNameLength),
		}}
	}
	if !stackNamePattern.MatchString(spec.Name) {
		return []errors.FieldError{{
			Field:   fieldName,
			Code:    errors.VErrStackNameInvalid,
			Message: "stack name can only contain lowercase letters, numbers, and hyphens, and must start and end with a letter or number",
		}}
	}
	return nil
}

func validateStackSettings(spec *models.Stack) []errors.FieldError {
	if spec.Settings == nil {
		return nil
	}
	s := spec.Settings
	var errs []errors.FieldError
	if s.ReleaseRetentionLimit > models.MaxReleaseRetentionLimit {
		errs = append(errs, errors.FieldError{
			Field:   fieldSpecSettings,
			Code:    errors.VErrStackSettingsInvalid,
			Message: fmt.Sprintf("release_retention_limit must be at most %d", models.MaxReleaseRetentionLimit),
		})
	}
	if s.MinSuccessfulReleases > models.MaxMinSuccessfulReleases {
		errs = append(errs, errors.FieldError{
			Field:   fieldSpecSettings,
			Code:    errors.VErrStackSettingsInvalid,
			Message: fmt.Sprintf("min_successful_releases must be at most %d", models.MaxMinSuccessfulReleases),
		})
	}
	if s.DeployTimeoutMinutes > models.MaxDeployTimeoutMinutes {
		errs = append(errs, errors.FieldError{
			Field:   fieldSpecSettings,
			Code:    errors.VErrStackSettingsInvalid,
			Message: fmt.Sprintf("deploy_timeout_minutes must be at most %d", models.MaxDeployTimeoutMinutes),
		})
	}
	if s.MinSuccessfulReleases > 0 && s.ReleaseRetentionLimit > 0 && s.MinSuccessfulReleases > s.ReleaseRetentionLimit {
		errs = append(errs, errors.FieldError{
			Field:   fieldSpecSettings,
			Code:    errors.VErrStackSettingsInvalid,
			Message: fmt.Sprintf("min_successful_releases (%d) must not exceed release_retention_limit (%d)", s.MinSuccessfulReleases, s.ReleaseRetentionLimit),
		})
	}
	return errs
}

// dedupeFieldErrors collapses field errors that are identical in field, code,
// and message. Different validation rules on the fat path (e.g. whole-stack
// name uniqueness and per-resource sibling rules) can independently detect
// the same underlying problem and report it with matching text; this keeps
// the first occurrence and drops later exact repeats without needing to know
// which rule produced which error.
func dedupeFieldErrors(ferrs []errors.FieldError) []errors.FieldError {
	if len(ferrs) < 2 {
		return ferrs
	}
	type key struct {
		field, code, message string
	}
	seen := make(map[key]struct{}, len(ferrs))
	out := make([]errors.FieldError, 0, len(ferrs))
	for _, fe := range ferrs {
		k := key{fe.Field, fe.Code, fe.Message}
		if _, exists := seen[k]; exists {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, fe)
	}
	return out
}

func (v *stackValidator) validateUniqueResourceNames(spec *models.Stack) []errors.FieldError {
	var errs []errors.FieldError
	seen := make(map[string]struct{}, len(spec.StackResources))
	for i, r := range spec.StackResources {
		if _, exists := seen[r.Name]; exists {
			errs = append(errs, errors.FieldError{
				Field:   fmt.Sprintf("spec.stack_resources[%d].name", i),
				Code:    errors.VErrResourceNameDuplicate,
				Message: fmt.Sprintf("duplicate stack resource name '%s'", r.Name),
			})
		}
		seen[r.Name] = struct{}{}
	}
	return errs
}

// validateConnections checks every stack connection, collecting one field
// error per invalid connection rather than stopping at the first. existing is
// nil on create; on update it supplies the volumes already persisted for the
// stack so connections may reference either a pre-existing or a
// newly-bundled volume.
func (v *stackValidator) validateConnections(ctx context.Context, existing *models.Stack, spec *models.Stack) []errors.FieldError {
	if len(spec.Connections) == 0 {
		return nil
	}

	resourceMap := spec.ResourcesMap()
	volumeMap := connectionVolumeMap(existing, spec)

	var errs []errors.FieldError
	for i, connection := range spec.Connections {
		label := connectionLabel(connection, i)
		if serr := v.validateSingleConnection(ctx, spec.OrganisationID, resourceMap, volumeMap, label, connection); serr != nil {
			errs = append(errs, errors.FieldError{
				Field:   fmt.Sprintf("spec.connections[%d]", i),
				Code:    errors.VErrConnectionInvalid,
				Message: serr.Reason,
			})
		}
	}

	return errs
}

func (v *stackValidator) validateSingleConnection(
	ctx context.Context,
	orgID string,
	resourceMap map[string]*models.StackResource,
	volumeMap map[string]*models.Volume,
	label string,
	connection models.StackConnection,
) *errors.ServiceError {
	if err := validateConnectionKind(label, connection.Kind); err != nil {
		return err
	}
	if err := validateConnectionTargetResource(resourceMap, label, connection.To); err != nil {
		return err
	}

	sourceOutputs, err := v.validateConnectionSource(ctx, orgID, resourceMap, volumeMap, label, connection)
	if err != nil {
		return err
	}
	if err := validateConnectionMappings(label, connection, sourceOutputs); err != nil {
		return err
	}

	return nil
}

func validateConnectionKind(label string, kind models.ConnectionKind) *errors.ServiceError {
	switch kind {
	case models.ConnectionKindEnv,
		models.ConnectionKindVolumeMount,
		models.ConnectionKindBuildArtifactSource:
		return nil
	default:
		return errors.BadRequest("connection '%s' has unsupported kind '%s'", label, kind)
	}
}

func (v *stackValidator) validateConnectionSource(
	ctx context.Context,
	orgID string,
	resourceMap map[string]*models.StackResource,
	volumeMap map[string]*models.Volume,
	label string,
	connection models.StackConnection,
) ([]models.OutputDescriptor, *errors.ServiceError) {
	switch connection.From.Type {
	case models.TopologyNodeTypeStackResource:
		if connection.From.Name == "" {
			return nil, errors.BadRequest("connection '%s' is missing from.name for stack_resource source", label)
		}
		resource, ok := resourceMap[connection.From.Name]
		if !ok {
			return nil, errors.BadRequest("connection '%s' references unknown stack resource '%s'", label, connection.From.Name)
		}
		if connection.Kind == models.ConnectionKindBuildArtifactSource {
			return nil, validateBuildArtifactSourceConfig(volumeMap, label, connection)
		}
		if len(connection.Config) > 0 {
			return nil, errors.BadRequest("connection '%s' does not support config for from.type '%s'", label, connection.From.Type)
		}
		return resource.EnsureDeclaredOutputs(), nil
	case models.TopologyNodeTypePostgresAddon:
		addon, err := v.validatePostgresConnectionConfig(ctx, orgID, label, connection)
		if err != nil {
			return nil, err
		}
		return addon.EnsureDeclaredOutputs(), nil
	case models.TopologyNodeTypeSecret:
		if connection.From.Id == "" {
			return nil, errors.BadRequest("connection '%s' is missing from.id for secret source", label)
		}
		if len(connection.Config) > 0 {
			return nil, errors.BadRequest("connection '%s' does not support config for from.type '%s'", label, connection.From.Type)
		}
		// Org-scoped lookup: InternalGetByID itself is unscoped, so a secret
		// belonging to another organisation must behave exactly like a
		// missing one — anything else leaks cross-org secret existence.
		secret, serviceErr := v.secretService.InternalGetByID(ctx, connection.From.Id)
		if serviceErr != nil || secret.OrganisationID != orgID {
			return nil, errors.BadRequest("connection '%s' references non-existent secret '%s'", label, connection.From.Id)
		}
		return secret.EnsureDeclaredOutputs(), nil
	case models.TopologyNodeTypeVolume:
		if err := validateVolumeConnectionConfig(volumeMap, label, connection); err != nil {
			return nil, err
		}
		return nil, nil
	default:
		if len(connection.Config) > 0 {
			return nil, errors.BadRequest("connection '%s' does not support config for from.type '%s'", label, connection.From.Type)
		}
		return nil, nil
	}
}

func validateConnectionMappings(label string, connection models.StackConnection, sourceOutputs []models.OutputDescriptor) *errors.ServiceError {
	if connection.Kind != models.ConnectionKindEnv {
		return nil
	}

	allowedOutputs := make(map[string]struct{}, len(sourceOutputs))
	for _, output := range sourceOutputs {
		allowedOutputs[output.Name] = struct{}{}
	}

	for _, mapping := range connection.Mappings {
		if mapping.Target.Type != models.ConnectionTargetTypeEnv {
			return errors.BadRequest("connection '%s' env mappings only support target type '%s'", label, models.ConnectionTargetTypeEnv)
		}
		if mapping.Target.Name == "" {
			return errors.BadRequest("connection '%s' has env mapping with empty target name", label)
		}
		if err := validateValueRef(label, connection.From, mapping.Value, allowedOutputs); err != nil {
			return err
		}
	}

	return nil
}

func validateValueRef(label string, source models.TopologyNodeRef, valueRef models.ValueRef, allowedOutputs map[string]struct{}) *errors.ServiceError {
	hasOutput := valueRef.Output != ""
	hasTemplate := valueRef.Template != ""

	if !hasOutput && !hasTemplate {
		return errors.BadRequest("connection '%s' mapping must specify either 'output' or 'template'", label)
	}
	if hasOutput && hasTemplate {
		return errors.BadRequest("connection '%s' mapping must specify either 'output' or 'template', not both", label)
	}
	if hasTemplate && len(valueRef.Values) == 0 {
		return errors.BadRequest("connection '%s' mapping has template but no values", label)
	}

	if hasOutput {
		if _, ok := allowedOutputs[valueRef.Output]; !ok {
			return errors.BadRequest("connection '%s' references unsupported output '%s' for source '%s'", label, valueRef.Output, topologyNodeRefLabel(source))
		}
	}

	if hasTemplate {
		if err := stackdeploy.ValidateTemplateKeys(valueRef.Template, valueRef.Values); err != nil {
			return errors.BadRequest("connection '%s' has invalid template: %s", label, err.Error())
		}
	}

	for _, ref := range valueRef.Values {
		if _, ok := allowedOutputs[ref.Output]; !ok {
			return errors.BadRequest("connection '%s' references unsupported output '%s' for source '%s'", label, ref.Output, topologyNodeRefLabel(source))
		}
	}

	return nil
}

func topologyNodeRefLabel(ref models.TopologyNodeRef) string {
	if ref.Name != "" {
		return fmt.Sprintf("%s:%s", ref.Type, ref.Name)
	}
	if ref.Id != "" {
		return fmt.Sprintf("%s:%s", ref.Type, ref.Id)
	}
	return string(ref.Type)
}

func connectionVolumeMap(existing *models.Stack, spec *models.Stack) map[string]*models.Volume {
	if existing == nil {
		return spec.VolumesMap()
	}

	volumeMap := existing.VolumesMap()
	for name, volume := range spec.VolumesMap() {
		volumeMap[name] = volume
	}
	return volumeMap
}

func validateConnectionTargetResource(resourceMap map[string]*models.StackResource, label string, ref models.TopologyNodeRef) *errors.ServiceError {
	switch ref.Type {
	case models.TopologyNodeTypeStackResource:
		if ref.Name == "" {
			return errors.BadRequest("connection '%s' is missing to.name for stack_resource target", label)
		}
		if _, ok := resourceMap[ref.Name]; !ok {
			return errors.BadRequest("connection '%s' references unknown stack resource '%s'", label, ref.Name)
		}
	case models.TopologyNodeTypeVolume:
		if ref.Name == "" {
			return errors.BadRequest("connection '%s' is missing to.name for volume target", label)
		}
	}
	return nil
}

func validateBuildArtifactSourceConfig(volumeMap map[string]*models.Volume, label string, connection models.StackConnection) *errors.ServiceError {
	if connection.To.Type != models.TopologyNodeTypeVolume {
		return errors.BadRequest("connection '%s' with kind '%s' requires to.type '%s'", label, connection.Kind, models.TopologyNodeTypeVolume)
	}
	if _, ok := volumeMap[connection.To.Name]; !ok {
		return errors.BadRequest("connection '%s' references unknown volume '%s'", label, connection.To.Name)
	}
	if err := validateConfigKeys(connection.Config, map[string]struct{}{
		string(models.ConnectionConfigKeySourcePath):      {},
		string(models.ConnectionConfigKeyDestinationPath): {},
	}, label, "build_artifact_source"); err != nil {
		return err
	}
	sourcePath, _, err := getOptionalStringConfig(connection.Config, string(models.ConnectionConfigKeySourcePath), label)
	if err != nil {
		return err
	}
	if sourcePath == "" {
		return errors.BadRequest("connection '%s' requires config.source_path for build artifact sources", label)
	}
	if _, _, err := getOptionalStringConfig(connection.Config, string(models.ConnectionConfigKeyDestinationPath), label); err != nil {
		return err
	}
	return nil
}

func (v *stackValidator) validatePostgresConnectionConfig(ctx context.Context, orgID string, label string, connection models.StackConnection) (*models.PostgresAddon, *errors.ServiceError) {
	if connection.Kind != models.ConnectionKindEnv {
		return nil, errors.BadRequest("connection '%s' with from.type '%s' only supports kind '%s'", label, connection.From.Type, models.ConnectionKindEnv)
	}
	if connection.From.Id == "" {
		return nil, errors.BadRequest("connection '%s' is missing from.id for postgres addon source", label)
	}

	if err := validateConfigKeys(connection.Config, map[string]struct{}{
		string(models.ConnectionConfigKeyDatabase):        {},
		string(models.ConnectionConfigKeyCredentialScope): {},
		string(models.ConnectionConfigKeySuperuser):       {},
	}, label, "postgres"); err != nil {
		return nil, err
	}

	database, _, err := getOptionalStringConfig(connection.Config, string(models.ConnectionConfigKeyDatabase), label)
	if err != nil {
		return nil, err
	}

	credentialScope, hasCredentialScope, err := getOptionalStringConfig(connection.Config, string(models.ConnectionConfigKeyCredentialScope), label)
	if err != nil {
		return nil, err
	}
	superuser, hasSuperuser, err := getOptionalBoolConfig(connection.Config, string(models.ConnectionConfigKeySuperuser), label)
	if err != nil {
		return nil, err
	}
	if hasCredentialScope && hasSuperuser {
		return nil, errors.BadRequest("connection '%s' cannot set both config.credential_scope and config.superuser", label)
	}

	scope := models.CredentialScopeOwner
	if hasCredentialScope {
		switch credentialScope {
		case models.CredentialScopeOwner, models.CredentialScopeSuperuser:
			scope = credentialScope
		default:
			return nil, errors.BadRequest("connection '%s' has unsupported postgres credential scope '%s'", label, credentialScope)
		}
	} else if hasSuperuser && superuser {
		scope = models.CredentialScopeSuperuser
	}

	// Org-scoped lookup: GetPostgresAddon itself is unscoped, so an addon
	// belonging to another organisation must behave exactly like a missing
	// one — anything else leaks cross-org addon existence.
	addon, serviceErr := v.postgresAddonService.GetPostgresAddon(ctx, connection.From.Id)
	if serviceErr != nil || addon.OrganisationID != orgID {
		return nil, errors.BadRequest("connection '%s' references non-existent postgres addon '%s'", label, connection.From.Id)
	}

	if scope == models.CredentialScopeSuperuser {
		if !addon.Configuration.EnableSuperuserAccess {
			return nil, errors.BadRequest("connection '%s' requests superuser access but addon '%s' does not have superuser access enabled", label, connection.From.Id)
		}
		return addon, nil
	}

	if database == "" {
		return nil, errors.BadRequest("connection '%s' requires config.database when postgres credential scope is owner", label)
	}
	if !addon.HasDatabase(database) {
		return nil, errors.BadRequest("connection '%s' references non-existent database '%s' in postgres addon '%s'", label, database, connection.From.Id)
	}

	return addon, nil
}

func validateVolumeConnectionConfig(volumeMap map[string]*models.Volume, label string, connection models.StackConnection) *errors.ServiceError {
	if connection.Kind != models.ConnectionKindVolumeMount {
		return errors.BadRequest("connection '%s' with from.type '%s' only supports kind '%s'", label, connection.From.Type, models.ConnectionKindVolumeMount)
	}
	if connection.From.Name == "" {
		return errors.BadRequest("connection '%s' is missing from.name for volume source", label)
	}
	if _, ok := volumeMap[connection.From.Name]; !ok {
		return errors.BadRequest("connection '%s' references unknown volume '%s'", label, connection.From.Name)
	}

	if err := validateConfigKeys(connection.Config, map[string]struct{}{
		string(models.ConnectionConfigKeyMountPath): {},
		string(models.ConnectionConfigKeySubPath):   {},
		string(models.ConnectionConfigKeyReadOnly):  {},
	}, label, "volume"); err != nil {
		return err
	}

	mountPath, _, err := getOptionalStringConfig(connection.Config, string(models.ConnectionConfigKeyMountPath), label)
	if err != nil {
		return err
	}
	if mountPath == "" {
		return errors.BadRequest("connection '%s' requires config.mount_path for volume mounts", label)
	}

	if _, _, err := getOptionalStringConfig(connection.Config, string(models.ConnectionConfigKeySubPath), label); err != nil {
		return err
	}
	if _, _, err := getOptionalBoolConfig(connection.Config, string(models.ConnectionConfigKeyReadOnly), label); err != nil {
		return err
	}

	return nil
}

func validateConfigKeys(config map[string]interface{}, allowed map[string]struct{}, label string, configType string) *errors.ServiceError {
	for key := range config {
		if _, ok := allowed[key]; !ok {
			return errors.BadRequest("connection '%s' has unsupported %s config key '%s'", label, configType, key)
		}
	}
	return nil
}

func getOptionalStringConfig(config map[string]interface{}, key string, label string) (string, bool, *errors.ServiceError) {
	value, ok := config[key]
	if !ok {
		return "", false, nil
	}
	asString, ok := value.(string)
	if !ok {
		return "", false, errors.BadRequest("connection '%s' config.%s must be a string", label, key)
	}
	return asString, true, nil
}

func getOptionalBoolConfig(config map[string]interface{}, key string, label string) (bool, bool, *errors.ServiceError) {
	value, ok := config[key]
	if !ok {
		return false, false, nil
	}
	asBool, ok := value.(bool)
	if !ok {
		return false, false, errors.BadRequest("connection '%s' config.%s must be a boolean", label, key)
	}
	return asBool, true, nil
}

func connectionLabel(connection models.StackConnection, index int) string {
	if connection.ID != "" {
		return connection.ID
	}
	return fmt.Sprintf("#%d", index)
}
