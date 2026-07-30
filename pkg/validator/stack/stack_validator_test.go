package stack

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/validator"
	"go.uber.org/mock/gomock"
)

// requireSingleFieldError extracts the sole aggregated field error from a
// ValidationFailed ServiceError, failing the test if err is nil, isn't a
// validation error, or carries anything other than exactly one field error.
func requireSingleFieldError(t *testing.T, err *errors.ServiceError) errors.FieldError {
	t.Helper()
	if err == nil {
		t.Fatal("expected a validation error, got nil")
	}
	details, ok := err.Details.(errors.ValidationErrorDetails)
	if !ok {
		t.Fatalf("expected errors.ValidationErrorDetails, got %#v", err.Details)
	}
	if len(details.Errors) != 1 {
		t.Fatalf("expected exactly 1 field error, got %d: %#v", len(details.Errors), details.Errors)
	}
	return details.Errors[0]
}

func fieldErrors(t *testing.T, err *errors.ServiceError) []errors.FieldError {
	t.Helper()
	if err == nil {
		t.Fatal("expected a validation error, got nil")
	}
	details, ok := err.Details.(errors.ValidationErrorDetails)
	if !ok {
		t.Fatalf("expected errors.ValidationErrorDetails, got %#v", err.Details)
	}
	return details.Errors
}

func TestValidateForCreateRejectsSecretMountConnectionKind(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithConnections(models.StackConnection{
		ID:   "secret-files",
		Kind: models.ConnectionKind("secret_mount"),
		From: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeSecret,
			Id:   "sec-1",
		},
		To: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeStackResource,
			Name: "web",
		},
	})

	err := v.ValidateForCreate(context.Background(), spec)
	fe := requireSingleFieldError(t, err)
	if got, want := fe.Message, "connection 'secret-files' has unsupported kind 'secret_mount'"; got != want {
		t.Fatalf("unexpected message: got %q want %q", got, want)
	}
	if fe.Code != errors.VErrConnectionInvalid {
		t.Fatalf("unexpected code: got %q want %q", fe.Code, errors.VErrConnectionInvalid)
	}
}

func TestValidateForCreateAllowsPostgresConnectionConfig(t *testing.T) {
	v, postgresAddons := newValidatorWithMockedPostgresAddonService(t)
	spec := stackWithConnections(models.StackConnection{
		ID:   "pg-env",
		Kind: models.ConnectionKindEnv,
		From: models.TopologyNodeRef{
			Type: models.TopologyNodeTypePostgresAddon,
			Id:   "pg-1",
		},
		To: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeStackResource,
			Name: "web",
		},
		Config: map[string]interface{}{
			"database": "app",
		},
	})
	postgresAddons.EXPECT().GetPostgresAddon(gomock.Any(), "pg-1").Return(&models.PostgresAddon{
		ID:             "pg-1",
		OrganisationID: "org-1",
		Databases: []models.PostgresAddonDatabase{
			{Name: "app"},
		},
	}, nil)

	err := v.ValidateForCreate(context.Background(), spec)
	if err != nil {
		t.Fatalf("expected postgres connection config to validate, got %v", err)
	}
}

func TestValidateForCreateAllowsPostgresSuperuserConnectionConfig(t *testing.T) {
	v, postgresAddons := newValidatorWithMockedPostgresAddonService(t)
	spec := stackWithConnections(models.StackConnection{
		ID:   "pg-env",
		Kind: models.ConnectionKindEnv,
		From: models.TopologyNodeRef{
			Type: models.TopologyNodeTypePostgresAddon,
			Id:   "pg-1",
		},
		To: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeStackResource,
			Name: "web",
		},
		Config: map[string]interface{}{
			"credential_scope": "superuser",
		},
	})
	postgresAddons.EXPECT().GetPostgresAddon(gomock.Any(), "pg-1").Return(&models.PostgresAddon{
		ID:             "pg-1",
		OrganisationID: "org-1",
		Configuration: models.PostgresConfiguration{
			EnableSuperuserAccess: true,
		},
	}, nil)

	err := v.ValidateForCreate(context.Background(), spec)
	if err != nil {
		t.Fatalf("expected postgres superuser connection config to validate, got %v", err)
	}
}

func TestValidateForCreateRejectsPostgresConnectionConfigWithoutDatabase(t *testing.T) {
	v, postgresAddons := newValidatorWithMockedPostgresAddonService(t)
	spec := stackWithConnections(models.StackConnection{
		ID:   "pg-env",
		Kind: models.ConnectionKindEnv,
		From: models.TopologyNodeRef{
			Type: models.TopologyNodeTypePostgresAddon,
			Id:   "pg-1",
		},
		To: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeStackResource,
			Name: "web",
		},
	})
	postgresAddons.EXPECT().GetPostgresAddon(gomock.Any(), "pg-1").Return(&models.PostgresAddon{ID: "pg-1", OrganisationID: "org-1"}, nil)

	err := v.ValidateForCreate(context.Background(), spec)
	fe := requireSingleFieldError(t, err)
	if got, want := fe.Message, "connection 'pg-env' requires config.database when postgres credential scope is owner"; got != want {
		t.Fatalf("unexpected message: got %q want %q", got, want)
	}
}

