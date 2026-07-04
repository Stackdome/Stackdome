package stack

import (
	"context"
	"fmt"
	"testing"

	"github.com/Stackdome/stackdome/pkg/clients"
	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/validator"
	"go.uber.org/mock/gomock"
)

func TestValidateForCreateRequiresNamedPorts(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithPorts(models.Port{
		Number:          8080,
		Protocol:        "http",
		ExposedToPublic: false,
	})

	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatalf("expected unnamed port to be rejected")
	}
	if got, want := err.Error(), "error: stack resource 'web' has port 8080 missing name"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func TestValidateForCreateRejectsDuplicatePortNames(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithPorts(
		models.Port{Name: "http", Number: 8080, Protocol: "http"},
		models.Port{Name: "http", Number: 9090, Protocol: "http"},
	)

	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatalf("expected duplicate port names to be rejected")
	}
	if got, want := err.Error(), "error: stack resource 'web' has duplicate port name 'http'"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func TestValidateForCreateRejectsDuplicatePortNumbers(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithPorts(
		models.Port{Name: "http", Number: 8080, Protocol: "http"},
		models.Port{Name: "metrics", Number: 8080, Protocol: "http"},
	)

	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatalf("expected duplicate port numbers to be rejected")
	}
	if got, want := err.Error(), "error: stack resource 'web' has duplicate port number 8080"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func TestValidateForCreateAllowsSelfOutputEnvVar(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.StackResources[0].ExecutionConfig = &models.ExecutionConfig{
		Env: []models.EnvVar{
			{Name: "INTERNAL_URL", SelfOutput: "url.http"},
		},
	}

	err := v.ValidateForCreate(context.Background(), spec)
	if err != nil {
		t.Fatalf("expected self_output env var to validate, got %v", err)
	}
}

func TestValidateForCreateRejectsEnvVarWithBothValueAndSelfOutput(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.StackResources[0].ExecutionConfig = &models.ExecutionConfig{
		Env: []models.EnvVar{
			{Name: "PUBLIC_URL", Value: "https://example.com", SelfOutput: "url.http"},
		},
	}

	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatalf("expected env var with both value and self_output to be rejected")
	}
	if got, want := err.Error(), "error: stack resource 'web' env var 'PUBLIC_URL' must set exactly one of value or self_output"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
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
	if err == nil {
		t.Fatalf("expected secret_mount connection to be rejected")
	}
	if got, want := err.Error(), "error: connection 'secret-files' has unsupported kind 'secret_mount'"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
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
	if err == nil {
		t.Fatalf("expected postgres owner connection without database to be rejected")
	}
	if got, want := err.Error(), "error: connection 'pg-env' requires config.database when postgres credential scope is owner"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
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
	if err == nil {
		t.Fatalf("expected unknown postgres config key to be rejected")
	}
	if got, want := err.Error(), "error: connection 'pg-env' has unsupported postgres config key 'oops'"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
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
	if err == nil {
		t.Fatalf("expected volume mount without mount_path to be rejected")
	}
	if got, want := err.Error(), "error: connection 'volume-mount' requires config.mount_path for volume mounts"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
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
	if err == nil {
		t.Fatalf("expected non-bool read_only to be rejected")
	}
	if got, want := err.Error(), "error: connection 'volume-mount' config.read_only must be a boolean"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
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
	if err == nil {
		t.Fatalf("expected unknown stack resource output to be rejected")
	}
	if got, want := err.Error(), "error: connection 'internal-api' references unsupported output 'url.grpc' for source 'stack_resource:web'"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
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
	if err == nil {
		t.Fatal("expected retention limit above max to be rejected")
	}
	if got, want := err.Error(), "error: release_retention_limit must be at most 50"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func TestValidateForCreateRejectsMinSuccessfulAboveMax(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.Settings = &models.StackSettings{MinSuccessfulReleases: models.MaxMinSuccessfulReleases + 1}

	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatal("expected min_successful_releases above max to be rejected")
	}
	if got, want := err.Error(), "error: min_successful_releases must be at most 20"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func TestValidateForCreateRejectsDeployTimeoutAboveMax(t *testing.T) {
	v := newTestValidator(t)
	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.Settings = &models.StackSettings{DeployTimeoutMinutes: models.MaxDeployTimeoutMinutes + 1}

	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatal("expected deploy_timeout_minutes above max to be rejected")
	}
	if got, want := err.Error(), "error: deploy_timeout_minutes must be at most 120"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
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
	if err == nil {
		t.Fatal("expected min_successful_releases > release_retention_limit to be rejected")
	}
	if got, want := err.Error(), "error: min_successful_releases (10) must not exceed release_retention_limit (5)"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
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

