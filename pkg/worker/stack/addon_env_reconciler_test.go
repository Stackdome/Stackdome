package stack

import (
	"context"
	"testing"

	serrors "github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func TestAddonEnvReconcilerResolvesPostgresEnvConnections(t *testing.T) {
	reconciler := NewAddonEnvReconciler(AddonEnvReconcilerSpec{
		PostgresAddonService: fakeWorkerPostgresAddonService{
			credentials: map[string]*models.PostgresCredentials{
				"pg-1": {
					Database:         "app",
					Host:             "pg-rw.default.svc.cluster.local",
					Port:             5432,
					Username:         "app_user",
					Password:         "secret",
					SSLMode:          "require",
					ConnectionString: "postgresql://app_user:secret@pg-rw.default.svc.cluster.local:5432/app",
				},
			},
		},
		AddonUsageService: newFakeAddonUsageService(),
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
				Id:   "pg-web",
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

func TestAddonEnvReconcilerSupportsTemplateConnectionValues(t *testing.T) {
	reconciler := NewAddonEnvReconciler(AddonEnvReconcilerSpec{
		PostgresAddonService: fakeWorkerPostgresAddonService{
			credentials: map[string]*models.PostgresCredentials{
				"pg-1": {
					Database:      "app",
					Host:          "pg-rw.default.svc.cluster.local",
					Port:          5432,
					Username:      "app_user",
					Password:      "secret",
					SSLMode:       "require",
					CACertificate: "ca-data",
				},
			},
		},
		AddonUsageService: newFakeAddonUsageService(),
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
				Id:   "pg-web",
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

func TestAddonEnvReconcilerRequeuesWhenConnectionCredentialsAreUnavailable(t *testing.T) {
	usages := newFakeAddonUsageService()
	reconciler := NewAddonEnvReconciler(AddonEnvReconcilerSpec{
		PostgresAddonService: fakeWorkerPostgresAddonService{
			err: serrors.BadRequest("not ready"),
		},
		AddonUsageService: usages,
	})

	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{ID: "res-1", Name: "web"},
		},
		Connections: models.StackConnections{
			{
				Id:   "pg-web",
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

func TestAddonEnvReconcilerResolvesStackResourceEnvConnections(t *testing.T) {
	reconciler := NewAddonEnvReconciler(AddonEnvReconcilerSpec{
		AddonUsageService: newFakeAddonUsageService(),
	})

	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{
				ID:   "res-api",
				Name: "api",
				Ports: models.Ports{
					{
						Name:            "http",
						Number:          8080,
						Protocol:        "http",
						ExposedToPublic: true,
						ExposedFqdn:     "api.example.com",
					},
				},
				Status: &models.StackResourceStatus{
					InternalServiceName: stringPtr("api.default.svc.cluster.local"),
					PublicIngresses: []models.Ingress{
						{TargetPort: 8080, URL: "https://api.example.com"},
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
				Id:   "api-web",
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
	if env[0].Name != "API_HOST" || env[0].Value != "api.default.svc.cluster.local" {
		t.Fatalf("unexpected API_HOST env var: %#v", env[0])
	}
	if env[1].Name != "API_URL" || env[1].Value != "http://api.default.svc.cluster.local:8080" {
		t.Fatalf("unexpected API_URL env var: %#v", env[1])
	}
	if env[2].Name != "API_PUBLIC_URL" || env[2].Value != "https://api.example.com" {
		t.Fatalf("unexpected API_PUBLIC_URL env var: %#v", env[2])
	}
}

func TestAddonEnvReconcilerResolvesSecretEnvConnections(t *testing.T) {
	reconciler := NewAddonEnvReconciler(AddonEnvReconcilerSpec{
		SecretService:     fakeWorkerSecretService{secrets: map[string]*models.Secret{"sec-1": {ID: "sec-1", Data: map[string]string{"tls.crt": "cert-data"}}}},
		AddonUsageService: newFakeAddonUsageService(),
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
				Id:   "tls-web",
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

func TestAddonEnvReconcilerResolvesSelfOutputEnvVars(t *testing.T) {
	reconciler := NewAddonEnvReconciler(AddonEnvReconcilerSpec{
		AddonUsageService: newFakeAddonUsageService(),
	})

	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{
				ID:   "res-web",
				Name: "web",
				Ports: models.Ports{
					{
						Name:            "http",
						Number:          3000,
						Protocol:        "http",
						ExposedToPublic: true,
						ExposedFqdn:     "app.example.com",
					},
				},
				Status: &models.StackResourceStatus{
					InternalServiceName: stringPtr("web.default.svc.cluster.local"),
					PublicIngresses: []models.Ingress{
						{TargetPort: 3000, URL: "https://app.example.com"},
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
	if env[0].Name != "PUBLIC_URL" || env[0].Value != "https://app.example.com" {
		t.Fatalf("unexpected PUBLIC_URL env var: %#v", env[0])
	}
	if env[1].Name != "INTERNAL_URL" || env[1].Value != "http://web.default.svc.cluster.local:3000" {
		t.Fatalf("unexpected INTERNAL_URL env var: %#v", env[1])
	}
}

type fakeWorkerPostgresAddonService struct {
	credentials map[string]*models.PostgresCredentials
	err         *serrors.ServiceError
}

func (f fakeWorkerPostgresAddonService) InternalGetPostgresAddon(_ context.Context, id string) (*models.PostgresAddon, *serrors.ServiceError) {
	return &models.PostgresAddon{ID: id}, nil
}

func (f fakeWorkerPostgresAddonService) InternalGetCredentials(_ context.Context, addonID string, _ string, _ bool) (*models.PostgresCredentials, *serrors.ServiceError) {
	if f.err != nil {
		return nil, f.err
	}
	creds, ok := f.credentials[addonID]
	if !ok {
		return nil, serrors.NotFound("postgres addon credentials not found")
	}
	return creds, nil
}

type fakeAddonUsageService struct {
	usages []*models.AddonUsage
}

func newFakeAddonUsageService() *fakeAddonUsageService {
	return &fakeAddonUsageService{}
}

func (f *fakeAddonUsageService) Create(_ context.Context, usage *models.AddonUsage) error {
	f.usages = append(f.usages, usage)
	return nil
}

func (f *fakeAddonUsageService) Delete(_ context.Context, addonType models.AddonType, addonID, stackID, resourceID string) error {
	filtered := make([]*models.AddonUsage, 0, len(f.usages))
	for _, usage := range f.usages {
		if usage.AddonType == addonType && usage.AddonID == addonID && usage.StackID == stackID && usage.StackResourceID == resourceID {
			continue
		}
		filtered = append(filtered, usage)
	}
	f.usages = filtered
	return nil
}

func (f *fakeAddonUsageService) GetByStackID(_ context.Context, stackID string) ([]*models.AddonUsage, error) {
	filtered := make([]*models.AddonUsage, 0, len(f.usages))
	for _, usage := range f.usages {
		if usage.StackID == stackID {
			filtered = append(filtered, usage)
		}
	}
	return filtered, nil
}

func (f *fakeAddonUsageService) ExistsByStackResourceAndAddon(_ context.Context, stackID, resourceID, addonID string) (bool, error) {
	for _, usage := range f.usages {
		if usage.StackID == stackID && usage.StackResourceID == resourceID && usage.AddonID == addonID {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeAddonUsageService) DeleteByStackID(_ context.Context, stackID string) error {
	filtered := make([]*models.AddonUsage, 0, len(f.usages))
	for _, usage := range f.usages {
		if usage.StackID == stackID {
			continue
		}
		filtered = append(filtered, usage)
	}
	f.usages = filtered
	return nil
}

type fakeWorkerSecretService struct {
	secrets map[string]*models.Secret
}

func (f fakeWorkerSecretService) InternalGetByID(_ context.Context, id string) (*models.Secret, *serrors.ServiceError) {
	secret, ok := f.secrets[id]
	if !ok {
		return nil, serrors.NotFound("secret not found")
	}
	return secret, nil
}

func stringPtr(in string) *string {
	return &in
}
