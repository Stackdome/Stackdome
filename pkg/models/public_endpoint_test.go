package models

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("PublicEndpointConfig.UsesTLS", func() {
	port := func(fqdn string, public bool) Port {
		return Port{ExposedToPublic: public, ExposedFqdn: fqdn}
	}

	ginkgo.DescribeTable("resolves the ingress TLS decision",
		func(config PublicEndpointConfig, p Port, want bool) {
			gomega.Expect(config.UsesTLS(p)).To(gomega.Equal(want))
		},
		ginkgo.Entry("BYOC real domain", PublicEndpointConfig{}, port("api.example.com", true), true),
		ginkgo.Entry("BYOC nip.io", PublicEndpointConfig{}, port("api.127-0-0-1.nip.io", true), false),
		ginkgo.Entry("BYOC sslip.io", PublicEndpointConfig{}, port("api.127-0-0-1.sslip.io", true), false),
		ginkgo.Entry("BYOC .local", PublicEndpointConfig{}, port("api.local", true), false),
		ginkgo.Entry("BYOC .localhost", PublicEndpointConfig{}, port("api.localhost", true), false),
		ginkgo.Entry("shared compute without platform TLS", PublicEndpointConfig{SharedCompute: true}, port("api.example.com", true), false),
		ginkgo.Entry("shared compute with platform TLS", PublicEndpointConfig{SharedCompute: true, PlatformTLSEnabled: true}, port("api.example.com", true), true),
		ginkgo.Entry("private port", PublicEndpointConfig{}, port("api.example.com", false), false),
		ginkgo.Entry("unassigned hostname", PublicEndpointConfig{}, port("", true), false),
	)
})
