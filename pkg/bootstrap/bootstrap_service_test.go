package bootstrap_test

import (
	"context"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/bootstrap"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

const (
	contactEmail = "ops@example.com"
	baseDomain   = "apps.example.com"
	dnsAPIToken  = "cloudflare-api-token"
	tlsNamespace = "stackdome-control-plane"
	storageSize  = "50Gi"
	storageClass = "standard"
	clusterURL   = "https://cluster.example.com"
	clusterCA    = "ca-data"
	clusterToken = "token-data"
	orgID        = "org-123"
	clusterID    = "cluster-123"
)

type bootstrapDeps struct {
	orgSvc     *mocks.MockOrganisationService
	clusterSvc *mocks.MockClusterService
	logger     *mocks.MockLogger
}

func newBootstrapDeps(ctrl *gomock.Controller) *bootstrapDeps {
	logger := mocks.NewMockLogger(ctrl)
	logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	return &bootstrapDeps{
		orgSvc:     mocks.NewMockOrganisationService(ctrl),
		clusterSvc: mocks.NewMockClusterService(ctrl),
		logger:     logger,
	}
}

func (d *bootstrapDeps) service(platformCfg *config.PlatformConfig, clusterCfg *config.ClusterConfig) *bootstrap.Service {
	return bootstrap.NewService(bootstrap.Spec{
		OrganisationService: d.orgSvc,
		ClusterService:      d.clusterSvc,
		PlatformConfig:      platformCfg,
		ClusterConfig:       clusterCfg,
		Logger:              d.logger,
	})
}

func fullPlatformConfig() *config.PlatformConfig {
	return &config.PlatformConfig{
		Email:                 contactEmail,
		BaseDomain:            baseDomain,
		DNSCloudflareAPIToken: dnsAPIToken,
		PlatformTLSEnabled:    true,
		ACMEEnvironment:       config.ACMEEnvironmentStaging,
		TLSNamespace:          tlsNamespace,
		OrgRegistry:           models.OrgRegistryDefaults{StorageSize: storageSize, StorageClass: storageClass},
	}
}

func setClusterConfig() *config.ClusterConfig {
	return &config.ClusterConfig{
		ClusterURL:    clusterURL,
		ClusterCAData: clusterCA,
		Token:         clusterToken,
	}
}

var _ = Describe("Bootstrap", func() {
	var (
		ctrl *gomock.Controller
		deps *bootstrapDeps
		ctx  context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		deps = newBootstrapDeps(ctrl)
		ctx = context.Background()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	When("no shared-compute cluster is configured", func() {
		It("no-ops without touching any service", func() {
			svc := deps.service(fullPlatformConfig(), &config.ClusterConfig{})
			Expect(svc.Run(ctx)).To(Succeed())
		})
	})

	When("bootstrapping a fresh install", func() {
		It("creates the platform org, shared-compute cluster, and platform wildcard TLS", func() {
			deps.orgSvc.EXPECT().InternalGetPlatformOrg(gomock.Any()).
				Return(nil, errors.NotFound("platform organisation not found"))
			deps.orgSvc.EXPECT().InternalCreate(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, spec *models.Organisation) (*models.Organisation, *errors.ServiceError) {
					Expect(spec.Name).To(Equal(models.PlatformOrganisationName))
					Expect(spec.Platform).To(BeTrue())
					return &models.Organisation{ID: orgID, Name: models.PlatformOrganisationName, Platform: true}, nil
				})

			sharedComputeCluster := &models.Cluster{ID: clusterID}
			deps.clusterSvc.EXPECT().InternalUpsertSharedComputeCluster(gomock.Any(), gomock.Any()).
				DoAndReturn(func(callCtx context.Context, spec *models.Cluster) (*models.Cluster, *errors.ServiceError) {
					identity := auth.GetIdentityFromCtx(callCtx)
					Expect(identity.IsSystem).To(BeTrue())
					Expect(identity.ContactEmail).To(Equal(contactEmail))
					Expect(spec.Name).To(Equal(models.SharedComputeClusterName))
					Expect(spec.OrganisationID).To(Equal(orgID))
					Expect(spec.SharedCompute).To(BeTrue())
					Expect(spec.ClusterURL).To(Equal(clusterURL))
					Expect(spec.ClusterCAData).To(Equal(clusterCA))
					Expect(spec.Token).To(Equal(clusterToken))
					return sharedComputeCluster, nil
				})

			platformCfg := fullPlatformConfig()
			deps.clusterSvc.EXPECT().InternalEnsurePlatformWildcardTLS(gomock.Any(), sharedComputeCluster, platformCfg).
				DoAndReturn(func(callCtx context.Context, cluster *models.Cluster, cfg *config.PlatformConfig) *errors.ServiceError {
					identity := auth.GetIdentityFromCtx(callCtx)
					Expect(identity.IsSystem).To(BeTrue())
					Expect(identity.ContactEmail).To(Equal(contactEmail))
					Expect(cluster).To(BeIdenticalTo(sharedComputeCluster))
					Expect(cfg).To(BeIdenticalTo(platformCfg))
					return nil
				})

			svc := deps.service(platformCfg, setClusterConfig())
			Expect(svc.Run(ctx)).To(Succeed())
		})

		It("creates the platform org and shared-compute cluster without platform TLS when disabled", func() {
			deps.orgSvc.EXPECT().InternalGetPlatformOrg(gomock.Any()).
				Return(nil, errors.NotFound("platform organisation not found"))
			deps.orgSvc.EXPECT().InternalCreate(gomock.Any(), gomock.Any()).
				Return(&models.Organisation{ID: orgID, Name: models.PlatformOrganisationName, Platform: true}, nil)

			deps.clusterSvc.EXPECT().InternalUpsertSharedComputeCluster(gomock.Any(), gomock.Any()).
				DoAndReturn(func(callCtx context.Context, spec *models.Cluster) (*models.Cluster, *errors.ServiceError) {
					identity := auth.GetIdentityFromCtx(callCtx)
					Expect(identity.IsSystem).To(BeTrue())
					Expect(identity.ContactEmail).To(BeEmpty())
					Expect(spec.SharedCompute).To(BeTrue())
					return &models.Cluster{ID: clusterID, SharedCompute: true}, nil
				})

			platformCfg := fullPlatformConfig()
			platformCfg.PlatformTLSEnabled = false
			svc := deps.service(platformCfg, setClusterConfig())
			Expect(svc.Run(ctx)).To(Succeed())
		})

		It("returns a contextual error when wildcard TLS provisioning fails", func() {
			deps.orgSvc.EXPECT().InternalGetPlatformOrg(gomock.Any()).
				Return(&models.Organisation{ID: orgID, Name: models.PlatformOrganisationName, Platform: true}, nil)

			sharedComputeCluster := &models.Cluster{ID: clusterID}
			deps.clusterSvc.EXPECT().InternalUpsertSharedComputeCluster(gomock.Any(), gomock.Any()).
				Return(sharedComputeCluster, nil)
			deps.clusterSvc.EXPECT().InternalEnsurePlatformWildcardTLS(gomock.Any(), sharedComputeCluster, gomock.Any()).
				Return(errors.GeneralError("certificate request failed"))

			svc := deps.service(fullPlatformConfig(), setClusterConfig())
			Expect(svc.Run(ctx)).To(MatchError(ContainSubstring("failed to create or update platform wildcard TLS")))
		})
	})

	When("re-running against an already-provisioned install", func() {
		It("adopts the existing platform org and is idempotent", func() {
			deps.orgSvc.EXPECT().InternalGetPlatformOrg(gomock.Any()).
				Return(&models.Organisation{ID: orgID, Name: models.PlatformOrganisationName, Platform: true}, nil)

			deps.clusterSvc.EXPECT().InternalUpsertSharedComputeCluster(gomock.Any(), gomock.Any()).
				Return(&models.Cluster{ID: clusterID}, nil)

			platformCfg := fullPlatformConfig()
			deps.clusterSvc.EXPECT().InternalEnsurePlatformWildcardTLS(gomock.Any(), &models.Cluster{ID: clusterID}, platformCfg).
				Return(nil)

			svc := deps.service(platformCfg, setClusterConfig())
			Expect(svc.Run(ctx)).To(Succeed())
		})
	})

	When("the platform org lookup fails", func() {
		It("propagates the error without provisioning anything", func() {
			deps.orgSvc.EXPECT().InternalGetPlatformOrg(gomock.Any()).
				Return(nil, errors.GeneralError("db down"))

			svc := deps.service(fullPlatformConfig(), setClusterConfig())
			Expect(svc.Run(ctx)).NotTo(Succeed())
		})
	})

})