// newTestValidatorSpec returns a spec with permissive probe dependencies: the
// resolver resolves everything anonymously and the registry client reports
// every image as pullable and every push as allowed.
func newTestValidatorSpec(t *testing.T) StackValidatorSpec {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	resolver := mocks.NewMockCredentialResolver(ctrl)
	resolver.EXPECT().
		RegistryCredentials(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&credentials.ResolvedRegistryCredential{Source: credentials.SourceAnonymous}, nil).
		AnyTimes()
	resolver.EXPECT().
		GitCredentials(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&credentials.ResolvedGitCredential{Source: credentials.SourceAnonymous}, nil).
		AnyTimes()

	registryClient := mocks.NewMockRegistryClient(ctrl)
	registryClient.EXPECT().CheckImage(gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
	registryClient.EXPECT().CheckPushAccess(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	provider := mocks.NewMockregistryClientProvider(ctrl)
	provider.EXPECT().ClientFor(gomock.Any()).Return(registryClient, nil).AnyTimes()

	gitClient := mocks.NewMockGitClient(ctrl)
	gitClient.EXPECT().CheckAccess(gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()

	gitClients := mocks.NewMockgitClientProvider(ctrl)
	gitClients.EXPECT().ClientFor(gomock.Any(), gomock.Any()).Return(gitClient, nil).AnyTimes()

	return StackValidatorSpec{
		CredentialResolver: resolver,
		RegistryClients:    provider,
		GitClients:         gitClients,
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

func TestValidateForCreateAcceptsValidWorkloadTypes(t *testing.T) {
	v := newTestValidator(t)
	for _, wt := range []models.WorkloadType{
		models.WorkloadTypeService,
		models.WorkloadTypeStatefulService,
		models.WorkloadTypeWorker,
		models.WorkloadTypeJob,
		models.WorkloadTypeCronJob,
	} {
		t.Run(string(wt), func(t *testing.T) {
			sr := &models.StackResource{
				Name:         "app",
				WorkloadType: wt,
				ImageConfig:  &models.ImageConfigSpec{Image: "nginx:latest"},
			}
			if wt == models.WorkloadTypeCronJob {
				sr.Schedule = "*/5 * * * *"
			}
			if wt == models.WorkloadTypeService || wt == models.WorkloadTypeStatefulService {
				sr.Ports = []models.Port{{Name: "http", Number: 8080, Protocol: "http"}}
			}
			spec := &models.Stack{
				Name:           "test-stack",
				OrganisationID: "org-1",
				UserID:         "user-1",
				StackResources: []*models.StackResource{sr},
			}
			err := v.ValidateForCreate(context.Background(), spec)
			if err != nil {
				t.Fatalf("expected workload type %s to be accepted, got %v", wt, err)
			}
		})
	}
}

func TestValidateForCreateRejectsInvalidWorkloadType(t *testing.T) {
	v := newTestValidator(t)
	spec := &models.Stack{
		Name:           "test-stack",
		OrganisationID: "org-1",
		UserID:         "user-1",
		StackResources: []*models.StackResource{
			{
				Name:         "app",
				WorkloadType: models.WorkloadType("InvalidType"),
				ImageConfig:  &models.ImageConfigSpec{Image: "nginx:latest"},
			},
		},
	}
	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatal("expected invalid workload type to be rejected")
	}
	if got, want := err.Error(), "error: stack resource 'app' has unsupported workload_type 'InvalidType'"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func TestValidateForCreateCronJobRequiresSchedule(t *testing.T) {
	v := newTestValidator(t)
	spec := &models.Stack{
		Name:           "test-stack",
		OrganisationID: "org-1",
		UserID:         "user-1",
		StackResources: []*models.StackResource{
			{
				Name:         "cron",
				WorkloadType: models.WorkloadTypeCronJob,
				ImageConfig:  &models.ImageConfigSpec{Image: "nginx:latest"},
			},
		},
	}
	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatal("expected CronJob without schedule to be rejected")
	}
	if got, want := err.Error(), "error: stack resource 'cron' requires schedule for workload_type 'CronJob'"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func TestValidateForCreateRejectsScheduleForNonCronJob(t *testing.T) {
	v := newTestValidator(t)
	spec := &models.Stack{
		Name:           "test-stack",
		OrganisationID: "org-1",
		UserID:         "user-1",
		StackResources: []*models.StackResource{
			{
				Name:         "web",
				WorkloadType: models.WorkloadTypeService,
				Schedule:     "* * * * *",
				ImageConfig:  &models.ImageConfigSpec{Image: "nginx:latest"},
				Ports:        []models.Port{{Name: "http", Number: 8080, Protocol: "http"}},
			},
		},
	}
	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatal("expected schedule on non-CronJob to be rejected")
	}
	if got, want := err.Error(), "error: stack resource 'web' cannot set schedule for workload_type 'Service'"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func TestValidateForCreateRejectsPortsOnWorkerJobCronJob(t *testing.T) {
	v := newTestValidator(t)
	for _, wt := range []models.WorkloadType{
		models.WorkloadTypeWorker,
		models.WorkloadTypeJob,
		models.WorkloadTypeCronJob,
	} {
		t.Run(string(wt), func(t *testing.T) {
			sr := &models.StackResource{
				Name:         "app",
				WorkloadType: wt,
				ImageConfig:  &models.ImageConfigSpec{Image: "nginx:latest"},
				Ports:        []models.Port{{Name: "http", Number: 8080, Protocol: "http"}},
			}
			if wt == models.WorkloadTypeCronJob {
				sr.Schedule = "*/5 * * * *"
			}
			spec := &models.Stack{
				Name:           "test-stack",
				OrganisationID: "org-1",
				UserID:         "user-1",
				StackResources: []*models.StackResource{sr},
			}
			err := v.ValidateForCreate(context.Background(), spec)
			if err == nil {
				t.Fatalf("expected ports on %s to be rejected", wt)
			}
		})
	}
}

func TestValidateForCreateRejectsReplicasOnJobCronJob(t *testing.T) {
	v := newTestValidator(t)
	replicas := int32(3)
	for _, wt := range []models.WorkloadType{
		models.WorkloadTypeJob,
		models.WorkloadTypeCronJob,
	} {
		t.Run(string(wt), func(t *testing.T) {
			sr := &models.StackResource{
				Name:         "app",
				WorkloadType: wt,
				Replicas:     &replicas,
				ImageConfig:  &models.ImageConfigSpec{Image: "nginx:latest"},
			}
			if wt == models.WorkloadTypeCronJob {
				sr.Schedule = "*/5 * * * *"
			}
			spec := &models.Stack{
				Name:           "test-stack",
				OrganisationID: "org-1",
				UserID:         "user-1",
				StackResources: []*models.StackResource{sr},
			}
			err := v.ValidateForCreate(context.Background(), spec)
			if err == nil {
				t.Fatalf("expected replicas on %s to be rejected", wt)
			}
		})
	}
}

func TestBuildConfigSpecValidateRejectsCommitOnly(t *testing.T) {
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
	if err == nil {
		t.Fatal("expected commit-only git revision to be rejected")
	}
	if got, want := err.Error(), "a branch or tag is required"; got != want {
		// The error message may include a prefix; check containment
		if got != want && !containsSubstr(got, want) {
			t.Fatalf("unexpected error: got %q want it to contain %q", got, want)
		}
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

func TestValidateForCreatePushSecretHappyPath(t *testing.T) {
	v, secrets := newValidatorWithMockedSecretService(t)
	secrets.EXPECT().ValidateImageRegistrySecretForStackResource(gomock.Any(), "push-secret-1").Return(nil)

	spec := &models.Stack{
		Name:           "test",
		OrganisationID: "org-1",
		UserID:         "user-1",
		StackResources: []*models.StackResource{
			{
				Name: "api",
				BuildConfig: &models.BuildConfigSpec{
					SourceContext:        models.BuildContextSource{Git: &models.GitBuildSource{RepoURL: "https://github.com/example/repo"}},
					SourceRevision:       models.BuildSourceRevision{Git: &models.GitRevision{Branch: "main"}},
					BuildImageRepository: models.BuildImageRepository{ExternalImageRef: "myregistry.io/org/repo"},
					RegistrySecretRef:    &models.SecretReference{SecretID: "push-secret-1"},
				},
			},
		},
	}
	if err := v.ValidateForCreate(context.Background(), spec); err != nil {
		t.Fatalf("expected push secret validation to pass, got %v", err)
	}
}

func TestValidateForCreatePushSecretWrongType(t *testing.T) {
	v, secrets := newValidatorWithMockedSecretService(t)
	secrets.EXPECT().ValidateImageRegistrySecretForStackResource(gomock.Any(), "bad-secret").
		Return(errors.BadRequest("secret type mismatch"))

	spec := &models.Stack{
		Name:           "test",
		OrganisationID: "org-1",
		UserID:         "user-1",
		StackResources: []*models.StackResource{
			{
				Name: "api",
				BuildConfig: &models.BuildConfigSpec{
					SourceContext:        models.BuildContextSource{Git: &models.GitBuildSource{RepoURL: "https://github.com/example/repo"}},
					SourceRevision:       models.BuildSourceRevision{Git: &models.GitRevision{Branch: "main"}},
					BuildImageRepository: models.BuildImageRepository{ExternalImageRef: "myregistry.io/org/repo"},
					RegistrySecretRef:    &models.SecretReference{SecretID: "bad-secret"},
				},
			},
		},
	}
	if err := v.ValidateForCreate(context.Background(), spec); err == nil {
		t.Fatalf("expected push secret with wrong type to be rejected")
	}
}

func TestValidateForCreatePushSecretEmptyID(t *testing.T) {
	v := newTestValidator(t)
	spec := &models.Stack{
		Name:           "test",
		OrganisationID: "org-1",
		UserID:         "user-1",
		StackResources: []*models.StackResource{
			{
				Name: "api",
				BuildConfig: &models.BuildConfigSpec{
					SourceContext:        models.BuildContextSource{Git: &models.GitBuildSource{RepoURL: "https://github.com/example/repo"}},
					SourceRevision:       models.BuildSourceRevision{Git: &models.GitRevision{Branch: "main"}},
					BuildImageRepository: models.BuildImageRepository{ExternalImageRef: "myregistry.io/org/repo"},
					RegistrySecretRef:    &models.SecretReference{SecretID: ""},
				},
			},
		},
	}
	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatalf("expected empty push secret ID to be rejected")
	}
	if !containsSubstr(err.Error(), "empty push secret ID") {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestValidateForCreateRejectsNegativeReplicas(t *testing.T) {
	v := newTestValidator(t)
	neg := int32(-1)
	spec := &models.Stack{
		Name:           "test",
		OrganisationID: "org-1",
		UserID:         "user-1",
		StackResources: []*models.StackResource{
			{
				Name:         "api",
				WorkloadType: models.WorkloadTypeService,
				Replicas:     &neg,
				ImageConfig:  &models.ImageConfigSpec{Image: "nginx:latest"},
			},
		},
	}
	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatalf("expected negative replicas to be rejected")
	}
	if !containsSubstr(err.Error(), "cannot be negative") {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestValidateForCreateAcceptsCronDescriptorSchedule(t *testing.T) {
	v := newTestValidator(t)
	spec := &models.Stack{
		Name:           "test",
		OrganisationID: "org-1",
		UserID:         "user-1",
		StackResources: []*models.StackResource{
			{
				Name:         "cleanup",
				WorkloadType: models.WorkloadTypeCronJob,
				Schedule:     "@daily",
				ImageConfig:  &models.ImageConfigSpec{Image: "busybox:latest"},
			},
		},
	}
	if err := v.ValidateForCreate(context.Background(), spec); err != nil {
		t.Fatalf("expected @daily schedule to be accepted, got %v", err)
	}
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

func newValidatorWithMockedSecretService(t *testing.T) (validator.StackValidator, *mocks.MocksecretService) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	secrets := mocks.NewMocksecretService(ctrl)
	spec := newTestValidatorSpec(t)
	spec.SecretService = secrets
	return NewStackValidator(spec), secrets
}

// --- credential resolver / registry probe tests ---

func newProbeMocks(t *testing.T) (*mocks.MockCredentialResolver, *mocks.MockregistryClientProvider, *mocks.MockRegistryClient, *mocks.MocksecretService) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	return mocks.NewMockCredentialResolver(ctrl),
		mocks.NewMockregistryClientProvider(ctrl),
		mocks.NewMockRegistryClient(ctrl),
		mocks.NewMocksecretService(ctrl)
}

// permissiveGitProbes stubs anonymous git resolution and a passing clone probe
// for tests focused on registry behavior.
func permissiveGitProbes(t *testing.T, resolver *mocks.MockCredentialResolver) gitClientProvider {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	resolver.EXPECT().
		GitCredentials(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&credentials.ResolvedGitCredential{Source: credentials.SourceAnonymous}, nil).
		AnyTimes()

	gitClient := mocks.NewMockGitClient(ctrl)
	gitClient.EXPECT().CheckAccess(gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
	gitClients := mocks.NewMockgitClientProvider(ctrl)
	gitClients.EXPECT().ClientFor(gomock.Any(), gomock.Any()).Return(gitClient, nil).AnyTimes()
	return gitClients
}

func stackWithBuildResource(repo models.BuildImageRepository, pushSecretRef *models.SecretReference) *models.Stack {
	return &models.Stack{
		Name:           "test-stack",
		OrganisationID: "org-1",
		UserID:         "user-1",
		StackResources: []*models.StackResource{
			{
				Name: "builder",
				BuildConfig: &models.BuildConfigSpec{
					SourceContext: models.BuildContextSource{
						Git: &models.GitBuildSource{RepoURL: "https://github.com/acme/api"},
					},
					SourceRevision: models.BuildSourceRevision{
						Git: &models.GitRevision{Branch: "main"},
					},
					BuildImageRepository: repo,
					RegistrySecretRef:    pushSecretRef,
				},
			},
		},
	}
}

func TestValidateForCreateVerifiesPrivateImageWithCredentials(t *testing.T) {
	resolver, provider, registryClient, secrets := newProbeMocks(t)

	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.StackResources[0].ImageConfig.Image = "ghcr.io/acme/private:latest"
	spec.StackResources[0].ImageConfig.PullSecretRef = &models.SecretReference{SecretID: "sec-1"}

	resolved := &credentials.ResolvedRegistryCredential{
		Source:   credentials.SourceSecretRef,
		Username: "alice",
		Password: "s3cret",
		SecretID: "sec-1",
	}
	secrets.EXPECT().ValidateImageRegistrySecretForStackResource(gomock.Any(), "sec-1").Return(nil)
	resolver.EXPECT().
		RegistryCredentials(gomock.Any(), "org-1", "ghcr.io/acme/private:latest", credentials.RegistryPurposePull, credentials.RegistryAuthSelector{SecretRef: spec.StackResources[0].ImageConfig.PullSecretRef}).
		Return(resolved, nil)
	provider.EXPECT().ClientFor(resolved).Return(registryClient, nil)
	registryClient.EXPECT().CheckImage(gomock.Any(), "ghcr.io/acme/private:latest").Return(true, nil)

	v := NewStackValidator(StackValidatorSpec{
		SecretService:      secrets,
		CredentialResolver: resolver,
		RegistryClients:    provider,
	})

	if err := v.ValidateForCreate(context.Background(), spec); err != nil {
		t.Fatalf("expected private image with credentials to validate, got %v", err)
	}
}

func TestValidateForCreateSoftPassesAnonymousRateLimit(t *testing.T) {
	resolver, provider, registryClient, _ := newProbeMocks(t)

	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})

	resolved := &credentials.ResolvedRegistryCredential{Source: credentials.SourceAnonymous}
	resolver.EXPECT().
		RegistryCredentials(gomock.Any(), "org-1", "nginx:latest", credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).
		Return(resolved, nil)
	provider.EXPECT().ClientFor(resolved).Return(registryClient, nil)
	registryClient.EXPECT().CheckImage(gomock.Any(), "nginx:latest").
		Return(false, fmt.Errorf("failed to check image: %w", clients.ErrRateLimited))

	v := NewStackValidator(StackValidatorSpec{
		CredentialResolver: resolver,
		RegistryClients:    provider,
	})

	if err := v.ValidateForCreate(context.Background(), spec); err != nil {
		t.Fatalf("expected anonymous rate-limited image check to soft-pass, got %v", err)
	}
}

func TestValidateForCreateFailsAuthenticatedRateLimit(t *testing.T) {
	resolver, provider, registryClient, secrets := newProbeMocks(t)

	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.StackResources[0].ImageConfig.PullSecretRef = &models.SecretReference{SecretID: "sec-1"}

	resolved := &credentials.ResolvedRegistryCredential{
		Source:   credentials.SourceSecretRef,
		Username: "alice",
		Password: "s3cret",
	}
	secrets.EXPECT().ValidateImageRegistrySecretForStackResource(gomock.Any(), "sec-1").Return(nil)
	resolver.EXPECT().
		RegistryCredentials(gomock.Any(), "org-1", "nginx:latest", credentials.RegistryPurposePull, credentials.RegistryAuthSelector{SecretRef: spec.StackResources[0].ImageConfig.PullSecretRef}).
		Return(resolved, nil)
	provider.EXPECT().ClientFor(resolved).Return(registryClient, nil)
	registryClient.EXPECT().CheckImage(gomock.Any(), "nginx:latest").
		Return(false, fmt.Errorf("failed to check image: %w", clients.ErrRateLimited))

	v := NewStackValidator(StackValidatorSpec{
		SecretService:      secrets,
		CredentialResolver: resolver,
		RegistryClients:    provider,
	})

	if err := v.ValidateForCreate(context.Background(), spec); err == nil {
		t.Fatal("expected rate-limited authenticated image check to fail validation")
	}
}

func TestValidateForCreateRejectsNonExistentImage(t *testing.T) {
	resolver, provider, registryClient, _ := newProbeMocks(t)

	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})

	resolved := &credentials.ResolvedRegistryCredential{Source: credentials.SourceAnonymous}
	resolver.EXPECT().
		RegistryCredentials(gomock.Any(), "org-1", "nginx:latest", credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).
		Return(resolved, nil)
	provider.EXPECT().ClientFor(resolved).Return(registryClient, nil)
	registryClient.EXPECT().CheckImage(gomock.Any(), "nginx:latest").Return(false, nil)

	v := NewStackValidator(StackValidatorSpec{
		CredentialResolver: resolver,
		RegistryClients:    provider,
	})

	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatal("expected non-existent image to be rejected")
	}
	if got, want := err.Error(), "error: stack resource 'web' image 'nginx:latest' does not exist or is not pullable"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func TestValidateForCreatePushProbeFailureNamesHost(t *testing.T) {
	resolver, provider, registryClient, secrets := newProbeMocks(t)

	pushRef := &models.SecretReference{SecretID: "push-sec"}
	spec := stackWithBuildResource(models.BuildImageRepository{ExternalImageRef: "ghcr.io/acme/app"}, pushRef)

	resolved := &credentials.ResolvedRegistryCredential{
		Source:   credentials.SourceSecretRef,
		Username: "alice",
		Password: "s3cret",
	}
	secrets.EXPECT().ValidateImageRegistrySecretForStackResource(gomock.Any(), "push-sec").Return(nil)
	resolver.EXPECT().
		RegistryCredentials(gomock.Any(), "org-1", "ghcr.io/acme/app", credentials.RegistryPurposePush, credentials.RegistryAuthSelector{SecretRef: pushRef}).
		Return(resolved, nil)
	provider.EXPECT().ClientFor(resolved).Return(registryClient, nil)
	registryClient.EXPECT().CheckPushAccess(gomock.Any(), "ghcr.io/acme/app").
		Return(fmt.Errorf("push access check failed: denied"))

	v := NewStackValidator(StackValidatorSpec{
		SecretService:      secrets,
		CredentialResolver: resolver,
		RegistryClients:    provider,
		GitClients:         permissiveGitProbes(t, resolver),
	})

	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatal("expected failing push probe to be rejected")
	}
	if !containsSubstr(err.Error(), "ghcr.io") {
		t.Fatalf("expected error to name the registry host, got %q", err.Error())
	}
}

func TestValidateForCreateSkipsPushProbeForInClusterRegistry(t *testing.T) {
	resolver, provider, _, secrets := newProbeMocks(t)

	pushRef := &models.SecretReference{SecretID: "push-sec"}
	spec := stackWithBuildResource(models.BuildImageRepository{UseInClusterRegistry: true}, pushRef)

	// No registry resolver/provider expectations: the push probe must be
	// skipped entirely. The git clone probe still runs.
	secrets.EXPECT().ValidateImageRegistrySecretForStackResource(gomock.Any(), "push-sec").Return(nil)

	v := NewStackValidator(StackValidatorSpec{
		SecretService:      secrets,
		CredentialResolver: resolver,
		RegistryClients:    provider,
		GitClients:         permissiveGitProbes(t, resolver),
	})

	if err := v.ValidateForCreate(context.Background(), spec); err != nil {
		t.Fatalf("expected in-cluster push target to validate without probing, got %v", err)
	}
}

func TestValidateForCreateAnonymousPrivateImageReturnsCredentialsRequired(t *testing.T) {
	resolver, provider, registryClient, _ := newProbeMocks(t)

	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.StackResources[0].ImageConfig.Image = "ghcr.io/acme/private:latest"

	resolved := &credentials.ResolvedRegistryCredential{Source: credentials.SourceAnonymous}
	resolver.EXPECT().
		RegistryCredentials(gomock.Any(), "org-1", "ghcr.io/acme/private:latest", credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).
		Return(resolved, nil)
	provider.EXPECT().ClientFor(resolved).Return(registryClient, nil)
	registryClient.EXPECT().CheckImage(gomock.Any(), "ghcr.io/acme/private:latest").
		Return(false, fmt.Errorf("authentication failed for image: %w", clients.ErrAuthFailed))

	v := NewStackValidator(StackValidatorSpec{
		CredentialResolver: resolver,
		RegistryClients:    provider,
	})

	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatal("expected private image without credentials to be rejected")
	}
	details, ok := err.Details.(errors.CredentialErrorDetails)
	if !ok {
		t.Fatalf("expected structured credential details, got %#v", err.Details)
	}
	if details.Code != errors.ErrorCodeCredentialsRequired {
		t.Fatalf("expected %s, got %s", errors.ErrorCodeCredentialsRequired, details.Code)
	}
	if details.Target.Kind != errors.CredentialTargetKindImagePull || details.Target.Host != "ghcr.io" {
		t.Fatalf("unexpected target %+v", details.Target)
	}
}

func TestValidateForCreateRejectsMalformedCommitSHA(t *testing.T) {
	spec := stackWithBuildResource(models.BuildImageRepository{UseInClusterRegistry: true}, nil)
	spec.StackResources[0].BuildConfig.SourceRevision.Git.Commit = "NOT-A-SHA"

	err := newTestValidator(t).ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatal("expected malformed commit SHA to be rejected")
	}
	if !containsSubstr(err.Error(), "invalid commit SHA") {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}

// --- Fix 1: push-probe error discrimination ---

// buildPushProbeValidator wires a validator whose push probe returns pushErr for
// the given anonymous/configured credential source. The git clone probe is
// stubbed permissive so only the push branch is under test.
func buildPushProbeValidator(t *testing.T, source credentials.Source, pushErr error) (validator.StackValidator, *models.Stack) {
	t.Helper()
	resolver, provider, registryClient, _ := newProbeMocks(t)

	spec := stackWithBuildResource(models.BuildImageRepository{ExternalImageRef: "ghcr.io/acme/app"}, nil)

	resolved := &credentials.ResolvedRegistryCredential{Source: source}
	if source != credentials.SourceAnonymous {
		resolved.Username = "alice"
		resolved.Password = "s3cret"
	}
	resolver.EXPECT().
		RegistryCredentials(gomock.Any(), "org-1", "ghcr.io/acme/app", credentials.RegistryPurposePush, credentials.RegistryAuthSelector{}).
		Return(resolved, nil)
	provider.EXPECT().ClientFor(resolved).Return(registryClient, nil)
	registryClient.EXPECT().CheckPushAccess(gomock.Any(), "ghcr.io/acme/app").Return(pushErr)

	v := NewStackValidator(StackValidatorSpec{
		CredentialResolver: resolver,
		RegistryClients:    provider,
		GitClients:         permissiveGitProbes(t, resolver),
	})
	return v, spec
}

func TestValidateForCreatePushAnonymousNonAuthFailureIsNotCredentialError(t *testing.T) {
	// A DNS/connection/TLS/5xx failure while pushing anonymously must surface as
	// its own validation error naming the host, NOT credentials_required.
	v, spec := buildPushProbeValidator(t, credentials.SourceAnonymous, fmt.Errorf("dial tcp: lookup ghcr.io: no such host"))

	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatal("expected non-auth push failure to be rejected")
	}
	if _, ok := err.Details.(errors.CredentialErrorDetails); ok {
		t.Fatalf("expected non-credential error, got credential details: %#v", err.Details)
	}
	if !containsSubstr(err.Error(), "ghcr.io") {
		t.Fatalf("expected error to name the registry host, got %q", err.Error())
	}
	if !containsSubstr(err.Error(), "no such host") {
		t.Fatalf("expected error to wrap the underlying cause, got %q", err.Error())
	}
}

func TestValidateForCreatePushAnonymousAuthFailureReturnsCredentialsRequired(t *testing.T) {
	v, spec := buildPushProbeValidator(t, credentials.SourceAnonymous,
		fmt.Errorf("push denied: %w", clients.ErrAuthFailed))

	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatal("expected anonymous push auth failure to be rejected")
	}
	details, ok := err.Details.(errors.CredentialErrorDetails)
	if !ok {
		t.Fatalf("expected structured credential details, got %#v", err.Details)
	}
	if details.Code != errors.ErrorCodeCredentialsRequired {
		t.Fatalf("expected %s, got %s", errors.ErrorCodeCredentialsRequired, details.Code)
	}
	if details.Target.Kind != errors.CredentialTargetKindImagePush || details.Target.Host != "ghcr.io" {
		t.Fatalf("unexpected target %+v", details.Target)
	}
}

func TestValidateForCreatePushConfiguredCredsAuthFailureReturnsCredentialsInvalid(t *testing.T) {
	v, spec := buildPushProbeValidator(t, credentials.SourceSecretRef,
		fmt.Errorf("push denied: %w", clients.ErrAuthFailed))

	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatal("expected configured-cred push auth failure to be rejected")
	}
	details, ok := err.Details.(errors.CredentialErrorDetails)
	if !ok {
		t.Fatalf("expected structured credential details, got %#v", err.Details)
	}
	if details.Code != errors.ErrorCodeCredentialsInvalid {
		t.Fatalf("expected %s, got %s", errors.ErrorCodeCredentialsInvalid, details.Code)
	}
}

