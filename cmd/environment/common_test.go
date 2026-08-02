package environment

import (
	"github.com/Stackdome/stackdome/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LoadEnv", func() {
	DescribeTable("builds an environment for every supported STACKDOME_ENV",
		func(stackdomeEnv, expectedName string) {
			GinkgoT().Setenv(config.EnvStackdomeEnv.Name, stackdomeEnv)

			env := LoadEnv()

			Expect(env.Environment().Name).To(Equal(expectedName))
			Expect(env.Environment().Config).NotTo(BeNil())
			Expect(env.Environment().BootstrapConfig).NotTo(BeNil())
		},
		Entry("development", DEVELOPMENT_ENV, config.EnvironmentDevelopment),
		Entry("production", PRODUCTION_ENV, config.EnvironmentProduction),
	)

	It("defaults to development when STACKDOME_ENV is unset", func() {
		GinkgoT().Setenv(config.EnvStackdomeEnv.Name, "")

		Expect(LoadEnv().Environment().Name).To(Equal(config.EnvironmentDevelopment))
	})

	It("panics for an unknown environment", func() {
		GinkgoT().Setenv(config.EnvStackdomeEnv.Name, "STAGING")

		Expect(func() { LoadEnv() }).To(Panic())
	})

	It("does not let the test environment manage its own database", func() {
		testEnv := NewTestEnvironment(nil).(*environmentImpl)

		Expect(testEnv.Name).To(Equal(config.EnvironmentTest))
		Expect(testEnv.spec.managed).To(BeFalse())
		Expect(testEnv.loggerName("api-server")).To(Equal("test-api-server"))
	})
})
