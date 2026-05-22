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

func TestValidateForCreateAllowsPostgresConnectionConfig(t *testing.T) {
	v, postgresAddons := newValidatorWithMockedPostgresAddonService(t)
	spec := stackWithConnections(models.StackConnection{
		Id:   "pg-env",
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
		Id:   "pg-env",
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
		Id:   "pg-env",
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
		Id:   "pg-env",
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
		Id:   "volume-mount",
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
		Id:   "volume-mount",
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
		Id:   "volume-mount",
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

func TestValidateForUpdateAllowsVolumeMountConnectionUsingExistingDBVolume(t *testing.T) {
	v := NewStackValidator(StackValidatorSpec{})
	existing := stackWithPorts(models.Port{Name: "http", Number: 8080, Protocol: "http"})
	existing.Volumes = []*models.Volume{
		{Name: "uploads"},
	}

	patch := stackWithConnections(models.StackConnection{
		Id:   "volume-mount",
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