func TestValidateForCreatePushAnonymousRateLimitSoftPasses(t *testing.T) {
	v, spec := buildPushProbeValidator(t, credentials.SourceAnonymous,
		fmt.Errorf("push throttled: %w", clients.ErrRateLimited))

	if err := v.ValidateForCreate(context.Background(), spec); err != nil {
		t.Fatalf("expected anonymous rate-limited push probe to soft-pass, got %v", err)
	}
}

// --- Fix 2: update-path probes only fire on changed sources ---

// updateProbeMocks returns registry + git probe mocks with NO call expectations,
// so any probe invocation fails the test. The credential resolver is left
// permissive-but-unexpected; callers add expectations for resources they expect
// to be probed.
func updateProbeMocks(t *testing.T) (
	*mocks.MockCredentialResolver,
	*mocks.MockregistryClientProvider,
	*mocks.MockRegistryClient,
	*mocks.MockgitClientProvider,
	*mocks.MockGitClient,
) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	return mocks.NewMockCredentialResolver(ctrl),
		mocks.NewMockregistryClientProvider(ctrl),
		mocks.NewMockRegistryClient(ctrl),
		mocks.NewMockgitClientProvider(ctrl),
		mocks.NewMockGitClient(ctrl)
}

func imageStack() *models.Stack {
	return &models.Stack{
		Name:           "test-stack",
		OrganisationID: "org-1",
		UserID:         "user-1",
		StackResources: []*models.StackResource{
			{
				Name:        "web",
				ImageConfig: &models.ImageConfigSpec{Image: "nginx:latest"},
			},
		},
	}
}