// TestValidateForCreateRejectsPostgresConnectionFromAnotherOrganisation asserts
// the connection-source postgres addon lookup is org-scoped: an addon that
// exists but belongs to a different organisation behaves exactly like a
// missing one, so cross-org addon existence never leaks through connection
// validation.
func TestValidateForCreateRejectsPostgresConnectionFromAnotherOrganisation(t *testing.T) {
	v, postgresAddons := newValidatorWithMockedPostgresAddonService(t)
	spec := stackWithConnections(models.StackConnection{
		ID:   "pg-env",
		Kind: models.ConnectionKindEnv,
		From: models.TopologyNodeRef{
			Type: models.TopologyNodeTypePostgresAddon,
			Id:   "pg-1",
		},
		To: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeStackResource,
			Name: "web",
		},
		Config: map[string]interface{}{
			string(models.ConnectionConfigKeyDatabase): "app",
		},
	})
	postgresAddons.EXPECT().GetPostgresAddon(gomock.Any(), "pg-1").Return(&models.PostgresAddon{
		ID:             "pg-1",
		OrganisationID: "org-other",
		Databases:      []models.PostgresAddonDatabase{{Name: "app"}},
	}, nil)

	err := v.ValidateForCreate(context.Background(), spec)
	fe := requireSingleFieldError(t, err)
	if fe.Code != errors.VErrConnectionInvalid {
		t.Fatalf("expected %s, got %s", errors.VErrConnectionInvalid, fe.Code)
	}
	if got, want := fe.Message, "connection 'pg-env' references non-existent postgres addon 'pg-1'"; got != want {
		t.Fatalf("unexpected message: got %q want %q", got, want)
	}
}

func TestValidateForCreateRejectsUnknownPostgresConnectionConfigKey(t *testing.T) {
	v, _ := newValidatorWithMockedPostgresAddonService(t)
	spec := stackWithConnections(models.StackConnection{
		ID:   "pg-env",
		Kind: models.ConnectionKindEnv,
		From: models.TopologyNodeRef{
			Type: models.TopologyNodeTypePostgresAddon,
			Id:   "pg-1",
		},
		To: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeStackResource,
			Name: "web",
		},
		Config: map[string]interface{}{
			"database": "app",
			"oops":     true,
		},
	})

	err := v.ValidateForCreate(context.Background(), spec)
	fe := requireSingleFieldError(t, err)
	if got, want := fe.Message, "connection 'pg-env' has unsupported postgres config key 'oops'"; got != want {
		t.Fatalf("unexpected message: got %q want %q", got, want)
	}
}

func TestValidateForCreateAllowsVolumeMountConnectionConfig(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithConnections(models.StackConnection{
		ID:   "volume-mount",
		Kind: models.ConnectionKindVolumeMount,
		From: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeVolume,
			Name: "uploads",
		},
		To: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeStackResource,
			Name: "web",
		},
		Config: map[string]interface{}{
			"mount_path": "/uploads",
			"sub_path":   "",
			"read_only":  false,
		},
	})
	spec.Volumes = []*models.Volume{
		{Name: "uploads"},
	}

	err := v.ValidateForCreate(context.Background(), spec)
	if err != nil {
		t.Fatalf("expected volume mount connection config to validate, got %v", err)
	}
}

func TestValidateForCreateRejectsVolumeMountConnectionWithoutMountPath(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithConnections(models.StackConnection{
		ID:   "volume-mount",
		Kind: models.ConnectionKindVolumeMount,
		From: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeVolume,
			Name: "uploads",
		},
		To: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeStackResource,
			Name: "web",
		},
	})
	spec.Volumes = []*models.Volume{
		{Name: "uploads"},
	}

	err := v.ValidateForCreate(context.Background(), spec)
	fe := requireSingleFieldError(t, err)
	if got, want := fe.Message, "connection 'volume-mount' requires config.mount_path for volume mounts"; got != want {
		t.Fatalf("unexpected message: got %q want %q", got, want)
	}
}

func TestValidateForCreateRejectsVolumeMountConnectionWithInvalidReadOnlyType(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithConnections(models.StackConnection{
		ID:   "volume-mount",
		Kind: models.ConnectionKindVolumeMount,
		From: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeVolume,
			Name: "uploads",
		},
		To: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeStackResource,
			Name: "web",
		},
		Config: map[string]interface{}{
			"mount_path": "/uploads",
			"read_only":  "nope",
		},
	})
	spec.Volumes = []*models.Volume{
		{Name: "uploads"},
	}

	err := v.ValidateForCreate(context.Background(), spec)
	fe := requireSingleFieldError(t, err)
	if got, want := fe.Message, "connection 'volume-mount' config.read_only must be a boolean"; got != want {
		t.Fatalf("unexpected message: got %q want %q", got, want)
	}
}

func TestValidateForCreateAllowsStackResourceEnvConnectionUsingDeclaredOutput(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithConnections(models.StackConnection{
		ID:   "internal-api",
		Kind: models.ConnectionKindEnv,
		From: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeStackResource,
			Name: "web",
		},
		To: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeStackResource,
			Name: "web",
		},
		Mappings: []models.ConnectionMapping{
			{
				Target: models.ConnectionTarget{
					Type: models.ConnectionTargetTypeEnv,
					Name: "SELF_URL",
				},
				Value: models.ValueRef{
					Output: models.OutputNameURL,
				},
			},
		},
	})

	err := v.ValidateForCreate(context.Background(), spec)
	if err != nil {
		t.Fatalf("expected stack resource connection output to validate, got %v", err)
	}
}

func TestValidateForCreateRejectsUnknownStackResourceConnectionOutput(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithConnections(models.StackConnection{
		ID:   "internal-api",
		Kind: models.ConnectionKindEnv,
		From: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeStackResource,
			Name: "web",
		},
		To: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeStackResource,
			Name: "web",
		},
		Mappings: []models.ConnectionMapping{
			{
				Target: models.ConnectionTarget{
					Type: models.ConnectionTargetTypeEnv,
					Name: "SELF_URL",
				},
				Value: models.ValueRef{
					Output: "url.grpc",
				},
			},
		},
	})

	err := v.ValidateForCreate(context.Background(), spec)
	fe := requireSingleFieldError(t, err)
	if got, want := fe.Message, "connection 'internal-api' references unsupported output 'url.grpc' for source 'stack_resource:web'"; got != want {
		t.Fatalf("unexpected message: got %q want %q", got, want)
	}
}

