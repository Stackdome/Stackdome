package models

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("PublicEndpointConfig.UsesTLS", func() {
	ginkgo.DescribeTable("decides TLS from the exposed hostname",
		func(fqdn string, want bool) {
			port := Port{ExposedToPublic: true, ExposedFqdn: fqdn}
			gomega.Expect((PublicEndpointConfig{}).UsesTLS(port)).To(gomega.Equal(want))
		},
		ginkgo.Entry("empty string", "", false),
		ginkgo.Entry("nip.io subdomain", "app.192-168-1-1.nip.io", false),
		ginkgo.Entry("sslip.io subdomain", "app.10-0-0-1.sslip.io", false),
		ginkgo.Entry("dot local", "myapp.local", false),
		ginkgo.Entry("dot localhost", "myapp.localhost", false),
		ginkgo.Entry("real domain", "app.example.com", true),
		ginkgo.Entry("subdomain", "api.staging.example.com", true),
		ginkgo.Entry("bare domain", "example.com", true),
		ginkgo.Entry("io TLD not matching nip.io", "myapp.io", true),
		ginkgo.Entry("domain ending in local but not .local suffix", "app.mylocal", true),
	)

	ginkgo.It("keeps TLS disabled for shared compute without platform TLS", func() {
		port := Port{ExposedToPublic: true, ExposedFqdn: "api.app.stackdome.com"}
		gomega.Expect(PublicEndpointConfig{SharedCompute: true, PlatformTLSEnabled: true}.UsesTLS(port)).To(gomega.BeTrue())
		gomega.Expect(PublicEndpointConfig{SharedCompute: true}.UsesTLS(port)).To(gomega.BeFalse())
	})

	ginkgo.It("requires a public port with an assigned hostname", func() {
		cfg := PublicEndpointConfig{SharedCompute: true, PlatformTLSEnabled: true}
		gomega.Expect(cfg.UsesTLS(Port{ExposedToPublic: false, ExposedFqdn: "api.example.com"})).To(gomega.BeFalse())
		gomega.Expect(cfg.UsesTLS(Port{ExposedToPublic: true, ExposedFqdn: ""})).To(gomega.BeFalse())
	})
})