func TestValidateForUpdateSkipsProbesWhenOnlyReplicasChanged(t *testing.T) {
	resolver, provider, _, gitProvider, _ := updateProbeMocks(t)

	existing := imageStack()
	desired := imageStack()
	replicas := int32(3)
	desired.StackResources[0].Replicas = &replicas

	v := NewStackValidator(StackValidatorSpec{
		CredentialResolver: resolver,
		RegistryClients:    provider,
		GitClients:         gitProvider,
	})

	if err := v.ValidateForUpdate(context.Background(), existing, desired); err != nil {
		t.Fatalf("expected unchanged-source update to validate without probing, got %v", err)
	}
}

func TestValidateForUpdateSkipsProbesOnConnectionMutation(t *testing.T) {
	resolver, provider, _, gitProvider, _ := updateProbeMocks(t)

	existing := imageStack()
	desired := imageStack()
	// Simulate a connection-mutation update: resources are byte-identical, only
	// a self-referential env connection is added.
	desired.StackResources[0].Ports = []models.Port{{Name: "http", Number: 8080, Protocol: "http"}}
	existing.StackResources[0].Ports = desired.StackResources[0].Ports
	desired.Connections = models.StackConnections{
		{
			ID:   "self",
			Kind: models.ConnectionKindEnv,
			From: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
			To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
			Mappings: []models.ConnectionMapping{
				{
					Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "SELF_URL"},
					Value:  models.ValueRef{Output: "url.http"},
				},
			},
		},
	}

	v := NewStackValidator(StackValidatorSpec{
		CredentialResolver: resolver,
		RegistryClients:    provider,
		GitClients:         gitProvider,
	})

	if err := v.ValidateForUpdate(context.Background(), existing, desired); err != nil {
		t.Fatalf("expected connection-mutation update to validate without probing, got %v", err)
	}
}