func TestValidateForCreateAllowsSecretConnectionUsingBracketAccessor(t *testing.T) {
	v, secrets := newValidatorWithMockedSecretService(t)
	spec := stackWithConnections(models.StackConnection{
		ID:   "tls-cert",
		Kind: models.ConnectionKindEnv,
		From: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeSecret,
			Id:   "sec-1",
		},
		To: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeStackResource,
			Name: "web",
		},
		Mappings: []models.ConnectionMapping{
			{
				Target: models.ConnectionTarget{
					Type: models.ConnectionTargetTypeEnv,
					Name: "TLS_CERT",
				},
				Value: models.ValueRef{
					Output: "tls.crt",
				},
			},
		},
	})
	secrets.EXPECT().InternalGetByID(gomock.Any(), "sec-1").Return(&models.Secret{
		ID:             "sec-1",
		OrganisationID: "org-1",
		Keys:           []string{"tls.crt"},
	}, nil)

	err := v.ValidateForCreate(context.Background(), spec)
	if err != nil {
		t.Fatalf("expected secret connection output to validate, got %v", err)
	}
}

// TestValidateForCreateRejectsSecretConnectionFromAnotherOrganisation asserts
// the connection-source secret lookup is org-scoped: a secret that exists but
// belongs to a different organisation behaves exactly like a missing one, so
// cross-org secret existence never leaks through connection validation.
func TestValidateForCreateRejectsSecretConnectionFromAnotherOrganisation(t *testing.T) {
	v, secrets := newValidatorWithMockedSecretService(t)
	spec := stackWithConnections(models.StackConnection{
		ID:   "tls-cert",
		Kind: models.ConnectionKindEnv,
		From: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeSecret,
			Id:   "sec-1",
		},
		To: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeStackResource,
			Name: "web",
		},
		Mappings: []models.ConnectionMapping{
			{
				Target: models.ConnectionTarget{
					Type: models.ConnectionTargetTypeEnv,
					Name: "TLS_CERT",
				},
				Value: models.ValueRef{
					Output: "tls.crt",
				},
			},
		},
	})
	secrets.EXPECT().InternalGetByID(gomock.Any(), "sec-1").Return(&models.Secret{
		ID:             "sec-1",
		OrganisationID: "org-other",
		Keys:           []string{"tls.crt"},
	}, nil)

	err := v.ValidateForCreate(context.Background(), spec)
	fe := requireSingleFieldError(t, err)
	if fe.Code != errors.VErrConnectionInvalid {
		t.Fatalf("expected %s, got %s", errors.VErrConnectionInvalid, fe.Code)
	}
	if got, want := fe.Message, "connection 'tls-cert' references non-existent secret 'sec-1'"; got != want {
		t.Fatalf("unexpected message: got %q want %q", got, want)
	}
}

func TestValidateForUpdateAllowsVolumeMountConnectionUsingExistingDBVolume(t *testing.T) {
	v := newTestValidator(t)
	existing := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	existing.Volumes = []*models.Volume{
		{Name: "uploads"},
	}

	patch := stackWithConnections(models.StackConnection{
		ID:   "volume-mount",
		Kind: models.ConnectionKindVolumeMount,
		From: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeVolume,
			Name: "uploads",
		},
		To: models.TopologyNodeRef{
			Type: models.TopologyNodeTypeStackResource,
			Name: "web",
		},
		Config: map[string]interface{}{
			"mount_path": "/uploads",
		},
	})
	patch.Name = existing.Name
	patch.UserID = existing.UserID
	patch.OrganisationID = existing.OrganisationID

	err := v.ValidateForUpdate(context.Background(), existing, patch)
	if err != nil {
		t.Fatalf("expected existing DB volume to validate, got %v", err)
	}
}

func TestValidateForCreateRejectsRetentionLimitAboveMax(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.Settings = &models.StackSettings{ReleaseRetentionLimit: models.MaxReleaseRetentionLimit + 1}

	err := v.ValidateForCreate(context.Background(), spec)
	fe := requireSingleFieldError(t, err)
	if got, want := fe.Message, "release_retention_limit must be at most 50"; got != want {
		t.Fatalf("unexpected message: got %q want %q", got, want)
	}
	if fe.Code != errors.VErrStackSettingsInvalid {
		t.Fatalf("unexpected code: got %q want %q", fe.Code, errors.VErrStackSettingsInvalid)
	}
}

func TestValidateForCreateRejectsMinSuccessfulAboveMax(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.Settings = &models.StackSettings{MinSuccessfulReleases: models.MaxMinSuccessfulReleases + 1}

	err := v.ValidateForCreate(context.Background(), spec)
	fe := requireSingleFieldError(t, err)
	if got, want := fe.Message, "min_successful_releases must be at most 20"; got != want {
		t.Fatalf("unexpected message: got %q want %q", got, want)
	}
}

func TestValidateForCreateRejectsDeployTimeoutAboveMax(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.Settings = &models.StackSettings{DeployTimeoutMinutes: models.MaxDeployTimeoutMinutes + 1}

	err := v.ValidateForCreate(context.Background(), spec)
	fe := requireSingleFieldError(t, err)
	if got, want := fe.Message, "deploy_timeout_minutes must be at most 120"; got != want {
		t.Fatalf("unexpected message: got %q want %q", got, want)
	}
}

