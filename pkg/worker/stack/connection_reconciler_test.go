package stack

import (
	"context"
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/builders"
	serrors "github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"go.uber.org/mock/gomock"
)

func TestConnectionReconcilerResolvesVolumeMountConnections(t *testing.T) {
	reconciler := NewConnectionReconciler(ConnectionReconcilerSpec{
		VolumeService: fakeVolumeService{
			volumes: []*models.Volume{
				{ID: "vol-1", Name: "uploads", VolumeSource: nil},
			},
		},
	})

	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{ID: "res-1", Name: "web"},
		},
		Connections: models.StackConnections{
			{
				ID:   "vol-web",
				Kind: models.ConnectionKindVolumeMount,
				From: models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Name: "uploads"},
				To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
				Config: map[string]interface{}{
					"mount_path": "/uploads",
					"sub_path":   "data",
				},
			},
		},
	}

	result, err := reconciler.Reconcile(context.Background(), stack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.resultNil {
		t.Fatalf("expected nil result, got %#v", result)
	}

	mounts := stack.StackResources[0].VolumeMounts
	if len(mounts) != 1 {
		t.Fatalf("expected 1 volume mount, got %d", len(mounts))
	}
	if mounts[0].SourceVolumeName != "uploads" {
		t.Fatalf("expected source volume name 'uploads', got '%s'", mounts[0].SourceVolumeName)
	}
	if mounts[0].SourceVolumeID != "vol-1" {
		t.Fatalf("expected source volume ID 'vol-1', got '%s'", mounts[0].SourceVolumeID)
	}
	if mounts[0].TargetPath != "/uploads" {
		t.Fatalf("expected target path '/uploads', got '%s'", mounts[0].TargetPath)
	}
	if mounts[0].SourceSubPath != "data" {
		t.Fatalf("expected sub path 'data', got '%s'", mounts[0].SourceSubPath)
	}
}

func TestConnectionReconcilerResolvesBuildArtifactSourceConnections(t *testing.T) {
	reconciler := NewConnectionReconciler(ConnectionReconcilerSpec{
		VolumeService: fakeVolumeService{
			volumes: []*models.Volume{
				{ID: "vol-1", Name: "assets"},
			},
		},
	})

	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{ID: "res-1", Name: "builder", BuildConfig: &models.BuildConfigSpec{}},
		},
		Connections: models.StackConnections{
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
		},
	}

	result, err := reconciler.Reconcile(context.Background(), stack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.resultNil {
		t.Fatalf("expected nil result, got %#v", result)
	}

	volume := stack.Volumes[0]
	if volume.VolumeSource == nil {
		t.Fatalf("expected volume source to be set")
	}
	if len(volume.VolumeSource.BuildSource) != 1 {
		t.Fatalf("expected 1 build source, got %d", len(volume.VolumeSource.BuildSource))
	}
	bs := volume.VolumeSource.BuildSource[0]
	if bs.ResourceName != "builder" {
		t.Fatalf("expected resource name 'builder', got '%s'", bs.ResourceName)
	}
	if bs.SourcePath != "/app/public" {
		t.Fatalf("expected source path '/app/public', got '%s'", bs.SourcePath)
	}
	if bs.DestinationPath != "/" {
		t.Fatalf("expected destination path '/', got '%s'", bs.DestinationPath)
	}
}

func TestConnectionReconcilerSkipsWhenNoConnections(t *testing.T) {
	reconciler := NewConnectionReconciler(ConnectionReconcilerSpec{
		VolumeService: fakeVolumeService{},
	})

	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{ID: "res-1", Name: "web"},
		},
	}

	result, err := reconciler.Reconcile(context.Background(), stack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.resultNil {
		t.Fatalf("expected nil result")
	}
}

func TestConnectionReconcilerErrorsOnUnknownVolume(t *testing.T) {
	reconciler := NewConnectionReconciler(ConnectionReconcilerSpec{
		VolumeService: fakeVolumeService{volumes: []*models.Volume{}},
	})

	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{ID: "res-1", Name: "web"},
		},
		Connections: models.StackConnections{
			{
				ID:   "bad-vol",
				Kind: models.ConnectionKindVolumeMount,
				From: models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Name: "nonexistent"},
				To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
				Config: map[string]interface{}{
					"mount_path": "/data",
				},
			},
		},
	}

	_, err := reconciler.Reconcile(context.Background(), stack)
	if err == nil {
		t.Fatal("expected error for unknown volume")
	}
}

