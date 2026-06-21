package stack

import (
	"context"
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/mocks"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/validator"
	"go.uber.org/mock/gomock"
)

func TestValidateForCreateRequiresNamedPorts(t *testing.T) {
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
	spec := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	spec.Settings = nil

	err := v.ValidateForCreate(context.Background(), spec)
	if err != nil {
		t.Fatalf("expected nil settings to pass, got %v", err)
	}
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
	return NewStackValidator(StackValidatorSpec{
		PostgresAddonService: postgresAddons,
	}), postgresAddons
}

func TestValidateForCreateAllowsBuildArtifactSourceConnection(t *testing.T) {
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
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
	v := NewStackValidator(StackValidatorSpec{})
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

func newValidatorWithMockedSecretService(t *testing.T) (validator.StackValidator, *mocks.MocksecretService) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	secrets := mocks.NewMocksecretService(ctrl)
	return NewStackValidator(StackValidatorSpec{
		SecretService: secrets,
	}), secrets
}
