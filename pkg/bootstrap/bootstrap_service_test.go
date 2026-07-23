package bootstrap_test

import (
	"context"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/bootstrap"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

const (
	baseDomain   = "apps.example.com"
	storageSize  = "50Gi"
	storageClass = "standard"
	clusterName  = "default"
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
		BaseDomain:           baseDomain,
		RegistryStorageSize:  storageSize,
		RegistryStorageClass: storageClass,
	}
}

func setClusterConfig() *config.ClusterConfig {
	return &config.ClusterConfig{
		Name:          clusterName,
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
		It("creates the platform org and provisions cluster and domain", func() {
			deps.orgSvc.EXPECT().InternalGetPlatformOrg(gomock.Any()).
				Return(nil, errors.NotFound("platform organisation not found"))
			deps.orgSvc.EXPECT().InternalCreate(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, spec *models.Organisation) (*models.Organisation, *errors.ServiceError) {
					Expect(spec.Name).To(Equal(models.PlatformOrganisationName))
					Expect(spec.Platform).To(BeTrue())
					return &models.Organisation{ID: orgID, Name: models.PlatformOrganisationName, Platform: true}, nil
				})

			deps.clusterSvc.EXPECT().InternalUpsertPlatformCluster(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, spec *models.Cluster) (*models.Cluster, *errors.ServiceError) {
					Expect(spec.Name).To(Equal(clusterName))
					Expect(spec.OrganisationID).To(Equal(orgID))
					Expect(spec.Platform).To(BeTrue())
					Expect(spec.ClusterURL).To(Equal(clusterURL))
					Expect(spec.ClusterCAData).To(Equal(clusterCA))
					Expect(spec.Token).To(Equal(clusterToken))
					return &models.Cluster{ID: clusterID}, nil
				})

			deps.domainSvc.EXPECT().GetDefaultDomainForOrganisation(gomock.Any(), orgID).
				Return(nil, errors.NotFound("domain not found"))
			deps.domainSvc.EXPECT().Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, spec *models.OrganisationDomain) (*models.OrganisationDomain, *errors.ServiceError) {
					Expect(spec.OrganisationID).To(Equal(orgID))
					Expect(spec.Domain).To(Equal(baseDomain))
					return &models.OrganisationDomain{ID: "dom-1"}, nil
				})

			svc := deps.service(fullBootstrapConfig(), setClusterConfig())
			Expect(svc.Run(ctx)).To(Succeed())
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

			svc := deps.service(fullBootstrapConfig(), setClusterConfig())
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
