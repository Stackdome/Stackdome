package stack

import (
	"context"
	"fmt"
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
		ID: "pg-1",
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
		ID: "pg-1",
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
	postgresAddons.EXPECT().GetPostgresAddon(gomock.Any(), "pg-1").Return(&models.PostgresAddon{ID: "pg-1"}, nil)

	err := v.ValidateForCreate(context.Background(), spec)
	fe := requireSingleFieldError(t, err)
	if got, want := fe.Message, "connection 'pg-env' requires config.database when postgres credential scope is owner"; got != want {
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
					Output: "url.http",
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
		ID:   "sec-1",
		Keys: []string{"tls.crt"},
	}, nil)

	err := v.ValidateForCreate(context.Background(), spec)
	if err != nil {
		t.Fatalf("expected secret connection output to validate, got %v", err)
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
		ID:        "pg-1",
		Databases: []models.PostgresAddonDatabase{{Name: "app"}},
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
		ID:        "pg-1",
		Databases: []models.PostgresAddonDatabase{{Name: "app"}},
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
		ID:        "pg-1",
		Databases: []models.PostgresAddonDatabase{{Name: "app"}},
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
