package stack

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/clients"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/validator"
)

//go:generate mockgen -source=stack_validator.go -destination=../../mocks/mock_stack_validator_dependencies.go -package=mocks
type organisationDomainService interface {
	ListByOrganisationID(ctx context.Context, orgID string) ([]*models.OrganisationDomain, *errors.ServiceError)
}

type secretService interface {
	ValidateImageRegistrySecretForStackResource(ctx context.Context, secretID string) *errors.ServiceError
	ValidateGitSecretForStackResource(ctx context.Context, secretID string) *errors.ServiceError
	ValidateSecretHasKeys(ctx context.Context, secretID string, requiredKeys []string) (bool, []string, *errors.ServiceError)
	ValidateSecretExists(ctx context.Context, secretID string) (bool, *errors.ServiceError)
	InternalGetByID(ctx context.Context, ID string) (*models.Secret, *errors.ServiceError)
}

type postgresAddonService interface {
	GetPostgresAddon(ctx context.Context, id string) (*models.PostgresAddon, *errors.ServiceError)
}

// Add only validations that take reasonable time to complete.
// Avoid validations that require network calls or long-running operations.
// Long running validations should be handled in the stack worker.
type stackValidator struct {
	interpolationValidator validator.InterpolationValidation
	domainService          organisationDomainService
	secretService          secretService
	postgresAddonService   postgresAddonService
}

type StackValidatorSpec struct {
	DomainService        organisationDomainService
	SecretService        secretService
	PostgresAddonService postgresAddonService
}

func NewStackValidator(
	spec StackValidatorSpec,
) validator.StackValidator {
	return &stackValidator{
		interpolationValidator: NewInterpolationValidation(),
		domainService:          spec.DomainService,
		secretService:          spec.SecretService,
		postgresAddonService:   spec.PostgresAddonService,
	}
}

