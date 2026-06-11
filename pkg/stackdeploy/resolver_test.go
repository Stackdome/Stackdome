package stackdeploy

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

func TestResolveDoesNotMutateInputStack(t *testing.T) {
	g := NewWithT(t)
	stack := &models.Stack{
		ID:        "stack-1",
		Name:      "app",
		Namespace: "ns-1",
		StackResources: []*models.StackResource{
			{
				ID:        "api-id",
				StackID:   "stack-1",
				Name:      "api",
				Namespace: "ns-1",
				ExecutionConfig: &models.ExecutionConfig{
					Env: []models.EnvVar{{Name: "PUBLIC_URL", SelfOutput: "public.http.url"}},
				},
				Ports: models.Ports{
					{Name: "http", Number: 8080, Protocol: "http", ExposedToPublic: true, ExposedFqdn: "api.example.com"},
				},
			},
		},
	}

	resolver := NewResolver(ResolverSpec{})
	effective, err := resolver.Resolve(context.Background(), stack)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(effective.StackResources[0].ExecutionConfig.Env[0].Value).To(Equal("http://api.example.com"))
	g.Expect(stack.StackResources[0].ExecutionConfig.Env[0].Value).To(Equal(""))
}

func TestResolveFailsOnUnknownSelfOutput(t *testing.T) {
	g := NewWithT(t)
	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{
				Name: "api",
				ExecutionConfig: &models.ExecutionConfig{
					Env: []models.EnvVar{{Name: "X", SelfOutput: "does.not.exist"}},
				},
			},
		},
	}

	resolver := NewResolver(ResolverSpec{})
	_, err := resolver.Resolve(context.Background(), stack)

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("does.not.exist"))
}

func TestResolveSecretEnvConnection(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	secretService := NewMockSecretService(ctrl)
	secretService.EXPECT().
		InternalGetByID(gomock.Any(), "sec-1").
		Return(&models.Secret{
			ID:   "sec-1",
			Data: map[string]string{"tls.crt": "cert-data"},
		}, nil)

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
						Value:  models.ValueRef{Output: "tls.crt"},
					},
				},
			},
		},
	}

	resolver := NewResolver(ResolverSpec{SecretService: secretService})
	effective, err := resolver.Resolve(context.Background(), stack)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(envValue(effective, "web", "TLS_CERT")).To(Equal("cert-data"))
	g.Expect(len(stack.StackResources[0].ExecutionConfig.Env)).To(Equal(0))
}

func TestResolveStackResourceEnvConnection(t *testing.T) {
	g := NewWithT(t)

	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{
				ID:        "res-api",
				Name:      "api",
				Namespace: "default",
				Ports: models.Ports{
					{Name: "http", Number: 8080, Protocol: "http", ExposedToPublic: true, ExposedFqdn: "api.example.com"},
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
					{
						Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "API_TEMPLATE_URL"},
						Value: models.ValueRef{
							Template: "http://{{public_host}}:{{port}}",
							Values: map[string]models.OutputValueRef{
								"public_host": {Output: "public.http.host"}, "port": {Output: "port.http"},
							},
						},
					},
				},
			},
		},
	}

	resolver := NewResolver(ResolverSpec{})
	effective, err := resolver.Resolve(context.Background(), stack)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(envValue(effective, "web", "API_HOST")).To(Equal("api.default.svc"))
	g.Expect(envValue(effective, "web", "API_URL")).To(Equal("http://api.default.svc:8080"))
	g.Expect(envValue(effective, "web", "API_PUBLIC_URL")).To(Equal("http://api.example.com"))
	g.Expect(envValue(effective, "web", "API_TEMPLATE_URL")).To(Equal("http://api.example.com:8080"))
	g.Expect(len(stack.StackResources[1].ExecutionConfig.Env)).To(Equal(0))
}

func TestResolvePostgresEnvConnection(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	addonService := NewMockPostgresAddonService(ctrl)
	addonService.EXPECT().
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
					{
						Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "PGHOST"},
						Value:  models.ValueRef{Output: "host"},
					},
				},
			},
		},
	}

	resolver := NewResolver(ResolverSpec{PostgresAddonService: addonService})
	effective, err := resolver.Resolve(context.Background(), stack)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(envValue(effective, "web", "DATABASE_URL")).To(Equal("postgresql://app_user:secret@pg-rw.default.svc.cluster.local:5432/app"))
	g.Expect(envValue(effective, "web", "PGHOST")).To(Equal("pg-rw.default.svc.cluster.local"))
	g.Expect(len(stack.StackResources[0].ExecutionConfig.Env)).To(Equal(1))
}

