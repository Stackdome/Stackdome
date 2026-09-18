package stackdeploy

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resolver public URL outputs", func() {
	buildStack := func(multiPort bool, fqdn string) *models.Stack {
		api := &models.StackResource{
			Name:            "api",
			Namespace:       "app",
			ExecutionConfig: &models.ExecutionConfig{},
			Ports: models.Ports{
				{Name: "http", Number: 8080, Protocol: models.PortProtocolHTTP, ExposedToPublic: true, ExposedFqdn: fqdn},
			},
		}
		suffix := ""
		if multiPort {
			suffix = ".http"
			api.Ports = append(api.Ports, models.Port{
				Name: "metrics", Number: 9090, Protocol: models.PortProtocolHTTP,
				ExposedToPublic: true, ExposedFqdn: "api.local",
			})
		}
		api.ExecutionConfig.Env = append(api.ExecutionConfig.Env,
			models.EnvVar{Name: "PUBLIC_URL", SelfOutput: models.OutputNamePublicURL + suffix},
			models.EnvVar{Name: "INTERNAL_URL", SelfOutput: models.OutputNameURL + suffix},
		)
		if multiPort {
			api.ExecutionConfig.Env = append(api.ExecutionConfig.Env,
				models.EnvVar{Name: "LOCAL_URL", SelfOutput: models.OutputNamePublicURL + ".metrics"},
			)
		}

		stack := &models.Stack{
			StackResources: []*models.StackResource{api, {Name: "web", ExecutionConfig: &models.ExecutionConfig{}}},
			Connections: models.StackConnections{
				{
					ID:   "api-web",
					Kind: models.ConnectionKindEnv,
					From: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "api"},
					To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
					Mappings: []models.ConnectionMapping{
						{
							Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "API_URL"},
							Value:  models.ValueRef{Output: models.OutputNamePublicURL + suffix},
						},
						{
							Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "API_CALLBACK"},
							Value: models.ValueRef{
								Template: "{{url}}/callback",
								Values:   map[string]models.OutputValueRef{"url": {Output: models.OutputNamePublicURL + suffix}},
							},
						},
					},
				},
			},
		}
		return stack
	}

	DescribeTable("resolves self outputs and resource connections to the endpoint scheme",
		func(multiPort bool, config models.PublicEndpointConfig, fqdn, wantURL string) {
			stack := buildStack(multiPort, fqdn)
			resolver := NewResolver(ResolverSpec{PublicEndpoints: config})
			effective, err := resolver.Resolve(context.Background(), stack)

			Expect(err).NotTo(HaveOccurred())
			Expect(envValue(effective, "api", "PUBLIC_URL")).To(Equal(wantURL))
			Expect(envValue(effective, "web", "API_URL")).To(Equal(wantURL))
			Expect(envValue(effective, "web", "API_CALLBACK")).To(Equal(wantURL + "/callback"))
			Expect(envValue(effective, "api", "INTERNAL_URL")).To(Equal("http://api.app.svc:8080"))
			if multiPort {
				Expect(envValue(effective, "api", "LOCAL_URL")).To(Equal("http://api.local"))
			}
		},
		Entry("HTTPS single port", false, models.PublicEndpointConfig{}, "api.example.com", "https://api.example.com"),
		Entry("HTTP single port", false, models.PublicEndpointConfig{}, "api.127-0-0-1.nip.io", "http://api.127-0-0-1.nip.io"),
		Entry("HTTPS multi port", true, models.PublicEndpointConfig{}, "api.example.com", "https://api.example.com"),
		Entry("HTTP multi port", true, models.PublicEndpointConfig{}, "api.127-0-0-1.nip.io", "http://api.127-0-0-1.nip.io"),
		Entry("shared compute without platform TLS", false, models.PublicEndpointConfig{SharedCompute: true}, "api.app.stackdome.com", "http://api.app.stackdome.com"),
		Entry("shared compute with platform TLS", false, models.PublicEndpointConfig{SharedCompute: true, PlatformTLSEnabled: true}, "api.app.stackdome.com", "https://api.app.stackdome.com"),
	)
})
