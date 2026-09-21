package stackdeploy

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Public URL output scheme", func() {
	httpsScheme := func(models.Port) string { return models.OutputURLSchemeHTTPS }

	It("resolves a self-output public_url with the public endpoint scheme", func() {
		stack := &models.Stack{
			ID: "stack-1",
			StackResources: []*models.StackResource{
				{
					ID:   "api-id",
					Name: "api",
					ExecutionConfig: &models.ExecutionConfig{
						Env: []models.EnvVar{{Name: "PUBLIC_URL", SelfOutput: models.OutputNamePublicURL}},
					},
					Ports: models.Ports{
						{Name: "http", Number: 8080, Protocol: models.PortProtocolHTTP, ExposedToPublic: true, ExposedFqdn: "api.example.com"},
					},
				},
			},
		}

		effective, err := NewResolver(ResolverSpec{PublicURLScheme: httpsScheme}).Resolve(context.Background(), stack)

		Expect(err).NotTo(HaveOccurred())
		Expect(envValue(effective, "api", "PUBLIC_URL")).To(Equal("https://api.example.com"))
	})

	It("resolves resource-output connections with the public endpoint scheme and keeps internal URLs on http", func() {
		stack := &models.Stack{
			ID: "stack-1",
			StackResources: []*models.StackResource{
				{
					ID:        "res-api",
					Name:      "api",
					Namespace: "default",
					Ports: models.Ports{
						{Name: "http", Number: 8080, Protocol: models.PortProtocolHTTP, ExposedToPublic: true, ExposedFqdn: "api.example.com"},
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
							Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "API_URL"},
							Value:  models.ValueRef{Output: models.OutputNameURL},
						},
						{
							Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "API_PUBLIC_URL"},
							Value:  models.ValueRef{Output: models.OutputNamePublicURL},
						},
						{
							Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "API_TEMPLATE_URL"},
							Value: models.ValueRef{
								Template: "{{public_url}}/health",
								Values:   map[string]models.OutputValueRef{"public_url": {Output: models.OutputNamePublicURL}},
							},
						},
					},
				},
			},
		}

		effective, err := NewResolver(ResolverSpec{PublicURLScheme: httpsScheme}).Resolve(context.Background(), stack)

		Expect(err).NotTo(HaveOccurred())
		Expect(envValue(effective, "web", "API_URL")).To(Equal("http://api.default.svc:8080"))
		Expect(envValue(effective, "web", "API_PUBLIC_URL")).To(Equal("https://api.example.com"))
		Expect(envValue(effective, "web", "API_TEMPLATE_URL")).To(Equal("https://api.example.com/health"))
	})

	It("defaults to http when no public scheme is configured", func() {
		stack := &models.Stack{
			ID: "stack-1",
			StackResources: []*models.StackResource{
				{
					ID:   "api-id",
					Name: "api",
					ExecutionConfig: &models.ExecutionConfig{
						Env: []models.EnvVar{{Name: "PUBLIC_URL", SelfOutput: models.OutputNamePublicURL}},
					},
					Ports: models.Ports{
						{Name: "http", Number: 8080, Protocol: models.PortProtocolHTTP, ExposedToPublic: true, ExposedFqdn: "api.example.com"},
					},
				},
			},
		}

		effective, err := NewResolver(ResolverSpec{}).Resolve(context.Background(), stack)

		Expect(err).NotTo(HaveOccurred())
		Expect(envValue(effective, "api", "PUBLIC_URL")).To(Equal("http://api.example.com"))
	})
})