func TestResolvePostgresCredentialsNotReadyReturnsDependencyNotReady(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	addonService := NewMockPostgresAddonService(ctrl)
	addonService.EXPECT().
		InternalGetCredentials(gomock.Any(), "pg-1", "app", false).
		Return(nil, errors.GeneralError("credentials not ready"))

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

	resolver := NewResolver(ResolverSpec{PostgresAddonService: addonService})
	_, err := resolver.Resolve(context.Background(), stack)

	g.Expect(err).To(HaveOccurred())
	var notReady DependencyNotReadyError
	g.Expect(stderrors.As(err, &notReady)).To(BeTrue())
}

func TestResolveDuplicateEnvVarFromConnectionFails(t *testing.T) {
	g := NewWithT(t)

	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{
				ID:        "res-api",
				Name:      "api",
				Namespace: "default",
				Ports: models.Ports{
					{Name: "http", Number: 8080, Protocol: "http"},
				},
			},
			{
				ID:   "res-web",
				Name: "web",
				ExecutionConfig: &models.ExecutionConfig{
					Env: []models.EnvVar{{Name: "API_HOST", Value: "hardcoded"}},
				},
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
				},
			},
		},
	}

	resolver := NewResolver(ResolverSpec{})
	_, err := resolver.Resolve(context.Background(), stack)

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("duplicate"))
	g.Expect(err.Error()).To(ContainSubstring("API_HOST"))
}

func TestResolveVolumeMountConnection(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	volumeSvc := NewMockVolumeService(ctrl)
	volumeSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), "stack-1").
		Return([]*models.Volume{{ID: "vol-1", Name: "uploads", VolumeSource: nil}}, nil)

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

	resolver := NewResolver(ResolverSpec{VolumeService: volumeSvc})
	effective, err := resolver.Resolve(context.Background(), stack)

	g.Expect(err).NotTo(HaveOccurred())
	mounts := effective.StackResources[0].VolumeMounts
	g.Expect(mounts).To(HaveLen(1))
	g.Expect(mounts[0].SourceVolumeName).To(Equal("uploads"))
	g.Expect(mounts[0].SourceVolumeID).To(Equal("vol-1"))
	g.Expect(mounts[0].TargetPath).To(Equal("/uploads"))
	g.Expect(mounts[0].SourceSubPath).To(Equal("data"))
	// Input stack untouched.
	g.Expect(stack.StackResources[0].VolumeMounts).To(BeNil())
}

func TestResolveBuildArtifactSourceConnection(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	volumeSvc := NewMockVolumeService(ctrl)
	volumeSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), "stack-1").
		Return([]*models.Volume{{ID: "vol-1", Name: "assets"}}, nil)

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

	resolver := NewResolver(ResolverSpec{VolumeService: volumeSvc})
	effective, err := resolver.Resolve(context.Background(), stack)

	g.Expect(err).NotTo(HaveOccurred())
	volume := effective.Volumes[0]
	g.Expect(volume.VolumeSource).NotTo(BeNil())
	g.Expect(volume.VolumeSource.BuildSource).To(HaveLen(1))
	bs := volume.VolumeSource.BuildSource[0]
	g.Expect(bs.ResourceName).To(Equal("builder"))
	g.Expect(bs.SourcePath).To(Equal("/app/public"))
	g.Expect(bs.DestinationPath).To(Equal("/"))
}

func TestResolveVolumeMountUnknownVolumeFails(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	volumeSvc := NewMockVolumeService(ctrl)
	volumeSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), "stack-1").
		Return([]*models.Volume{}, nil)

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

	resolver := NewResolver(ResolverSpec{VolumeService: volumeSvc})
	_, err := resolver.Resolve(context.Background(), stack)

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("nonexistent"))
}

func TestResolveVolumeMountUnknownResourceFails(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	volumeSvc := NewMockVolumeService(ctrl)
	volumeSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), "stack-1").
		Return([]*models.Volume{{ID: "vol-1", Name: "uploads"}}, nil)

	stack := &models.Stack{
		ID:             "stack-1",
		StackResources: []*models.StackResource{},
		Connections: models.StackConnections{
			{
				ID:   "bad-ref",
				Kind: models.ConnectionKindVolumeMount,
				From: models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Name: "uploads"},
				To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "missing-resource"},
				Config: map[string]interface{}{
					"mount_path": "/data",
				},
			},
		},
	}

	resolver := NewResolver(ResolverSpec{VolumeService: volumeSvc})
	_, err := resolver.Resolve(context.Background(), stack)

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("missing-resource"))
}

// --- test helpers ---

func envValue(stack *models.Stack, resourceName, envName string) string {
	for _, resource := range stack.StackResources {
		if resource.Name != resourceName {
			continue
		}
		if resource.ExecutionConfig == nil {
			return ""
		}
		for _, env := range resource.ExecutionConfig.Env {
			if env.Name == envName {
				return env.Value
			}
		}
	}
	return ""
}
