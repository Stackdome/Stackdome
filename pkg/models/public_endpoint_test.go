package models

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("PublicEndpointConfig.UsesTLS", func() {
	usesTLS := func(config PublicEndpointConfig, port Port) bool {
		return config.UsesTLS(port)
	}

	ginkgo.DescribeTable("applies the hostname policy",
		func(fqdn string, want bool) {
			port := Port{ExposedToPublic: true, ExposedFqdn: fqdn}
			gomega.Expect(usesTLS(PublicEndpointConfig{}, port)).To(gomega.Equal(want))
		},
		ginkgo.Entry("empty hostname", "", false),
		ginkgo.Entry("nip.io subdomain", "app.192-168-1-1.nip.io", false),
		ginkgo.Entry("sslip.io subdomain", "app.10-0-0-1.sslip.io", false),
		ginkgo.Entry("dot local", "myapp.local", false),
		ginkgo.Entry("dot localhost", "myapp.localhost", false),
		ginkgo.Entry("real domain", "app.example.com", true),
		ginkgo.Entry("subdomain", "api.staging.example.com", true),
		ginkgo.Entry("bare domain", "example.com", true),
		ginkgo.Entry("io TLD not matching nip.io", "myapp.io", true),
		ginkgo.Entry("domain ending in local but not .local suffix", "app.mylocal", true),
		ginkgo.Entry("case-insensitive local suffix", "App.LOCAL", false),
	)

	ginkgo.It("is false for a port that is not exposed publicly", func() {
		port := Port{ExposedToPublic: false, ExposedFqdn: "app.example.com"}
		gomega.Expect(usesTLS(PublicEndpointConfig{}, port)).To(gomega.BeFalse())
	})

	ginkgo.It("is false on shared compute when platform TLS is disabled", func() {
		port := Port{ExposedToPublic: true, ExposedFqdn: "app.example.com"}
		gomega.Expect(usesTLS(PublicEndpointConfig{SharedCompute: true}, port)).To(gomega.BeFalse())
	})

	ginkgo.It("is true on shared compute when platform TLS is enabled", func() {
		port := Port{ExposedToPublic: true, ExposedFqdn: "app.example.com"}
		gomega.Expect(usesTLS(PublicEndpointConfig{SharedCompute: true, PlatformTLSEnabled: true}, port)).To(gomega.BeTrue())
	})
})