func TestValidateForUpdateProbesOnlyChangedResource(t *testing.T) {
	resolver, provider, registryClient, gitProvider, gitClient := updateProbeMocks(t)

	buildResource := func(repoURL string) *models.StackResource {
		return &models.StackResource{
			Name: "api",
			BuildConfig: &models.BuildConfigSpec{
				SourceContext:        models.BuildContextSource{Git: &models.GitBuildSource{RepoURL: repoURL}},
				SourceRevision:       models.BuildSourceRevision{Git: &models.GitRevision{Branch: "main"}},
				BuildImageRepository: models.BuildImageRepository{UseInClusterRegistry: true},
			},
		}
	}

	existing := imageStack()
	existing.StackResources = append(existing.StackResources, buildResource("https://github.com/acme/api"))

	desired := imageStack()
	desired.StackResources = append(desired.StackResources, buildResource("https://github.com/acme/api-renamed"))

	// Only the changed 'api' resource's git clone probe should fire. The
	// unchanged 'web' image probe (registryClient/provider) must not be called.
	_ = registryClient
	_ = provider
	resolver.EXPECT().
		GitCredentials(gomock.Any(), "org-1", "https://github.com/acme/api-renamed", credentials.GitAuthSelector{}).
		Return(&credentials.ResolvedGitCredential{Source: credentials.SourceAnonymous}, nil)
	gitProvider.EXPECT().ClientFor("https://github.com/acme/api-renamed", gomock.Any()).Return(gitClient, nil)
	gitClient.EXPECT().CheckAccess(gomock.Any(), "https://github.com/acme/api-renamed").Return(true, nil)

	v := NewStackValidator(StackValidatorSpec{
		CredentialResolver: resolver,
		RegistryClients:    provider,
		GitClients:         gitProvider,
	})

	if err := v.ValidateForUpdate(context.Background(), existing, desired); err != nil {
		t.Fatalf("expected changed-source update to validate, got %v", err)
	}
}

