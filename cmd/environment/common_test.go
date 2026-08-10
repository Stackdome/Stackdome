package environment

import (
	"context"
	"errors"

	"github.com/Stackdome/stackdome/config"
	serviceerrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
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

	It("requires shared compute provisioning for shared compute", func() {
		applicationConfig := config.NewApplicationConfig()
		applicationConfig.ComputeMode = config.ComputeModeShared
		testEnv := NewTestEnvironment(nil, WithApplicationConfig(applicationConfig)).(*environmentImpl)

		Expect(testEnv.validateSharedComputeProvisioning()).To(MatchError(config.ErrSharedComputeProvisioningRequired))
	})

	It("keeps shared compute provisioning disabled for BYOC", func() {
		applicationConfig := config.NewApplicationConfig()
		applicationConfig.ComputeMode = config.ComputeModeBYOC
		testEnv := NewTestEnvironment(nil, WithApplicationConfig(applicationConfig)).(*environmentImpl)

		Expect(testEnv.validateSharedComputeProvisioning()).To(Succeed())
	})

	Describe("persisted compute topology", func() {
		var (
			ctx          context.Context
			clusterStore *mocks.MockClusterStore
		)

		BeforeEach(func() {
			ctx = context.Background()
			clusterStore = mocks.NewMockClusterStore(gomock.NewController(GinkgoT()))
		})

		It("allows bring-your-own compute when no shared-compute cluster exists", func() {
			clusterStore.EXPECT().ListAll(ctx).Return(nil, nil)

			Expect(checkPersistedComputeTopology(ctx, config.ComputeModeBYOC, clusterStore)).To(Succeed())
		})

		It("allows bring-your-own compute with tenant-owned clusters", func() {
			clusterStore.EXPECT().ListAll(ctx).Return([]*models.Cluster{
				{ID: "tenant-cluster", Platform: false},
			}, nil)

			Expect(checkPersistedComputeTopology(ctx, config.ComputeModeBYOC, clusterStore)).To(Succeed())
		})

		It("rejects bring-your-own compute when a shared-compute cluster exists", func() {
			clusterStore.EXPECT().ListAll(ctx).Return([]*models.Cluster{
				{ID: "platform-cluster", Platform: true},
			}, nil)

			err := checkPersistedComputeTopology(ctx, config.ComputeModeBYOC, clusterStore)

			Expect(err).To(MatchError(
				"bring-your-own compute cannot start while shared-compute cluster \"platform-cluster\" exists; " +
					"set COMPUTE_MODE=shared or remove the shared-compute cluster and dependent resources",
			))
		})

		It("allows shared compute before its shared-compute cluster is bootstrapped", func() {
			clusterStore.EXPECT().ListAll(ctx).Return(nil, nil)

			Expect(checkPersistedComputeTopology(ctx, config.ComputeModeShared, clusterStore)).To(Succeed())
		})

		It("allows shared compute with a shared-compute cluster", func() {
			clusterStore.EXPECT().ListAll(ctx).Return([]*models.Cluster{
				{ID: "platform-cluster", Platform: true},
			}, nil)

			Expect(checkPersistedComputeTopology(ctx, config.ComputeModeShared, clusterStore)).To(Succeed())
		})

		It("rejects shared compute when a tenant-owned cluster exists", func() {
			clusterStore.EXPECT().ListAll(ctx).Return([]*models.Cluster{
				{ID: "tenant-cluster", Platform: false},
			}, nil)

			err := checkPersistedComputeTopology(ctx, config.ComputeModeShared, clusterStore)

			Expect(err).To(MatchError(
				"shared compute cannot start while tenant-owned cluster \"tenant-cluster\" exists; " +
					"set COMPUTE_MODE=bring_your_own or remove the tenant-owned cluster and dependent resources",
			))
		})

		DescribeTable("returns cluster lookup failures",
			func(mode config.ComputeMode) {
				clusterStore.EXPECT().ListAll(ctx).
					Return(nil, serviceerrors.GeneralError("database unavailable"))

				err := checkPersistedComputeTopology(ctx, mode, clusterStore)

				Expect(err).To(MatchError("list persisted clusters: error: database unavailable"))
			},
			Entry("in bring-your-own mode", config.ComputeModeBYOC),
			Entry("in shared mode", config.ComputeModeShared),
		)
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
