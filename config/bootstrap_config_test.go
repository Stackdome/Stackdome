package config

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func fullCluster() *ClusterConfig {
	return &ClusterConfig{
		ClusterURL:    "https://cluster.example.com",
		ClusterCAData: "Y2FkYXRh",
		Token:         "token",
	}
}

func emptyCluster() *ClusterConfig {
	return &ClusterConfig{}
}

func emptyRouting() *BootstrapConfig {
	return &BootstrapConfig{}
}

func routingWithoutTLS() *BootstrapConfig {
	return &BootstrapConfig{BaseDomain: "apps.example.com"}
}

func routingWithTLS() *BootstrapConfig {
	return &BootstrapConfig{
		BaseDomain:            "apps.example.com",
		PlatformTLSEnabled:    true,
		Email:                 "ops@example.com",
		DNSCloudflareAPIToken: "cloudflare-token",
		ACMEEnvironment:       ACMEEnvironmentProduction,
		TLSNamespace:          DefaultPlatformTLSNamespace,
	}
}

var _ = Describe("ClusterConfig set predicates", func() {
	Describe("IsSet", func() {
		It("is true only when all three fields are set", func() {
			Expect(fullCluster().IsSet()).To(BeTrue())
		})

		DescribeTable("is false when any field is missing",
			func(mutate func(*ClusterConfig)) {
				c := fullCluster()
				mutate(c)
				Expect(c.IsSet()).To(BeFalse())
			},
			Entry("no url", func(c *ClusterConfig) { c.ClusterURL = "" }),
			Entry("no ca data", func(c *ClusterConfig) { c.ClusterCAData = "" }),
			Entry("no token", func(c *ClusterConfig) { c.Token = "" }),
			Entry("all empty", func(c *ClusterConfig) { *c = ClusterConfig{} }),
		)
	})

	Describe("AnySet", func() {
		It("is false when all fields are empty", func() {
			Expect(emptyCluster().AnySet()).To(BeFalse())
		})

		DescribeTable("is true when any single field is set",
			func(mutate func(*ClusterConfig)) {
				c := emptyCluster()
				mutate(c)
				Expect(c.AnySet()).To(BeTrue())
			},
			Entry("only url", func(c *ClusterConfig) { c.ClusterURL = "x" }),
			Entry("only ca data", func(c *ClusterConfig) { c.ClusterCAData = "x" }),
			Entry("only token", func(c *ClusterConfig) { c.Token = "x" }),
		)
	})
})