func TestValidateForCreateRejectsMinSuccessfulExceedingRetention(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.Settings = &models.StackSettings{
		ReleaseRetentionLimit: 5,
		MinSuccessfulReleases: 10,
	}

	err := v.ValidateForCreate(context.Background(), spec)
	fe := requireSingleFieldError(t, err)
	if got, want := fe.Message, "min_successful_releases (10) must not exceed release_retention_limit (5)"; got != want {
		t.Fatalf("unexpected message: got %q want %q", got, want)
	}
}

func TestValidateShellRejectsRetentionLimitAboveMax(t *testing.T) {
	v := newTestValidator(t)
	spec := &models.Stack{
		Name:     "demo",
		Settings: &models.StackSettings{ReleaseRetentionLimit: models.MaxReleaseRetentionLimit + 1},
	}

	err := v.ValidateShell(context.Background(), spec)
	fe := requireSingleFieldError(t, err)
	if fe.Code != errors.VErrStackSettingsInvalid {
		t.Fatalf("unexpected code: got %q want %q", fe.Code, errors.VErrStackSettingsInvalid)
	}
}

func TestValidateShellRejectsDeployTimeoutAboveMax(t *testing.T) {
	v := newTestValidator(t)
	spec := &models.Stack{
		Name:     "demo",
		Settings: &models.StackSettings{DeployTimeoutMinutes: models.MaxDeployTimeoutMinutes + 1},
	}

	err := v.ValidateShell(context.Background(), spec)
	fe := requireSingleFieldError(t, err)
	if fe.Code != errors.VErrStackSettingsInvalid {
		t.Fatalf("unexpected code: got %q want %q", fe.Code, errors.VErrStackSettingsInvalid)
	}
}

func TestValidateShellRejectsMinSuccessfulExceedingRetention(t *testing.T) {
	v := newTestValidator(t)
	spec := &models.Stack{
		Name: "demo",
		Settings: &models.StackSettings{
			ReleaseRetentionLimit: 5,
			MinSuccessfulReleases: 10,
		},
	}

	err := v.ValidateShell(context.Background(), spec)
	fe := requireSingleFieldError(t, err)
	if got, want := fe.Message, "min_successful_releases (10) must not exceed release_retention_limit (5)"; got != want {
		t.Fatalf("unexpected message: got %q want %q", got, want)
	}
}

func TestValidateShellAcceptsValidSettings(t *testing.T) {
	v := newTestValidator(t)
	spec := &models.Stack{
		Name: "demo",
		Settings: &models.StackSettings{
			ReleaseRetentionLimit: 20,
			MinSuccessfulReleases: 10,
			DeployTimeoutMinutes:  30,
		},
	}

	if err := v.ValidateShell(context.Background(), spec); err != nil {
		t.Fatalf("expected valid settings to pass, got %v", err)
	}
}

func TestValidateShellAcceptsNilSettings(t *testing.T) {
	v := newTestValidator(t)
	spec := &models.Stack{Name: "demo"}

	if err := v.ValidateShell(context.Background(), spec); err != nil {
		t.Fatalf("expected nil settings to pass, got %v", err)
	}
}

func TestValidateForCreateAcceptsValidSettings(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.Settings = &models.StackSettings{
		ReleaseRetentionLimit: 20,
		MinSuccessfulReleases: 10,
		DeployTimeoutMinutes:  30,
	}

	err := v.ValidateForCreate(context.Background(), spec)
	if err != nil {
		t.Fatalf("expected valid settings to pass, got %v", err)
	}
}

func TestValidateForCreateAcceptsNilSettings(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.Settings = nil

	err := v.ValidateForCreate(context.Background(), spec)
	if err != nil {
		t.Fatalf("expected nil settings to pass, got %v", err)
	}
}

// TestValidateForCreateAcceptsEmptyStackResources exercises the thin
// stack-shell create path: the handler zeroes out StackResources, Volumes,
// and Connections before calling CreateStack, so the fat validator must run
// clean over a stack with no children rather than panicking or erroring.
func TestValidateForCreateAcceptsEmptyStackResources(t *testing.T) {
	v := newTestValidator(t)
	spec := &models.Stack{
		Name:           "shell-stack",
		OrganisationID: "org-1",
		UserID:         "user-1",
	}

	err := v.ValidateForCreate(context.Background(), spec)
	if err != nil {
		t.Fatalf("expected empty stack resources to pass, got %v", err)
	}
}

// TestValidateForUpdateAcceptsEmptyStackResources mirrors the shell-update
// path (PUT /stacks/{id}): existing and desired specs both carry no
// resources/volumes/connections.
func TestValidateForUpdateAcceptsEmptyStackResources(t *testing.T) {
	v := newTestValidator(t)
	existing := &models.Stack{
		Name:           "shell-stack",
		OrganisationID: "org-1",
		UserID:         "user-1",
	}
	desired := &models.Stack{
		Name:           "shell-stack",
		OrganisationID: "org-1",
		UserID:         "user-1",
	}

	err := v.ValidateForUpdate(context.Background(), existing, desired)
	if err != nil {
		t.Fatalf("expected empty stack resources to pass, got %v", err)
	}
}