func TestConnectionReconcilerResolvesPostgresEnvConnections(t *testing.T) {
	ctrl := gomock.NewController(t)
	postgresAddons := NewMockpostgresAddonService(ctrl)
	postgresAddons.EXPECT().
		InternalGetCredentials(gomock.Any(), "pg-1", "app", false).
		Return(&models.PostgresCredentials{
			Database:         "app",
			Host:             "pg-rw.default.svc.cluster.local",
			Port:             5432,
			Username:         "app_user",
			Password:         "secret",
			SSLMode:          "require",
			ConnectionString: "postgresql://app_user:secret@pg-rw.default.svc.cluster.local:5432/app",
		}, nil)

	reconciler := NewConnectionReconciler(ConnectionReconcilerSpec{
		PostgresAddonService: postgresAddons,
	})

	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{
				ID:   "res-1",
				Name: "web",
				ExecutionConfig: &models.ExecutionConfig{
					Env: []models.EnvVar{{Name: "APP_ENV", Value: "prod"}},
				},
			},
		},
		Connections: models.StackConnections{
			{
				ID:   "pg-web",
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
				Mappings: []models.ConnectionMapping{
					{
						Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "DATABASE_URL"},
						Value:  models.ValueRef{Output: "url"},
					},
					{
						Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "PGHOST"},
						Value:  models.ValueRef{Output: "host"},
					},
				},
			},
		},
	}

	result, err := reconciler.Reconcile(context.Background(), stack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.resultNil {
		t.Fatalf("expected nil result, got %#v", result)
	}

	env := stack.StackResources[0].ExecutionConfig.Env
	if len(env) != 3 {
		t.Fatalf("expected 3 env vars after resolution, got %d", len(env))
	}
	if env[1].Name != "DATABASE_URL" || env[1].Value != "postgresql://app_user:secret@pg-rw.default.svc.cluster.local:5432/app" {
		t.Fatalf("unexpected DATABASE_URL env var: %#v", env[1])
	}
	if env[2].Name != "PGHOST" || env[2].Value != "pg-rw.default.svc.cluster.local" {
		t.Fatalf("unexpected PGHOST env var: %#v", env[2])
	}
}

func TestConnectionReconcilerSupportsTemplateConnectionValues(t *testing.T) {
	ctrl := gomock.NewController(t)
	postgresAddons := NewMockpostgresAddonService(ctrl)
	postgresAddons.EXPECT().
		InternalGetCredentials(gomock.Any(), "pg-1", "app", false).
		Return(&models.PostgresCredentials{
			Database:      "app",
			Host:          "pg-rw.default.svc.cluster.local",
			Port:          5432,
			Username:      "app_user",
			Password:      "secret",
			SSLMode:       "require",
			CACertificate: "ca-data",
		}, nil)

	reconciler := NewConnectionReconciler(ConnectionReconcilerSpec{
		PostgresAddonService: postgresAddons,
	})

	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{
				ID:              "res-1",
				Name:            "web",
				ExecutionConfig: &models.ExecutionConfig{},
			},
		},
		Connections: models.StackConnections{
			{
				ID:   "pg-web",
				Kind: models.ConnectionKindEnv,
				From: models.TopologyNodeRef{Type: models.TopologyNodeTypePostgresAddon, Id: "pg-1"},
				To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
				Config: map[string]interface{}{
					"database": "app",
				},
				Mappings: []models.ConnectionMapping{
					{
						Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "DATABASE_URL"},
						Value: models.ValueRef{
							Template: "postgres://{{ username }}:{{ password }}@{{ host }}:{{ port }}/{{ database }}",
							Values: map[string]models.OutputValueRef{
								"username": {Output: "username"},
								"password": {Output: "password"},
								"host":     {Output: "host"},
								"port":     {Output: "port"},
								"database": {Output: "database"},
							},
						},
					},
				},
			},
		},
	}

	_, err := reconciler.Reconcile(context.Background(), stack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	got := stack.StackResources[0].ExecutionConfig.Env[0]
	if got.Name != "DATABASE_URL" || got.Value != "postgres://app_user:secret@pg-rw.default.svc.cluster.local:5432/app" {
		t.Fatalf("unexpected templated env var: %#v", got)
	}
}