func TestValidateForUpdateProbesWhenInlineCredentialsResubmitted(t *testing.T) {
	resolver, provider, registryClient, gitProvider, _ := updateProbeMocks(t)

	existing := imageStack()
	desired := imageStack()
	// Same image ref, but inline credentials were re-submitted: the managed
	// secret's data may have changed even though the ref string did not.
	desired.StackResources[0].ImageConfig.InlineCredentials = &models.InlineCredentials{
		Username: "alice",
		Password: "s3cret",
	}

	resolved := &credentials.ResolvedRegistryCredential{Source: credentials.SourceAnonymous}
	resolver.EXPECT().
		RegistryCredentials(gomock.Any(), "org-1", "nginx:latest", credentials.RegistryPurposePull, gomock.Any()).
		Return(resolved, nil)
	provider.EXPECT().ClientFor(resolved).Return(registryClient, nil)
	registryClient.EXPECT().CheckImage(gomock.Any(), "nginx:latest").Return(true, nil)

	v := NewStackValidator(StackValidatorSpec{
		CredentialResolver: resolver,
		RegistryClients:    provider,
		GitClients:         gitProvider,
	})

	if err := v.ValidateForUpdate(context.Background(), existing, desired); err != nil {
		t.Fatalf("expected inline-credential re-submission to probe and validate, got %v", err)
	}
}