// TestValidateForCreateDedupesDuplicateNameErrors exercises the fat path's
// two independent duplicate-name detectors: stackValidator's own
// validateUniqueResourceNames, and the per-resource sibling rule that the
// (mocked) ResourceValidator would run in production. Both report the same
// offending resource with matching field/code/message text; the aggregated
// result must collapse that overlap to exactly one error per offending
// resource index rather than surfacing 2-3 copies.
func TestValidateForCreateDedupesDuplicateNameErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	resourceValidator := mocks.NewMockValidator(ctrl)
	resourceValidator.EXPECT().
		Validate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ *models.Stack, resource *models.StackResource, siblings []*models.StackResource) ([]errors.FieldError, *errors.ServiceError) {
			// Mirrors stackresource.validateSiblingRules' name-dup check.
			for _, s := range siblings {
				if s.Name == resource.Name {
					return []errors.FieldError{{
						Field:   "name",
						Code:    errors.VErrResourceNameDuplicate,
						Message: fmt.Sprintf("duplicate stack resource name '%s'", resource.Name),
					}}, nil
				}
			}
			return nil, nil
		}).
		AnyTimes()

	v := NewStackValidator(StackValidatorSpec{ResourceValidator: resourceValidator})

	spec := &models.Stack{
		Name:           "test-stack",
		OrganisationID: "org-1",
		UserID:         "user-1",
		StackResources: []*models.StackResource{
			{Name: "web", ImageConfig: &models.ImageConfigSpec{Image: "nginx:latest"}},
			{Name: "web", ImageConfig: &models.ImageConfigSpec{Image: "nginx:latest"}},
		},
	}

	err := v.ValidateForCreate(context.Background(), spec)
	errs := fieldErrors(t, err)

	dupErrs := map[string]int{}
	for _, fe := range errs {
		if fe.Code == errors.VErrResourceNameDuplicate {
			dupErrs[fe.Field]++
		}
	}
	if len(dupErrs) != 2 {
		t.Fatalf("expected duplicate-name errors on exactly 2 resource indices, got %v (all errors: %#v)", dupErrs, errs)
	}
	for field, count := range dupErrs {
		if count != 1 {
			t.Fatalf("expected exactly 1 duplicate-name error for field %q, got %d (all errors: %#v)", field, count, errs)
		}
	}
}

// newTestValidatorSpec returns a spec with a permissive ResourceValidator
// stub: every call to Validate succeeds with no field errors, so tests can
// focus on stack-level behavior (connections, settings) without needing to
// wire the full stackresource.Validator dependency graph.
func newTestValidatorSpec(t *testing.T) StackValidatorSpec {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	resourceValidator := mocks.NewMockValidator(ctrl)
	resourceValidator.EXPECT().
		Validate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil).
		AnyTimes()

	return StackValidatorSpec{
		ResourceValidator: resourceValidator,
	}
}

func newTestValidator(t *testing.T) validator.StackValidator {
	t.Helper()
	return NewStackValidator(newTestValidatorSpec(t))
}

func stackWithPorts(ports ...models.Port) *models.Stack {
	return &models.Stack{
		Name:           "test-stack",
		OrganisationID: "org-1",
		UserID:         "user-1",
		StackResources: []*models.StackResource{
			{
				Name:        "web",
				ImageConfig: &models.ImageConfigSpec{Image: "nginx:latest"},
				Ports:       ports,
			},
		},
	}
}

func stackWithConnections(connections ...models.StackConnection) *models.Stack {
	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.Connections = connections
	return spec
}

func newValidatorWithMockedPostgresAddonService(t *testing.T) (validator.StackValidator, *mocks.MockpostgresAddonService) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	postgresAddons := mocks.NewMockpostgresAddonService(ctrl)
	spec := newTestValidatorSpec(t)
	spec.PostgresAddonService = postgresAddons
	return NewStackValidator(spec), postgresAddons
}

func newValidatorWithMockedSecretService(t *testing.T) (validator.StackValidator, *mocks.MocksecretService) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	secrets := mocks.NewMocksecretService(ctrl)
	spec := newTestValidatorSpec(t)
	spec.SecretService = secrets
	return NewStackValidator(spec), secrets
}

func TestValidateForCreateAllowsBuildArtifactSourceConnection(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.StackResources = append(spec.StackResources, &models.StackResource{
		Name: "builder",
		BuildConfig: &models.BuildConfigSpec{
			SourceContext: models.BuildContextSource{
				Volume: &models.VolumeBuildSource{SourceVolumeName: "src"},
			},
			SourceRevision: models.BuildSourceRevision{
				Volume: &models.VolumeRevision{CurrentVolumeHash: "abc"},
			},
			BuildImageRepository: models.BuildImageRepository{UseInClusterRegistry: true},
		},
	})
	spec.Volumes = []*models.Volume{{Name: "src"}, {Name: "assets"}}
	spec.Connections = models.StackConnections{
		{
			ID:   "build-assets",
			Kind: models.ConnectionKindBuildArtifactSource,
			From: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "builder"},
			To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Name: "assets"},
			Config: map[string]interface{}{
				"source_path":      "/app/public",
				"destination_path": "/",
			},
		},
	}

	err := v.ValidateForCreate(context.Background(), spec)
	if err != nil {
		t.Fatalf("expected build_artifact_source connection to validate, got %v", err)
	}
}

func TestValidateForCreateRejectsBuildArtifactSourceWithoutSourcePath(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.StackResources[0].BuildConfig = &models.BuildConfigSpec{}
	spec.Volumes = []*models.Volume{{Name: "assets"}}
	spec.Connections = models.StackConnections{
		{
			ID:   "build-assets",
			Kind: models.ConnectionKindBuildArtifactSource,
			From: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
			To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Name: "assets"},
			Config: map[string]interface{}{
				"destination_path": "/",
			},
		},
	}

	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatal("expected build_artifact_source without source_path to be rejected")
	}
}

func TestValidateForCreateRejectsBuildArtifactSourceTargetingNonVolume(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.StackResources[0].BuildConfig = &models.BuildConfigSpec{}
	spec.Connections = models.StackConnections{
		{
			ID:   "build-assets",
			Kind: models.ConnectionKindBuildArtifactSource,
			From: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
			To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
			Config: map[string]interface{}{
				"source_path": "/app/public",
			},
		},
	}

	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatal("expected build_artifact_source targeting non-volume to be rejected")
	}
}