var _ = Describe("shared compute provisioning and platform routing", func() {
	BeforeEach(func() {
		for _, name := range []string{
			EnvPlatformEmail.Name,
			EnvPlatformBaseDomain.Name,
			EnvPlatformDNSCloudflareAPIToken.Name,
			EnvPlatformTLSEnabled.Name,
			EnvPlatformACMEEnvironment.Name,
			EnvPlatformTLSNamespace.Name,
		} {
			Expect(os.Unsetenv(name)).To(Succeed())
		}
	})

	DescribeTable("validates independent shared compute and routing configuration",
		func(runtime RuntimeMode, mode ComputeMode, cluster *ClusterConfig, bootstrap *BootstrapConfig, expectedError error) {
			err := ValidateSharedComputeProvisioning(mode, cluster)
			if err == nil {
				err = ValidatePlatformRouting(runtime, mode, bootstrap)
			}
			if expectedError == nil {
				Expect(err).NotTo(HaveOccurred())
				return
			}
			Expect(err).To(MatchError(expectedError))
		},
		Entry("BYOC without shared config", RuntimeModeSelfHosted, ComputeModeBYOC, emptyCluster(), emptyRouting(), nil),
		Entry("shared without TLS", RuntimeModeSelfHosted, ComputeModeShared, fullCluster(), routingWithoutTLS(), nil),
		Entry("self-hosted shared with TLS", RuntimeModeSelfHosted, ComputeModeShared, fullCluster(), routingWithTLS(), nil),
		Entry("cloud shared without TLS", RuntimeModeStackdomeCloud, ComputeModeShared, fullCluster(), routingWithoutTLS(), ErrPlatformTLSRequired),
		Entry("cloud shared with TLS", RuntimeModeStackdomeCloud, ComputeModeShared, fullCluster(), routingWithTLS(), nil),
		Entry("BYOC with platform routing", RuntimeModeSelfHosted, ComputeModeBYOC, emptyCluster(), routingWithoutTLS(), ErrPlatformRoutingNotAllowed),
	)

	It("rejects an incomplete shared compute cluster", func() {
		Expect(ValidateSharedComputeProvisioning(ComputeModeShared, &ClusterConfig{ClusterURL: "https://cluster.example.com"})).
			To(MatchError(ErrIncompleteSharedComputeClusterConfig))
	})

	It("requires platform TLS configuration to be explicitly enabled", func() {
		bootstrap := routingWithoutTLS()
		bootstrap.Email = "ops@example.com"

		Expect(ValidatePlatformRouting(RuntimeModeSelfHosted, ComputeModeShared, bootstrap)).
			To(MatchError(ErrPlatformTLSConfigNotAllowed))
	})

	It("loads TLS defaults only when TLS is enabled", func() {
		GinkgoT().Setenv(EnvPlatformTLSEnabled.Name, "false")
		Expect(os.Unsetenv(EnvPlatformACMEEnvironment.Name)).To(Succeed())
		Expect(os.Unsetenv(EnvPlatformTLSNamespace.Name)).To(Succeed())

		disabled := NewBootstrapConfig()
		Expect(disabled.LoadEnvVariables()).To(Succeed())
		Expect(disabled.ACMEEnvironment).To(BeEmpty())
		Expect(disabled.TLSNamespace).To(BeEmpty())

		GinkgoT().Setenv(EnvPlatformTLSEnabled.Name, "true")
		enabled := NewBootstrapConfig()
		Expect(enabled.LoadEnvVariables()).To(Succeed())
		Expect(enabled.ACMEEnvironment).To(Equal(ACMEEnvironmentProduction))
		Expect(enabled.TLSNamespace).To(Equal(DefaultPlatformTLSNamespace))
	})

	It("preserves legacy TLS configuration when the enable flag is absent", func() {
		GinkgoT().Setenv(EnvPlatformEmail.Name, "ops@example.com")
		GinkgoT().Setenv(EnvPlatformBaseDomain.Name, "apps.example.com")
		GinkgoT().Setenv(EnvPlatformDNSCloudflareAPIToken.Name, "cloudflare-token")

		bootstrap := NewBootstrapConfig()
		Expect(bootstrap.LoadEnvVariables()).To(Succeed())

		Expect(bootstrap.PlatformTLSEnabled).To(BeTrue())
		Expect(bootstrap.ACMEEnvironment).To(Equal(ACMEEnvironmentProduction))
		Expect(bootstrap.TLSNamespace).To(Equal(DefaultPlatformTLSNamespace))
		Expect(ValidatePlatformRouting(RuntimeModeSelfHosted, ComputeModeShared, bootstrap)).To(Succeed())
	})

	It("keeps routing without TLS when no TLS configuration is present", func() {
		GinkgoT().Setenv(EnvPlatformBaseDomain.Name, "apps.example.com")

		bootstrap := NewBootstrapConfig()
		Expect(bootstrap.LoadEnvVariables()).To(Succeed())

		Expect(bootstrap.PlatformTLSEnabled).To(BeFalse())
		Expect(ValidatePlatformRouting(RuntimeModeSelfHosted, ComputeModeShared, bootstrap)).To(Succeed())
	})

	It("retains explicit TLS configuration when TLS is disabled", func() {
		GinkgoT().Setenv(EnvPlatformTLSEnabled.Name, "false")
		GinkgoT().Setenv(EnvPlatformACMEEnvironment.Name, ACMEEnvironmentStaging)
		GinkgoT().Setenv(EnvPlatformTLSNamespace.Name, "custom-tls")

		bootstrap := NewBootstrapConfig()
		Expect(bootstrap.LoadEnvVariables()).To(Succeed())
		Expect(bootstrap.ACMEEnvironment).To(Equal(ACMEEnvironmentStaging))
		Expect(bootstrap.TLSNamespace).To(Equal("custom-tls"))

		bootstrap.BaseDomain = "apps.example.com"
		Expect(ValidatePlatformRouting(RuntimeModeSelfHosted, ComputeModeShared, bootstrap)).
			To(MatchError(ErrPlatformTLSConfigNotAllowed))
	})

	It("surfaces a cluster validation error", func() {
		invalid := fullCluster()
		invalid.ClusterURL = ""
		Expect(invalid.Validate()).To(MatchError("cluster url is required"))
	})
})