func (v *stackValidator) ValidateForCreate(ctx context.Context, spec *models.Stack) *errors.ServiceError {
	if err := v.validateUniqueResourceNames(spec); err != nil {
		return err
	}
	if err := v.validateImageSource(spec); err != nil {
		return err
	}
	if err := v.validateStackEnvVars(spec); err != nil {
		return err
	}
	if err := v.validateStackPorts(spec); err != nil {
		return err
	}
	if err := v.validateVolumeMounts(spec); err != nil {
		return err
	}
	if err := v.validateDomainExistence(ctx, spec); err != nil {
		return err
	}
	if err := v.validateBuildSourceVolumes(spec); err != nil {
		return err
	}
	if err := v.validateConnections(ctx, nil, spec); err != nil {
		return err
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

	if err := v.validateUniqueResourceNames(spec); err != nil {
		return err
	}
	if err := v.validateImageSource(spec); err != nil {
		return err
	}
	if err := v.validateStackEnvVars(spec); err != nil {
		return err
	}
	if err := v.validateStackPorts(spec); err != nil {
		return err
	}
	if err := v.validateVolumeMounts(spec); err != nil {
		return err
	}
	if err := v.validateDomainExistence(ctx, spec); err != nil {
		return err
	}
	if err := v.validateBuildSourceVolumes(spec); err != nil {
		return err
	}
	if err := v.validateConnections(ctx, existing, spec); err != nil {
		return err
	}
	return nil
}

func (v *stackValidator) validateUniqueResourceNames(spec *models.Stack) *errors.ServiceError {
	seen := make(map[string]struct{}, len(spec.StackResources))
	for _, r := range spec.StackResources {
		if _, exists := seen[r.Name]; exists {
			return errors.BadRequest("duplicate stack resource name '%s'", r.Name)
		}
		seen[r.Name] = struct{}{}
	}
	return nil
}

func (v *stackValidator) validateImageSource(spec *models.Stack) *errors.ServiceError {
	for i := range spec.StackResources {
		currentResource := spec.StackResources[i]

		// Validate that resource has exactly one config type
		if err := v.validateResourceConfigType(currentResource); err != nil {
			return err
		}

		// Validate image config if present
		if currentResource.ImageConfig != nil {
			if err := v.validateImageConfig(currentResource); err != nil {
				return err
			}
		}

		// Validate build config if present
		if currentResource.BuildConfig != nil {
			if err := v.validateBuildConfig(currentResource); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *stackValidator) validateResourceConfigType(resource *models.StackResource) *errors.ServiceError {
	hasBuild := resource.BuildConfig != nil
	hasImage := resource.ImageConfig != nil

	if hasBuild && hasImage {
		return errors.BadRequest("stack resource '%s' cannot have both build and image config", resource.Name)
	}
	if !hasBuild && !hasImage {
		return errors.BadRequest("stack resource '%s' must have either build or image config", resource.Name)
	}
	return nil
}

func (v *stackValidator) validateImageConfig(resource *models.StackResource) *errors.ServiceError {
	if err := resource.ImageConfig.Validate(); err != nil {
		return errors.BadRequest("stack resource '%s' has invalid image config: %s", resource.Name, err.Error())
	}

	if resource.ImageConfig.PullSecretRef != nil {
		if err := v.validateImageWithSecret(resource); err != nil {
			return err
		}
	}

	return nil
}

func (v *stackValidator) validateImageWithSecret(resource *models.StackResource) *errors.ServiceError {
	secretRef := resource.ImageConfig.PullSecretRef

	if secretRef.SecretID == "" {
		return errors.BadRequest("stack resource '%s' has empty pull secret ID", resource.Name)
	}

	if err := v.secretService.ValidateImageRegistrySecretForStackResource(context.Background(), secretRef.SecretID); err != nil {
		return errors.BadRequest("stack resource '%s' has invalid pull secret: %s", resource.Name, err.Error())
	}

	return nil
}

func (v *stackValidator) validateImageAnonymously(resource *models.StackResource) *errors.ServiceError {
	client, err := clients.NewRegistryClientAnonymous()
	if err != nil {
		return errors.GeneralError("failed to create anonymous registry client for stack resource '%s': %s", resource.Name, err.Error())
	}

	return v.checkImageExists(client, resource, "does not exist or is not pullable")
}

func (v *stackValidator) checkImageExists(client clients.RegistryClient, resource *models.StackResource, errorSuffix string) *errors.ServiceError {
	exists, err := client.CheckImage(context.Background(), resource.ImageConfig.Image)
	if err != nil {
		return errors.GeneralError("failed to check image for stack resource '%s': %s", resource.Name, err.Error())
	}
	if !exists {
		return errors.BadRequest("stack resource '%s' image '%s' %s", resource.Name, resource.ImageConfig.Image, errorSuffix)
	}
	return nil
}

func (v *stackValidator) validateBuildConfig(resource *models.StackResource) *errors.ServiceError {
	if err := resource.BuildConfig.Validate(); err != nil {
		return errors.BadRequest("stack resource '%s' has invalid build config: %s", resource.Name, err.Error())
	}

	if resource.BuildConfig.SourceContext.Git != nil {
		return v.validateGitSource(resource)
	}

	return nil
}

func (v *stackValidator) validateGitSource(resource *models.StackResource) *errors.ServiceError {
	git := resource.BuildConfig.SourceContext.Git

	if git.GitSecretRef != nil {
		return v.validateGitWithSecret(resource)
	}

	return nil
}

func (v *stackValidator) validateGitWithSecret(resource *models.StackResource) *errors.ServiceError {
	secretRef := resource.BuildConfig.SourceContext.Git.GitSecretRef

	if secretRef.SecretID == "" {
		return errors.BadRequest("stack resource '%s' has empty git secret ID", resource.Name)
	}

	if err := v.secretService.ValidateGitSecretForStackResource(context.Background(), secretRef.SecretID); err != nil {
		return errors.BadRequest("stack resource '%s' has invalid git secret: %s", resource.Name, err.Error())
	}

	return nil
}

func (v *stackValidator) validateBuildSourceVolumes(spec *models.Stack) *errors.ServiceError {
	// Populate the volume mounts with the source volume names.
	definedVolumesMap := spec.VolumesMap()
	for i := range spec.StackResources {
		spec.StackResources[i].UserID = spec.UserID
		if spec.StackResources[i].BuildConfig != nil {
			buildConfig := spec.StackResources[i].BuildConfig
			if buildConfig.SourceContext.Volume != nil {
				volume, found := definedVolumesMap[buildConfig.SourceContext.Volume.SourceVolumeName]
				if !found {
					return errors.BadRequest("volume '%s' does not exist", buildConfig.SourceContext.Volume.SourceVolumeName)
				}
				buildConfig.SourceContext.Volume.SourceVolumeName = volume.Name
			}
		}
	}
	return nil
}

func (v *stackValidator) validateVolumeMounts(spec *models.Stack) *errors.ServiceError {
	if len(spec.Volumes) == 0 && spec.HasVolumeMounts() {
		return errors.BadRequest("stack '%s' has volume mounts but no volumes defined", spec.Name)
	}
	if !spec.HasVolumeMounts() {
		return nil
	}

	definedVolumes := spec.Volumes
	definedVolumesMap := make(map[string]*models.Volume)
	for i := range definedVolumes {
		definedVolumesMap[definedVolumes[i].Name] = definedVolumes[i]
	}

	for i := range spec.StackResources {
		currentResource := spec.StackResources[i]
		for j := range spec.StackResources[i].VolumeMounts {
			currentVolumeMount := currentResource.VolumeMounts[j]
			if _, found := definedVolumesMap[currentVolumeMount.SourceVolumeName]; !found {
				return errors.BadRequest("volume '%s' does not exist", currentVolumeMount.SourceVolumeName)
			}
		}
	}
	return nil
}

func (v *stackValidator) validateStackEnvVars(spec *models.Stack) *errors.ServiceError {
	for i := range spec.StackResources {
		currentResource := spec.StackResources[i]
		if currentResource.ExecutionConfig == nil || currentResource.ExecutionConfig.Env == nil {
			continue
		}
		currentEnvVars := currentResource.ExecutionConfig.Env
		allowedSelfOutputs := make(map[string]struct{}, len(currentResource.EnsureDeclaredOutputs()))
		for _, output := range currentResource.EnsureDeclaredOutputs() {
			allowedSelfOutputs[output.Name] = struct{}{}
		}
		keys := make(map[string]struct{})
		for _, envVar := range currentEnvVars {
			if len(envVar.Name) == 0 {
				return errors.BadRequest("stack resource '%s' has empty env var name", currentResource.Name)
			}
			hasValue := len(envVar.Value) > 0
			hasSelfOutput := len(envVar.SelfOutput) > 0
			if hasValue == hasSelfOutput {
				return errors.BadRequest("stack resource '%s' env var '%s' must set exactly one of value or self_output", currentResource.Name, envVar.Name)
			}
			if hasSelfOutput {
				if _, ok := allowedSelfOutputs[envVar.SelfOutput]; !ok {
					return errors.BadRequest("stack resource '%s' env var '%s' references unsupported self_output '%s'", currentResource.Name, envVar.Name, envVar.SelfOutput)
				}
			}
			if _, exists := keys[envVar.Name]; exists {
				return errors.BadRequest("stack resource '%s' has duplicate env var name '%s'", currentResource.Name, envVar.Name)
			}
			keys[envVar.Name] = struct{}{}
		}

	}

	if err := v.interpolationValidator.ValidateStackInterpolations(spec); err != nil {
		return errors.BadRequest("stack resource '%s' has invalid interpolation: %s", spec.Name, err.Error())
	}

	return nil
}

func (v *stackValidator) validateConnections(ctx context.Context, existing *models.Stack, spec *models.Stack) *errors.ServiceError {
	if len(spec.Connections) == 0 {
		return nil
	}

	resourceMap := spec.ResourcesMap()
	volumeMap := connectionVolumeMap(existing, spec)

	for i, connection := range spec.Connections {
		label := connectionLabel(connection, i)

		if err := validateConnectionKind(label, connection.Kind); err != nil {
			return err
		}
		if err := validateConnectionTargetResource(resourceMap, label, connection.To); err != nil {
			return err
		}

		sourceOutputs, err := v.validateConnectionSource(ctx, resourceMap, volumeMap, label, connection)
		if err != nil {
			return err
		}
		if err := validateConnectionMappings(label, connection, sourceOutputs); err != nil {
			return err
		}
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
		addon, err := v.validatePostgresConnectionConfig(ctx, label, connection)
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
		secret, serviceErr := v.secretService.InternalGetByID(ctx, connection.From.Id)
		if serviceErr != nil {
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
		"source_path":      {},
		"destination_path": {},
	}, label, "build_artifact_source"); err != nil {
		return err
	}
	sourcePath, _, err := getOptionalStringConfig(connection.Config, "source_path", label)
	if err != nil {
		return err
	}
	if sourcePath == "" {
		return errors.BadRequest("connection '%s' requires config.source_path for build artifact sources", label)
	}
	if _, _, err := getOptionalStringConfig(connection.Config, "destination_path", label); err != nil {
		return err
	}
	return nil
}

func (v *stackValidator) validatePostgresConnectionConfig(ctx context.Context, label string, connection models.StackConnection) (*models.PostgresAddon, *errors.ServiceError) {
	if connection.Kind != models.ConnectionKindEnv {
		return nil, errors.BadRequest("connection '%s' with from.type '%s' only supports kind '%s'", label, connection.From.Type, models.ConnectionKindEnv)
	}
	if connection.From.Id == "" {
		return nil, errors.BadRequest("connection '%s' is missing from.id for postgres addon source", label)
	}

	if err := validateConfigKeys(connection.Config, map[string]struct{}{
		"database":         {},
		"credential_scope": {},
		"superuser":        {},
	}, label, "postgres"); err != nil {
		return nil, err
	}

	database, _, err := getOptionalStringConfig(connection.Config, "database", label)
	if err != nil {
		return nil, err
	}

	credentialScope, hasCredentialScope, err := getOptionalStringConfig(connection.Config, "credential_scope", label)
	if err != nil {
		return nil, err
	}
	superuser, hasSuperuser, err := getOptionalBoolConfig(connection.Config, "superuser", label)
	if err != nil {
		return nil, err
	}
	if hasCredentialScope && hasSuperuser {
		return nil, errors.BadRequest("connection '%s' cannot set both config.credential_scope and config.superuser", label)
	}

	scope := "owner"
	if hasCredentialScope {
		switch credentialScope {
		case "owner", "superuser":
			scope = credentialScope
		default:
			return nil, errors.BadRequest("connection '%s' has unsupported postgres credential scope '%s'", label, credentialScope)
		}
	} else if hasSuperuser && superuser {
		scope = "superuser"
	}

	addon, serviceErr := v.postgresAddonService.GetPostgresAddon(ctx, connection.From.Id)
	if serviceErr != nil {
		return nil, errors.BadRequest("connection '%s' references non-existent postgres addon '%s'", label, connection.From.Id)
	}

	if scope == "superuser" {
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
		"mount_path": {},
		"sub_path":   {},
		"read_only":  {},
	}, label, "volume"); err != nil {
		return err
	}

	mountPath, _, err := getOptionalStringConfig(connection.Config, "mount_path", label)
	if err != nil {
		return err
	}
	if mountPath == "" {
		return errors.BadRequest("connection '%s' requires config.mount_path for volume mounts", label)
	}

	if _, _, err := getOptionalStringConfig(connection.Config, "sub_path", label); err != nil {
		return err
	}
	if _, _, err := getOptionalBoolConfig(connection.Config, "read_only", label); err != nil {
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

func (v *stackValidator) validateDomainExistence(ctx context.Context, spec *models.Stack) *errors.ServiceError {
	if !spec.HasExposedPorts() {
		return nil
	}

	orgDomains, err := v.domainService.ListByOrganisationID(ctx, spec.OrganisationID)
	if err != nil {
		return errors.GeneralError("failed to list domains for organisation '%s': %s", spec.OrganisationID, err.Error())
	}

	if len(orgDomains) == 0 {
		return errors.BadRequest("stack '%s' has publicly exposed ports but no domains defined for organisation '%s'", spec.Name, spec.OrganisationID)
	}
	return nil
}

func (v *stackValidator) validateStackPorts(spec *models.Stack) *errors.ServiceError {
	for i := range spec.StackResources {
		currentResource := spec.StackResources[i]
		if len(currentResource.Ports) == 0 {
			continue
		}
		currentPorts := currentResource.Ports
		portNames := make(map[string]struct{}, len(currentPorts))
		portNumbers := make(map[int]struct{}, len(currentPorts))

		for _, port := range currentPorts {
			if port.Number <= 0 {
				return errors.BadRequest("stack resource '%s' has invalid port number", currentResource.Name)
			}
			if port.Name == "" {
				return errors.BadRequest("stack resource '%s' has port %d missing name", currentResource.Name, port.Number)
			}
			if _, exists := portNames[port.Name]; exists {
				return errors.BadRequest("stack resource '%s' has duplicate port name '%s'", currentResource.Name, port.Name)
			}
			portNames[port.Name] = struct{}{}
			if _, exists := portNumbers[port.Number]; exists {
				return errors.BadRequest("stack resource '%s' has duplicate port number %d", currentResource.Name, port.Number)
			}
			portNumbers[port.Number] = struct{}{}
			if err := validatePortName(port.Name); err != nil {
				return errors.BadRequest("stack resource '%s' has invalid port name '%s': %s", currentResource.Name, port.Name, err.Error())
			}
		}
	}
	return nil
}

func validatePortName(name string) error {
	for i, r := range name {
		valid := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-'
		if !valid {
			return fmt.Errorf("must contain only lowercase letters, numbers, and hyphens")
		}
		if i == 0 && r == '-' {
			return fmt.Errorf("must start with a lowercase letter or number")
		}
	}
	if name[len(name)-1] == '-' {
		return fmt.Errorf("must end with a lowercase letter or number")
	}
	return nil
}
