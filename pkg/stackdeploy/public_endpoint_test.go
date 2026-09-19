package stackdeploy

import (
	"context"
	"strings"

	"github.com/Stackdome/stackdome/pkg/builders"
	"github.com/Stackdome/stackdome/pkg/models"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("resolving public URL schemes", func() {
	type endpointCase struct {
		name   string
		config models.PublicEndpointConfig
		fqdn   string
		want   string
	}

	ginkgo.DescribeTable("matches the ingress TLS decision for self outputs and connections",
		func(tc endpointCase) {
			api := &models.StackResource{
				ID:        "res-api",
				Name:      "api",
				Namespace: "app",
				Ports: models.Ports{
					{Name: "http", Number: 8080, Protocol: models.PortProtocolHTTP, ExposedToPublic: true, ExposedFqdn: tc.fqdn},
					{Name: "local", Number: 9090, Protocol: models.PortProtocolHTTP, ExposedToPublic: true, ExposedFqdn: "api.local"},
				},
				ExecutionConfig: &models.ExecutionConfig{
					Env: []models.EnvVar{
						{Name: "PUBLIC_URL", SelfOutput: models.OutputNamePublicURL + ".http"},
						{Name: "INTERNAL_URL", SelfOutput: models.OutputNameURL + ".http"},
						{Name: "LOCAL_URL", SelfOutput: models.OutputNamePublicURL + ".local"},
					},
				},
			}
			stack := &models.Stack{
				ID:             "stack-1",
				StackResources: []*models.StackResource{api, {ID: "res-web", Name: "web", ExecutionConfig: &models.ExecutionConfig{}}},
				Connections: models.StackConnections{{
					ID:   "api-web",
					Kind: models.ConnectionKindEnv,
					From: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "api"},
					To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
					Mappings: []models.ConnectionMapping{
						{
							Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "API_URL"},
							Value:  models.ValueRef{Output: models.OutputNamePublicURL + ".http"},
						},
						{
							Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "API_CALLBACK"},
							Value: models.ValueRef{
								Template: "{{url}}/callback",
								Values:   map[string]models.OutputValueRef{"url": {Output: models.OutputNamePublicURL + ".http"}},
							},
						},
					},
				}},
			}

			resolver := NewResolver(ResolverSpec{PublicEndpoints: tc.config})
			effective, err := resolver.Resolve(context.Background(), stack)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(envValue(effective, "api", "PUBLIC_URL")).To(gomega.Equal(tc.want))
			gomega.Expect(envValue(effective, "api", "INTERNAL_URL")).To(gomega.Equal("http://api.app.svc:8080"))
			gomega.Expect(envValue(effective, "api", "LOCAL_URL")).To(gomega.Equal("http://api.local"))
			gomega.Expect(envValue(effective, "web", "API_URL")).To(gomega.Equal(tc.want))
			gomega.Expect(envValue(effective, "web", "API_CALLBACK")).To(gomega.Equal(tc.want + "/callback"))

			// The builder must configure the ingress with the same decision the
			// outputs advertise.
			builder := builders.NewClusterResourceBuilder(builders.ClusterResourceBuilderSpec{
				PublicEndpoints:    tc.config,
				PlatformBaseDomain: "app.stackdome.com",
			})
			cr, err := builder.BuildStackResourceCR(effective.StackResources[0], "stack", "org")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(cr.Spec.Ports[0].TLS).To(gomega.Equal(strings.HasPrefix(tc.want, "https://")))
			gomega.Expect(cr.Spec.Ports[1].TLS).To(gomega.BeFalse())
		},
		ginkgo.Entry("shared compute with platform TLS", endpointCase{"shared tls", models.PublicEndpointConfig{SharedCompute: true, PlatformTLSEnabled: true}, "api.app.stackdome.com", "https://api.app.stackdome.com"}),
		ginkgo.Entry("shared compute without platform TLS", endpointCase{"shared http", models.PublicEndpointConfig{SharedCompute: true}, "api.app.stackdome.com", "http://api.app.stackdome.com"}),
		ginkgo.Entry("BYOC real domain", endpointCase{"byoc tls", models.PublicEndpointConfig{}, "api.example.com", "https://api.example.com"}),
		ginkgo.Entry("BYOC local hostname", endpointCase{"byoc local", models.PublicEndpointConfig{}, "api.127-0-0-1.nip.io", "http://api.127-0-0-1.nip.io"}),
	)

	ginkgo.It("keeps single-port public_url scheme in self outputs", func() {
		api := &models.StackResource{
			ID:        "res-api",
			Name:      "api",
			Namespace: "app",
			Ports: models.Ports{
				{Name: "http", Number: 8080, Protocol: models.PortProtocolHTTP, ExposedToPublic: true, ExposedFqdn: "api.example.com"},
			},
			ExecutionConfig: &models.ExecutionConfig{
				Env: []models.EnvVar{{Name: "PUBLIC_URL", SelfOutput: models.OutputNamePublicURL}},
			},
		}
		stack := &models.Stack{
			ID:             "stack-1",
			StackResources: []*models.StackResource{api},
		}

		effective, err := NewResolver(ResolverSpec{}).Resolve(context.Background(), stack)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(envValue(effective, "api", "PUBLIC_URL")).To(gomega.Equal("https://api.example.com"))
	})
})