func TestConnectionReconcilerRequeuesWhenCredentialsUnavailable(t *testing.T) {
	ctrl := gomock.NewController(t)
	postgresAddons := NewMockpostgresAddonService(ctrl)
	postgresAddons.EXPECT().
		InternalGetCredentials(gomock.Any(), "pg-1", "app", false).
		Return(nil, serrors.BadRequest("not ready"))

	reconciler := NewConnectionReconciler(ConnectionReconcilerSpec{
		PostgresAddonService: postgresAddons,
	})

	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{ID: "res-1", Name: "web"},
		},
		Connections: models.StackConnections{
			{
				ID:   "pg-web",
				Kind: models.ConnectionKindEnv,
				From: models.TopologyNodeRef{Type: models.TopologyNodeTypePostgresAddon, Id: "pg-1"},
				To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
				Config: map[string]interface{}{
					"database": "app",
				},
				Mappings: []models.ConnectionMapping{
					{
						Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "DATABASE_URL"},
						Value:  models.ValueRef{Output: "url"},
					},
				},
			},
		},
	}

	result, err := reconciler.Reconcile(context.Background(), stack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.resultRequeueAfter == nil {
		t.Fatalf("expected requeue-after result, got %#v", result)
	}
	if *result.resultRequeueAfter != addonReadinessRequeueInterval {
		t.Fatalf("expected requeue-after %s, got %s", addonReadinessRequeueInterval, *result.resultRequeueAfter)
	}
}

func TestConnectionReconcilerResolvesStackResourceEnvConnections(t *testing.T) {
	reconciler := NewConnectionReconciler(ConnectionReconcilerSpec{})

	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{
				ID:        "res-api",
				Name:      "api",
				Namespace: "default",
				Ports: models.Ports{
					{
						Name:            "http",
						Number:          8080,
						Protocol:        "http",
						ExposedToPublic: true,
						ExposedFqdn:     "api.example.com",
					},
				},
			},
			{
				ID:              "res-web",
				Name:            "web",
				ExecutionConfig: &models.ExecutionConfig{},
			},
		},
		Connections: models.StackConnections{
			{
				ID:   "api-web",
				Kind: models.ConnectionKindEnv,
				From: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "api"},
				To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
				Mappings: []models.ConnectionMapping{
					{
						Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "API_HOST"},
						Value:  models.ValueRef{Output: "host"},
					},
					{
						Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "API_URL"},
						Value:  models.ValueRef{Output: "url.http"},
					},
					{
						Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "API_PUBLIC_URL"},
						Value:  models.ValueRef{Output: "public.http.url"},
					},
				},
			},
		},
	}

	_, err := reconciler.Reconcile(context.Background(), stack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	env := stack.StackResources[1].ExecutionConfig.Env
	if len(env) != 3 {
		t.Fatalf("expected 3 env vars after resolution, got %d", len(env))
	}
	if env[0].Name != "API_HOST" || env[0].Value != "api.default.svc" {
		t.Fatalf("unexpected API_HOST env var: %#v", env[0])
	}
	if env[1].Name != "API_URL" || env[1].Value != "http://api.default.svc:8080" {
		t.Fatalf("unexpected API_URL env var: %#v", env[1])
	}
	if env[2].Name != "API_PUBLIC_URL" || env[2].Value != "http://api.example.com" {
		t.Fatalf("unexpected API_PUBLIC_URL env var: %#v", env[2])
	}

	stackCR, buildErr := builders.NewClusterResourceBuilder(builders.ClusterResourceBuilderSpec{}).BuildStackCR(stack)
	if buildErr != nil {
		t.Fatalf("expected stack CR build to succeed, got %v", buildErr)
	}
	crEnv := stackCR.Spec.StackResources[1].Spec.EnvironmentVariables
	if len(crEnv) != 3 {
		t.Fatalf("expected 3 normal CR env vars, got %d", len(crEnv))
	}
	if crEnv[0].Name != "API_HOST" || crEnv[0].Value != "api.default.svc" {
		t.Fatalf("unexpected API_HOST CR env var: %#v", crEnv[0])
	}
	if crEnv[1].Name != "API_URL" || crEnv[1].Value != "http://api.default.svc:8080" {
		t.Fatalf("unexpected API_URL CR env var: %#v", crEnv[1])
	}
	if crEnv[2].Name != "API_PUBLIC_URL" || crEnv[2].Value != "http://api.example.com" {
		t.Fatalf("unexpected API_PUBLIC_URL CR env var: %#v", crEnv[2])
	}
}