func TestValidateForCreateRejectsBuildArtifactSourceWithUnknownVolume(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.StackResources[0].BuildConfig = &models.BuildConfigSpec{}
	spec.Connections = models.StackConnections{
		{
			ID:   "build-assets",
			Kind: models.ConnectionKindBuildArtifactSource,
			From: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
			To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Name: "nonexistent"},
			Config: map[string]interface{}{
				"source_path": "/app/public",
			},
		},
	}

	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatal("expected build_artifact_source with unknown volume to be rejected")
	}
}

func TestValidateForCreateRejectsValueRefWithTemplateMissingValues(t *testing.T) {
	v, postgresAddons := newValidatorWithMockedPostgresAddonService(t)
	postgresAddons.EXPECT().GetPostgresAddon(gomock.Any(), "pg-1").Return(&models.PostgresAddon{
		ID:             "pg-1",
		OrganisationID: "org-1",
		Databases:      []models.PostgresAddonDatabase{{Name: "app"}},
	}, nil)

	spec := stackWithConnections(models.StackConnection{
		Kind: models.ConnectionKindEnv,
		From: models.TopologyNodeRef{Type: models.TopologyNodeTypePostgresAddon, Id: "pg-1"},
		To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
		Config: map[string]interface{}{
			"database": "app",
		},
		Mappings: []models.ConnectionMapping{
			{
				Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "DATABASE_URL"},
				Value:  models.ValueRef{Template: "postgres://{{ host }}:{{ port }}/{{ db }}"},
			},
		},
	})

	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatal("expected template with empty values to be rejected")
	}
}

func TestValidateForCreateRejectsValueRefWithBothOutputAndTemplate(t *testing.T) {
	v, postgresAddons := newValidatorWithMockedPostgresAddonService(t)
	postgresAddons.EXPECT().GetPostgresAddon(gomock.Any(), "pg-1").Return(&models.PostgresAddon{
		ID:             "pg-1",
		OrganisationID: "org-1",
		Databases:      []models.PostgresAddonDatabase{{Name: "app"}},
	}, nil)

	spec := stackWithConnections(models.StackConnection{
		Kind: models.ConnectionKindEnv,
		From: models.TopologyNodeRef{Type: models.TopologyNodeTypePostgresAddon, Id: "pg-1"},
		To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
		Config: map[string]interface{}{
			"database": "app",
		},
		Mappings: []models.ConnectionMapping{
			{
				Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "PG_HOST"},
				Value:  models.ValueRef{Output: "host", Template: "{{ host }}"},
			},
		},
	})

	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatal("expected value ref with both output and template to be rejected")
	}
}

func TestValidateForCreateRejectsValueRefWithNeitherOutputNorTemplate(t *testing.T) {
	v, postgresAddons := newValidatorWithMockedPostgresAddonService(t)
	postgresAddons.EXPECT().GetPostgresAddon(gomock.Any(), "pg-1").Return(&models.PostgresAddon{
		ID:             "pg-1",
		OrganisationID: "org-1",
		Databases:      []models.PostgresAddonDatabase{{Name: "app"}},
	}, nil)

	spec := stackWithConnections(models.StackConnection{
		Kind: models.ConnectionKindEnv,
		From: models.TopologyNodeRef{Type: models.TopologyNodeTypePostgresAddon, Id: "pg-1"},
		To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
		Config: map[string]interface{}{
			"database": "app",
		},
		Mappings: []models.ConnectionMapping{
			{
				Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "PG_HOST"},
				Value:  models.ValueRef{},
			},
		},
	})

	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatal("expected empty value ref to be rejected")
	}
}

func TestBuildConfigSpecValidateAcceptsCommitOnly(t *testing.T) {
	cfg := models.BuildConfigSpec{
		SourceContext: models.BuildContextSource{
			Volume: &models.VolumeBuildSource{SourceVolumeName: "src"},
		},
		SourceRevision: models.BuildSourceRevision{
			Git: &models.GitRevision{Commit: "abc1234"},
		},
		BuildImageRepository: models.BuildImageRepository{UseInClusterRegistry: true},
	}
	err := cfg.Validate()
	if err != nil {
		t.Fatalf("expected commit-only git revision to validate, got %v", err)
	}
}

func TestBuildConfigSpecValidateAcceptsBranchAndCommit(t *testing.T) {
	cfg := models.BuildConfigSpec{
		SourceContext: models.BuildContextSource{
			Volume: &models.VolumeBuildSource{SourceVolumeName: "src"},
		},
		SourceRevision: models.BuildSourceRevision{
			Git: &models.GitRevision{
				Branch: "main",
				Commit: "abc1234",
			},
		},
		BuildImageRepository: models.BuildImageRepository{UseInClusterRegistry: true},
	}
	err := cfg.Validate()
	if err != nil {
		t.Fatalf("expected branch+commit to validate, got %v", err)
	}
}

