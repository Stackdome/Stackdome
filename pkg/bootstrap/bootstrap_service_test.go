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
	domainSvc  *mocks.MockOrganisationDomainsService
	logger     *mocks.MockLogger
}

func newBootstrapDeps(ctrl *gomock.Controller) *bootstrapDeps {
	logger := mocks.NewMockLogger(ctrl)
	logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	return &bootstrapDeps{
		orgSvc:     mocks.NewMockOrganisationService(ctrl),
		clusterSvc: mocks.NewMockClusterService(ctrl),
		domainSvc:  mocks.NewMockOrganisationDomainsService(ctrl),
		logger:     logger,
	}
}

func (d *bootstrapDeps) service(bootstrapCfg *config.BootstrapConfig, clusterCfg *config.ClusterConfig) *bootstrap.Service {
	return bootstrap.NewService(bootstrap.Spec{
		OrganisationService:       d.orgSvc,
		ClusterService:            d.clusterSvc,
		OrganisationDomainService: d.domainSvc,
		BootstrapConfig:           bootstrapCfg,
		ClusterConfig:             clusterCfg,
		Logger:                    d.logger,
	})
}

func fullBootstrapConfig() *config.BootstrapConfig {
	return &config.BootstrapConfig{
		Email:                 contactEmail,
		BaseDomain:            baseDomain,
		DNSCloudflareAPIToken: dnsAPIToken,
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

	When("no platform cluster is configured", func() {
		It("no-ops without touching any service", func() {
			svc := deps.service(fullBootstrapConfig(), &config.ClusterConfig{})
			Expect(svc.Run(ctx)).To(Succeed())
		})
	})

	When("bootstrapping a fresh install", func() {
		It("creates the platform org, cluster, domain, and wildcard TLS", func() {
			deps.orgSvc.EXPECT().InternalGetPlatformOrg(gomock.Any()).
				Return(nil, errors.NotFound("platform organisation not found"))
			deps.orgSvc.EXPECT().InternalCreate(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, spec *models.Organisation) (*models.Organisation, *errors.ServiceError) {
					Expect(spec.Name).To(Equal(models.PlatformOrganisationName))
					Expect(spec.Platform).To(BeTrue())
					return &models.Organisation{ID: orgID, Name: models.PlatformOrganisationName, Platform: true}, nil
				})

			platformCluster := &models.Cluster{ID: clusterID}
			deps.clusterSvc.EXPECT().InternalUpsertPlatformCluster(gomock.Any(), gomock.Any()).
				DoAndReturn(func(callCtx context.Context, spec *models.Cluster) (*models.Cluster, *errors.ServiceError) {
					identity := auth.GetIdentityFromCtx(callCtx)
					Expect(identity.IsSystem).To(BeTrue())
					Expect(identity.ContactEmail).To(Equal(contactEmail))
					Expect(spec.Name).To(Equal(models.PlatformClusterName))
					Expect(spec.OrganisationID).To(Equal(orgID))
					Expect(spec.Platform).To(BeTrue())
					Expect(spec.ClusterURL).To(Equal(clusterURL))
					Expect(spec.ClusterCAData).To(Equal(clusterCA))
					Expect(spec.Token).To(Equal(clusterToken))
					return platformCluster, nil
				})

			deps.domainSvc.EXPECT().GetDefaultDomainForOrganisation(gomock.Any(), orgID).
				Return(nil, errors.NotFound("domain not found"))
			deps.domainSvc.EXPECT().Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, spec *models.OrganisationDomain) (*models.OrganisationDomain, *errors.ServiceError) {
					Expect(spec.OrganisationID).To(Equal(orgID))
					Expect(spec.Domain).To(Equal(baseDomain))
					return &models.OrganisationDomain{ID: "dom-1"}, nil
				})

			bootstrapCfg := fullBootstrapConfig()
			deps.clusterSvc.EXPECT().InternalEnsurePlatformWildcardTLS(gomock.Any(), platformCluster, bootstrapCfg).
				DoAndReturn(func(callCtx context.Context, cluster *models.Cluster, cfg *config.BootstrapConfig) *errors.ServiceError {
					identity := auth.GetIdentityFromCtx(callCtx)
					Expect(identity.IsSystem).To(BeTrue())
					Expect(identity.ContactEmail).To(Equal(contactEmail))
					Expect(cluster).To(BeIdenticalTo(platformCluster))
					Expect(cfg).To(BeIdenticalTo(bootstrapCfg))
					return nil
				})

			svc := deps.service(bootstrapCfg, setClusterConfig())
			Expect(svc.Run(ctx)).To(Succeed())
		})

		It("returns a contextual error when wildcard TLS provisioning fails", func() {
			deps.orgSvc.EXPECT().InternalGetPlatformOrg(gomock.Any()).
				Return(&models.Organisation{ID: orgID, Name: models.PlatformOrganisationName, Platform: true}, nil)

			platformCluster := &models.Cluster{ID: clusterID}
			deps.clusterSvc.EXPECT().InternalUpsertPlatformCluster(gomock.Any(), gomock.Any()).
				Return(platformCluster, nil)
			deps.domainSvc.EXPECT().GetDefaultDomainForOrganisation(gomock.Any(), orgID).
				Return(&models.OrganisationDomain{Domain: baseDomain}, nil)
			deps.clusterSvc.EXPECT().InternalEnsurePlatformWildcardTLS(gomock.Any(), platformCluster, gomock.Any()).
				Return(errors.GeneralError("certificate request failed"))

			svc := deps.service(fullBootstrapConfig(), setClusterConfig())
			Expect(svc.Run(ctx)).To(MatchError(ContainSubstring("failed to create or update platform wildcard TLS")))
		})
	})

	When("re-running against an already-provisioned install", func() {
		It("adopts the existing platform org and is idempotent", func() {
			deps.orgSvc.EXPECT().InternalGetPlatformOrg(gomock.Any()).
				Return(&models.Organisation{ID: orgID, Name: models.PlatformOrganisationName, Platform: true}, nil)

			deps.clusterSvc.EXPECT().InternalUpsertPlatformCluster(gomock.Any(), gomock.Any()).
				Return(&models.Cluster{ID: clusterID}, nil)

			deps.domainSvc.EXPECT().GetDefaultDomainForOrganisation(gomock.Any(), orgID).
				Return(&models.OrganisationDomain{Domain: baseDomain}, nil)
			bootstrapCfg := fullBootstrapConfig()
			deps.clusterSvc.EXPECT().InternalEnsurePlatformWildcardTLS(gomock.Any(), &models.Cluster{ID: clusterID}, bootstrapCfg).
				Return(nil)

			svc := deps.service(bootstrapCfg, setClusterConfig())
			Expect(svc.Run(ctx)).To(Succeed())
		})
	})

	When("the platform org lookup fails", func() {
		It("propagates the error without provisioning anything", func() {
			deps.orgSvc.EXPECT().InternalGetPlatformOrg(gomock.Any()).
				Return(nil, errors.GeneralError("db down"))

			svc := deps.service(fullBootstrapConfig(), setClusterConfig())
			Expect(svc.Run(ctx)).NotTo(Succeed())
		})
	})

})