func TestConnectionReconcilerResolvesSecretEnvConnections(t *testing.T) {
	ctrl := gomock.NewController(t)
	secrets := NewMocksecretOutputService(ctrl)
	secrets.EXPECT().
		InternalGetByID(gomock.Any(), "sec-1").
		Return(&models.Secret{ID: "sec-1", Data: map[string]string{"tls.crt": "cert-data"}}, nil)

	reconciler := NewConnectionReconciler(ConnectionReconcilerSpec{
		SecretService: secrets,
	})

	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{
				ID:              "res-web",
				Name:            "web",
				ExecutionConfig: &models.ExecutionConfig{},
			},
		},
		Connections: models.StackConnections{
			{
				ID:   "tls-web",
				Kind: models.ConnectionKindEnv,
				From: models.TopologyNodeRef{Type: models.TopologyNodeTypeSecret, Id: "sec-1"},
				To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
				Mappings: []models.ConnectionMapping{
					{
						Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "TLS_CERT"},
						Value:  models.ValueRef{Output: "key['tls.crt']"},
					},
				},
			},
		},
	}

	_, err := reconciler.Reconcile(context.Background(), stack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	env := stack.StackResources[0].ExecutionConfig.Env
	if len(env) != 1 {
		t.Fatalf("expected 1 env var after resolution, got %d", len(env))
	}
	if env[0].Name != "TLS_CERT" || env[0].Value != "cert-data" {
		t.Fatalf("unexpected TLS_CERT env var: %#v", env[0])
	}
}

func TestConnectionReconcilerResolvesSelfOutputEnvVars(t *testing.T) {
	reconciler := NewConnectionReconciler(ConnectionReconcilerSpec{})

	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{
				ID:        "res-web",
				Name:      "web",
				Namespace: "default",
				Ports: models.Ports{
					{
						Name:            "http",
						Number:          3000,
						Protocol:        "http",
						ExposedToPublic: true,
						ExposedFqdn:     "app.example.com",
					},
				},
				ExecutionConfig: &models.ExecutionConfig{
					Env: []models.EnvVar{
						{Name: "PUBLIC_URL", SelfOutput: "public.http.url"},
						{Name: "INTERNAL_URL", SelfOutput: "url.http"},
					},
				},
			},
		},
	}

	_, err := reconciler.Reconcile(context.Background(), stack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	env := stack.StackResources[0].ExecutionConfig.Env
	if env[0].Name != "PUBLIC_URL" || env[0].Value != "http://app.example.com" {
		t.Fatalf("unexpected PUBLIC_URL env var: %#v", env[0])
	}
	if env[1].Name != "INTERNAL_URL" || env[1].Value != "http://web.default.svc:3000" {
		t.Fatalf("unexpected INTERNAL_URL env var: %#v", env[1])
	}
}

type fakeVolumeService struct {
	volumes []*models.Volume
}

func (f fakeVolumeService) ListVolumesUsedByStack(_ context.Context, _ string) ([]*models.Volume, *serrors.ServiceError) {
	return f.volumes, nil
}

func (f fakeVolumeService) InternalDeleteVolumesUsedByStackFromDB(_ context.Context, _ string) *serrors.ServiceError {
	return nil
}