func containsSubstr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestBuildConfigSpecValidateRejectsBranchAndTag(t *testing.T) {
	cfg := models.BuildConfigSpec{
		SourceContext: models.BuildContextSource{
			Volume: &models.VolumeBuildSource{SourceVolumeName: "src"},
		},
		SourceRevision: models.BuildSourceRevision{
			Git: &models.GitRevision{
				Branch: "main",
				Tag:    "v1.0.0",
			},
		},
		BuildImageRepository: models.BuildImageRepository{UseInClusterRegistry: true},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatalf("expected branch+tag to be rejected")
	}
	if !containsSubstr(err.Error(), "branch and tag cannot both be set") {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

// --- delegation / aggregation tests ---

func TestValidateForCreateAggregatesFieldErrorsAcrossResources(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	resourceValidator := mocks.NewMockValidator(ctrl)
	resourceValidator.EXPECT().
		Validate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ *models.Stack, resource *models.StackResource, _ []*models.StackResource) ([]errors.FieldError, *errors.ServiceError) {
			return []errors.FieldError{
				{Field: "name", Code: errors.VErrResourceNameInvalid, Message: fmt.Sprintf("bad resource '%s'", resource.Name)},
			}, nil
		}).
		Times(2)

	spec := &models.Stack{
		Name:           "test-stack",
		OrganisationID: "org-1",
		UserID:         "user-1",
		StackResources: []*models.StackResource{
			{Name: "web", ImageConfig: &models.ImageConfigSpec{Image: "nginx:latest"}},
			{Name: "worker", ImageConfig: &models.ImageConfigSpec{Image: "nginx:latest"}},
		},
	}

	v := NewStackValidator(StackValidatorSpec{ResourceValidator: resourceValidator})

	err := v.ValidateForCreate(context.Background(), spec)
	errs := fieldErrors(t, err)
	if len(errs) != 2 {
		t.Fatalf("expected 2 aggregated field errors, got %d: %#v", len(errs), errs)
	}
	seen := map[string]bool{}
	for _, fe := range errs {
		seen[fe.Field] = true
	}
	if !seen["spec.stack_resources[0].name"] || !seen["spec.stack_resources[1].name"] {
		t.Fatalf("expected prefixed fields for both resources, got %#v", errs)
	}
}

func TestValidateForCreatePropagatesResourceValidatorServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	resourceValidator := mocks.NewMockValidator(ctrl)
	infraErr := errors.GeneralError("db unreachable")
	resourceValidator.EXPECT().
		Validate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, infraErr)

	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	v := NewStackValidator(StackValidatorSpec{ResourceValidator: resourceValidator})

	err := v.ValidateForCreate(context.Background(), spec)
	if err != infraErr {
		t.Fatalf("expected infra ServiceError to propagate unchanged, got %v", err)
	}
}

func TestValidateForUpdateDelegatesToResourceValidatorWithPrefix(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	resourceValidator := mocks.NewMockValidator(ctrl)
	resourceValidator.EXPECT().
		Validate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]errors.FieldError{
			{Field: "name", Code: errors.VErrResourceNameInvalid, Message: "bad"},
		}, nil)

	existing := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	desired := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	desired.Name = existing.Name
	desired.UserID = existing.UserID
	desired.OrganisationID = existing.OrganisationID

	v := NewStackValidator(StackValidatorSpec{ResourceValidator: resourceValidator})

	err := v.ValidateForUpdate(context.Background(), existing, desired)
	fe := requireSingleFieldError(t, err)
	if got, want := fe.Field, "spec.stack_resources[0].name"; got != want {
		t.Fatalf("unexpected field: got %q want %q", got, want)
	}
}

// TestValidateForCreateReportsMissingMountedVolumeFromResourceValidator
// covers the fat-path behavior the (now removed) stackValidator-local
// validateVolumeReferences used to provide on its own: a mount referencing
// a volume the request doesn't declare must still surface as a 400 with
// VErrVolumeNotFound, prefixed to the offending resource's index. That
// detection now happens solely inside the delegated
// stackresource.Validator - production's real Validate checks the
// request's own bundled volumes first, then falls back to a
// namespace-scoped DB lookup that would 404 for a name the payload never
// declared. Here the mock stands in for that whole (payload-first + DB
// fallback) rule and returns the not-found error directly; what this test
// verifies is that stackValidator.validateResources still prefixes and
// surfaces it correctly now that it's the only source of this error.
func TestValidateForCreateReportsMissingMountedVolumeFromResourceValidator(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	resourceValidator := mocks.NewMockValidator(ctrl)
	resourceValidator.EXPECT().
		Validate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]errors.FieldError{
			{
				Field:   "volume_mounts[0].source_volume",
				Code:    errors.VErrVolumeNotFound,
				Message: "volume 'missing-volume' does not exist",
			},
		}, nil)

	spec := &models.Stack{
		Name:           "test-stack",
		OrganisationID: "org-1",
		UserID:         "user-1",
		StackResources: []*models.StackResource{
			{
				Name:        "web",
				ImageConfig: &models.ImageConfigSpec{Image: "nginx:latest"},
				VolumeMounts: []*models.VolumeMount{
					{SourceVolumeName: "missing-volume", TargetPath: "/data"},
				},
			},
		},
	}

	v := NewStackValidator(StackValidatorSpec{ResourceValidator: resourceValidator})

	err := v.ValidateForCreate(context.Background(), spec)
	fe := requireSingleFieldError(t, err)
	if got, want := fe.Field, "spec.stack_resources[0].volume_mounts[0].source_volume"; got != want {
		t.Fatalf("unexpected field: got %q want %q", got, want)
	}
	if got, want := fe.Code, errors.VErrVolumeNotFound; got != want {
		t.Fatalf("unexpected code: got %q want %q", got, want)
	}
}

func TestValidateConnectionsRejectsUnknownTargetResource(t *testing.T) {
	// No expectations set on the resource validator mock: ValidateConnections
	// must not invoke per-resource validation at all.
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	resourceValidator := mocks.NewMockValidator(ctrl)
	v := NewStackValidator(StackValidatorSpec{ResourceValidator: resourceValidator})

	spec := stackWithConnections(models.StackConnection{
		ID:   "internal-api",
		Kind: models.ConnectionKindEnv,
		From: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
		To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "phantom"},
		Mappings: []models.ConnectionMapping{
			{
				Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "WEB_URL"},
				Value:  models.ValueRef{Output: models.OutputNameURL},
			},
		},
	})

	err := v.ValidateConnections(context.Background(), spec)
	fe := requireSingleFieldError(t, err)
	if got, want := fe.Field, "spec.connections[0]"; got != want {
		t.Fatalf("unexpected field: got %q want %q", got, want)
	}
	if got, want := fe.Code, errors.VErrConnectionInvalid; got != want {
		t.Fatalf("unexpected code: got %q want %q", got, want)
	}
	if got, want := fe.Message, "connection 'internal-api' references unknown stack resource 'phantom'"; got != want {
		t.Fatalf("unexpected message: got %q want %q", got, want)
	}
}