func TestValidateForCreateAnonymousPrivateRepoReturnsCredentialsRequired(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	resolver := mocks.NewMockCredentialResolver(ctrl)
	gitClient := mocks.NewMockGitClient(ctrl)
	gitClients := mocks.NewMockgitClientProvider(ctrl)

	spec := stackWithBuildResource(models.BuildImageRepository{UseInClusterRegistry: true}, nil)

	resolver.EXPECT().
		GitCredentials(gomock.Any(), "org-1", "https://github.com/acme/api", credentials.GitAuthSelector{}).
		Return(&credentials.ResolvedGitCredential{Source: credentials.SourceAnonymous}, nil)
	gitClients.EXPECT().ClientFor("https://github.com/acme/api", gomock.Any()).Return(gitClient, nil)
	gitClient.EXPECT().CheckAccess(gomock.Any(), "https://github.com/acme/api").
		Return(false, fmt.Errorf("repository not found: %w", gitclient.ErrNotFound))

	v := NewStackValidator(StackValidatorSpec{
		CredentialResolver: resolver,
		GitClients:         gitClients,
	})

	err := v.ValidateForCreate(context.Background(), spec)
	if err == nil {
		t.Fatal("expected anonymous private repo to be rejected")
	}
	details, ok := err.Details.(errors.CredentialErrorDetails)
	if !ok {
		t.Fatalf("expected structured credential details, got %#v", err.Details)
	}
	if details.Code != errors.ErrorCodeCredentialsRequired || details.Target.Kind != errors.CredentialTargetKindGitClone {
		t.Fatalf("unexpected details %+v", details)
	}
}
