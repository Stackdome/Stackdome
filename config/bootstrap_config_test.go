package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func fullCluster() *ClusterConfig {
	return &ClusterConfig{
		Name:          "default",
		ClusterURL:    "https://cluster.example.com",
		ClusterCAData: "Y2FkYXRh",
		Token:         "token",
	}
}

var _ = Describe("ClusterConfig set predicates", func() {
	Describe("IsSet", func() {
		It("is true only when all four fields are set", func() {
			Expect(fullCluster().IsSet()).To(BeTrue())
		})

		DescribeTable("is false when any field is missing",
			func(mutate func(*ClusterConfig)) {
				c := fullCluster()
				mutate(c)
				Expect(c.IsSet()).To(BeFalse())
			},
			Entry("no name", func(c *ClusterConfig) { c.Name = "" }),
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
			Entry("only name", func(c *ClusterConfig) { c.Name = "x" }),
			Entry("only url", func(c *ClusterConfig) { c.ClusterURL = "x" }),
			Entry("only ca data", func(c *ClusterConfig) { c.ClusterCAData = "x" }),
			Entry("only token", func(c *ClusterConfig) { c.Token = "x" }),
		)
	})
})

var _ = Describe("ValidateDefaultProvisioning", func() {
	It("returns nil when nothing is configured", func() {
		Expect(ValidateDefaultProvisioning(&ClusterConfig{}, "", "")).To(Succeed())
	})

	It("rejects a partially-set cluster config", func() {
		partial := &ClusterConfig{
			ClusterURL: "https://cluster.example.com",
			Token:      "token",
		}
		Expect(ValidateDefaultProvisioning(partial, "example.com", "admin@example.com")).
			To(MatchError(ErrIncompleteClusterConfig))
	})

	It("rejects a full cluster with no base domain", func() {
		Expect(ValidateDefaultProvisioning(fullCluster(), "", "admin@example.com")).
			To(MatchError(ErrClusterDomainMismatch))
	})

	It("rejects a base domain with no cluster", func() {
		Expect(ValidateDefaultProvisioning(&ClusterConfig{}, "example.com", "")).
			To(MatchError(ErrClusterDomainMismatch))
	})

	It("requires an admin email when a cluster is configured", func() {
		Expect(ValidateDefaultProvisioning(fullCluster(), "example.com", "")).
			To(MatchError(ErrBootstrapAdminEmail))
	})

	It("accepts a fully valid configuration", func() {
		Expect(ValidateDefaultProvisioning(fullCluster(), "example.com", "admin@example.com")).To(Succeed())
	})

	It("surfaces a cluster validation error", func() {
		invalid := fullCluster()
		invalid.ClusterURL = ""
		Expect(invalid.Validate()).To(MatchError("cluster url is required"))
	})
})