func TestValidateConnectionsAcceptsValidConnection(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	resourceValidator := mocks.NewMockValidator(ctrl)
	v := NewStackValidator(StackValidatorSpec{ResourceValidator: resourceValidator})

	spec := stackWithConnections(models.StackConnection{
		ID:   "internal-api",
		Kind: models.ConnectionKindEnv,
		From: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
		To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
		Mappings: []models.ConnectionMapping{
			{
				Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "SELF_URL"},
				Value:  models.ValueRef{Output: models.OutputNameURL},
			},
		},
	})

	if err := v.ValidateConnections(context.Background(), spec); err != nil {
		t.Fatalf("expected valid connection to pass, got %v", err)
	}
}

// TestValidateConnectionsIgnoresUnrelatedResourceInvalidity is the point of
// the narrow gate: a connection-only mutation must not be blocked by a
// pre-existing, unrelated invalidity elsewhere in the stack (e.g. a bad port
// on a resource the connection doesn't touch) that the connection form gives
// the user no way to fix. ValidateForUpdate's full per-resource pass would
// surface it; ValidateConnections must not even invoke the per-resource
// validator.
func TestValidateConnectionsIgnoresUnrelatedResourceInvalidity(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	resourceValidator := mocks.NewMockValidator(ctrl)
	resourceValidator.EXPECT().
		Validate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ *models.Stack, resource *models.StackResource, _ []*models.StackResource) ([]errors.FieldError, *errors.ServiceError) {
			if resource.Name == "worker" {
				return []errors.FieldError{
					{
						Field:   "ports[0].number",
						Code:    errors.VErrPortNumberInvalid,
						Message: "port number is invalid",
					},
				}, nil
			}
			return nil, nil
		}).
		AnyTimes()

	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.StackResources = append(spec.StackResources, &models.StackResource{
		Name:  "worker",
		Ports: []models.Port{{Name: "bad", Number: -1, Protocol: "http"}},
	})
	spec.Connections = models.StackConnections{
		{
			ID:   "internal-api",
			Kind: models.ConnectionKindEnv,
			From: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
			To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "worker"},
			Mappings: []models.ConnectionMapping{
				{
					Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "WEB_URL"},
					Value:  models.ValueRef{Output: models.OutputNameURL},
				},
			},
		},
	}

	v := NewStackValidator(StackValidatorSpec{ResourceValidator: resourceValidator})

	// Sanity check: the full-stack path does surface the unrelated resource's
	// invalid port.
	if err := v.ValidateForUpdate(context.Background(), spec, spec); err == nil {
		t.Fatal("expected ValidateForUpdate to surface the unrelated resource's invalid port")
	}

	// The connection-scoped gate must ignore it.
	if err := v.ValidateConnections(context.Background(), spec); err != nil {
		t.Fatalf("expected ValidateConnections to ignore unrelated resource invalidity, got %v", err)
	}
}

func shellStack(name string) *models.Stack {
	return &models.Stack{
		Name:           name,
		OrganisationID: "org-1",
		UserID:         "user-1",
	}
}

// requireStackNameInvalid asserts err carries exactly one field error
// addressed to "name" with code VErrStackNameInvalid, and returns it.
func requireStackNameInvalid(t *testing.T, err *errors.ServiceError) errors.FieldError {
	t.Helper()
	fe := requireSingleFieldError(t, err)
	if fe.Field != "name" {
		t.Fatalf("unexpected field: got %q want %q", fe.Field, "name")
	}
	if fe.Code != errors.VErrStackNameInvalid {
		t.Fatalf("unexpected code: got %q want %q", fe.Code, errors.VErrStackNameInvalid)
	}
	return fe
}

func TestValidateForCreateAcceptsStackNameAtNamespaceBudget(t *testing.T) {
	v := newTestValidator(t)
	spec := shellStack(strings.Repeat("a", models.MaxStackNameLength))

	if err := v.ValidateForCreate(context.Background(), spec); err != nil {
		t.Fatalf("expected %d-character stack name to pass, got %v", models.MaxStackNameLength, err)
	}
}

func TestValidateForCreateRejectsStackNameOverNamespaceBudget(t *testing.T) {
	v := newTestValidator(t)
	spec := shellStack(strings.Repeat("a", models.MaxStackNameLength+1))

	err := v.ValidateForCreate(context.Background(), spec)
	requireStackNameInvalid(t, err)
}

func TestValidateForCreateRejectsEmptyStackName(t *testing.T) {
	v := newTestValidator(t)
	spec := shellStack("")

	err := v.ValidateForCreate(context.Background(), spec)
	fe := requireStackNameInvalid(t, err)
	if got, want := fe.Message, "stack name is required"; got != want {
		t.Fatalf("unexpected message: got %q want %q", got, want)
	}
}

func TestValidateForCreateRejectsNonDNSLabelStackNames(t *testing.T) {
	cases := map[string]string{
		"uppercase":       "MyStack",
		"underscore":      "my_stack",
		"leading hyphen":  "-stack",
		"trailing hyphen": "stack-",
		"dot":             "my.stack",
		"space":           "my stack",
	}
	for label, name := range cases {
		t.Run(label, func(t *testing.T) {
			v := newTestValidator(t)
			err := v.ValidateForCreate(context.Background(), shellStack(name))
			requireStackNameInvalid(t, err)
		})
	}
}
