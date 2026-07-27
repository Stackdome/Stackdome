package config

import (
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
			Expect((&ClusterConfig{}).AnySet()).To(BeFalse())
		})

		DescribeTable("is true when any single field is set",
			func(mutate func(*ClusterConfig)) {
				c := &ClusterConfig{}
				mutate(c)
				Expect(c.AnySet()).To(BeTrue())
			},
			Entry("only url", func(c *ClusterConfig) { c.ClusterURL = "x" }),
			Entry("only ca data", func(c *ClusterConfig) { c.ClusterCAData = "x" }),
			Entry("only token", func(c *ClusterConfig) { c.Token = "x" }),
		)
	})
})

var _ = Describe("ValidatePlatformProvisioning", func() {
	It("returns nil when nothing is configured", func() {
		Expect(ValidatePlatformProvisioning(&ClusterConfig{}, "", "")).To(Succeed())
	})

	It("rejects a partially-set cluster config", func() {
		partial := &ClusterConfig{
			ClusterURL: "https://cluster.example.com",
			Token:      "token",
		}
		Expect(ValidatePlatformProvisioning(partial, "example.com", "ops@example.com")).
			To(MatchError(ErrIncompleteClusterConfig))
	})

	It("rejects a full cluster with no base domain", func() {
		Expect(ValidatePlatformProvisioning(fullCluster(), "", "ops@example.com")).
			To(MatchError(ErrClusterDomainMismatch))
	})

	It("rejects a base domain with no cluster", func() {
		Expect(ValidatePlatformProvisioning(&ClusterConfig{}, "example.com", "")).
			To(MatchError(ErrClusterDomainMismatch))
	})

	It("requires a platform email when a cluster is configured", func() {
		Expect(ValidatePlatformProvisioning(fullCluster(), "example.com", "")).
			To(MatchError(ErrPlatformEmailRequired))
	})

	It("accepts a fully valid configuration", func() {
		Expect(ValidatePlatformProvisioning(fullCluster(), "example.com", "ops@example.com")).To(Succeed())
	})

	It("surfaces a cluster validation error", func() {
		invalid := fullCluster()
		invalid.ClusterURL = ""
		Expect(invalid.Validate()).To(MatchError("cluster url is required"))
	})
})
