package environment

import (
	"errors"

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
		Expect(testEnv.loggerName("api-server")).To(Equal("test-api-server"))
	})

	It("declares whether each environment creates or receives dependencies", func() {
		Expect(developmentSpec.dependencySource).To(Equal(dependenciesCreated))
		Expect(productionSpec.dependencySource).To(Equal(dependenciesCreated))
		Expect(testSpec.dependencySource).To(Equal(dependenciesInjected))
	})

	DescribeTable("identifies dependency ownership",
		func(source dependencySource, creates bool) {
			Expect(source.createsDependencies()).To(Equal(creates))
		},
		Entry("created dependencies", dependenciesCreated, true),
		Entry("injected dependencies", dependenciesInjected, false),
	)

	It("panics for an unset dependency source", func() {
		Expect(func() { dependencySource("").createsDependencies() }).To(PanicWith("unsupported dependency source: \"\""))
	})

	It("requires platform provisioning in Stackdome Cloud mode", func() {
		applicationConfig := config.NewApplicationConfig()
		applicationConfig.RuntimeMode = config.RuntimeModeStackdomeCloud
		testEnv := NewTestEnvironment(nil, WithApplicationConfig(applicationConfig)).(*environmentImpl)

		Expect(testEnv.validatePlatformProvisioning()).To(MatchError(config.ErrPlatformProvisioningRequired))
	})

	It("keeps platform provisioning optional in self-hosted mode", func() {
		applicationConfig := config.NewApplicationConfig()
		applicationConfig.RuntimeMode = config.RuntimeModeSelfHosted
		testEnv := NewTestEnvironment(nil, WithApplicationConfig(applicationConfig)).(*environmentImpl)

		Expect(testEnv.validatePlatformProvisioning()).To(Succeed())
	})

	It("stops loading configuration after the first error", func() {
		loadErr := errors.New("invalid cloud configuration")
		var loaded []string

		err := runConfigLoaders([]configLoader{
			func() error {
				loaded = append(loaded, "application environment")
				return nil
			},
			func() error {
				loaded = append(loaded, "bootstrap environment")
				return loadErr
			},
			func() error {
				loaded = append(loaded, "cloud configuration")
				return nil
			},
		})

		Expect(err).To(MatchError(ContainSubstring("load configuration: invalid cloud configuration")))
		Expect(loaded).To(Equal([]string{"application environment", "bootstrap environment"}))
	})
})
