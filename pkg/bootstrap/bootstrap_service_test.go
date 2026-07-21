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
	"golang.org/x/crypto/bcrypt"
)

const (
	adminEmail    = "admin@platform.test"
	adminName     = "Platform Admin"
	adminPassword = "sup3r-s3cret"
	orgName       = "Platform"
	baseDomain    = "apps.example.com"
	registryName  = "platform-registry"
	storageSize   = "50Gi"
	storageClass  = "standard"
	clusterName   = "default"
	clusterURL    = "https://cluster.example.com"
	clusterCA     = "ca-data"
	clusterToken  = "token-data"
	orgID         = "org-123"
	userID        = "user-123"
	clusterID     = "cluster-123"
)

type bootstrapDeps struct {
	userSvc     *mocks.MockUserService
	orgSvc      *mocks.MockOrganisationService
	projSvc     *mocks.MockProjectService
	clusterSvc  *mocks.MockClusterService
	registrySvc *mocks.MockImageRegistryService
	domainSvc   *mocks.MockOrganisationDomainsService
	policyMgr   *mocks.MockResourceAccessPolicyManager
	logger      *mocks.MockLogger
}

func newBootstrapDeps(ctrl *gomock.Controller) *bootstrapDeps {
	logger := mocks.NewMockLogger(ctrl)
	logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	return &bootstrapDeps{
		userSvc:     mocks.NewMockUserService(ctrl),
		orgSvc:      mocks.NewMockOrganisationService(ctrl),
		projSvc:     mocks.NewMockProjectService(ctrl),
		clusterSvc:  mocks.NewMockClusterService(ctrl),
		registrySvc: mocks.NewMockImageRegistryService(ctrl),
		domainSvc:   mocks.NewMockOrganisationDomainsService(ctrl),
		policyMgr:   mocks.NewMockResourceAccessPolicyManager(ctrl),
		logger:      logger,
	}
}

func (d *bootstrapDeps) service(bootstrapCfg *config.BootstrapConfig, clusterCfg *config.ClusterConfig) *bootstrap.Service {
	return bootstrap.NewService(bootstrap.Spec{
		UserService:               d.userSvc,
		OrganisationService:       d.orgSvc,
		ProjectService:            d.projSvc,
		ClusterService:            d.clusterSvc,
		ImageRegistryService:      d.registrySvc,
		OrganisationDomainService: d.domainSvc,
		PolicyManager:             d.policyMgr,
		BootstrapConfig:           bootstrapCfg,
		ClusterConfig:             clusterCfg,
		Logger:                    d.logger,
	})
}

func fullBootstrapConfig() *config.BootstrapConfig {
	return &config.BootstrapConfig{
		DefaultUser: &config.DefaultPlatformAdminConfig{
			Email:    adminEmail,
			Name:     adminName,
			Password: adminPassword,
		},
		PlatformOrgName:      orgName,
		BaseDomain:           baseDomain,
		RegistryStorageSize:  storageSize,
		RegistryStorageClass: storageClass,
		RegistryName:         registryName,
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

func hashOf(pw string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	Expect(err).NotTo(HaveOccurred())
	return string(hashed)
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

	When("no default cluster is configured", func() {
		It("no-ops without touching any service", func() {
			svc := deps.service(fullBootstrapConfig(), &config.ClusterConfig{})
			Expect(svc.Run(ctx)).To(Succeed())
		})
	})

	When("bootstrapping a fresh install", func() {
		It("provisions org, admin, policies, cluster, registry and domain", func() {
			deps.userSvc.EXPECT().InternalGetByEmail(gomock.Any(), adminEmail).
				Return(nil, errors.NotFound("user not found"))

			deps.orgSvc.EXPECT().InternalCreate(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, spec *models.Organisation) (*models.Organisation, *errors.ServiceError) {
					Expect(spec.Name).To(Equal(orgName))
					Expect(spec.Default).To(BeTrue())
					return &models.Organisation{ID: orgID, Name: orgName, Default: true}, nil
				})

			deps.projSvc.EXPECT().InternalCreateDefaultProject(gomock.Any(), orgID).
				Return(&models.Project{}, nil)

			deps.userSvc.EXPECT().InternalCreate(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, u *models.User) (*models.User, *errors.ServiceError) {
					Expect(u.Email).To(Equal(adminEmail))
					Expect(u.Name).To(Equal(adminName))
					Expect(u.Role).To(Equal(models.OrgAdminRole))
					Expect(u.OrganisationID).To(Equal(orgID))
					Expect(bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(adminPassword))).To(Succeed())
					return &models.User{ID: userID, Email: adminEmail, OrganisationID: orgID}, nil
				})

			deps.policyMgr.EXPECT().AddGroupingPolicy(userID, string(models.OrgAdminRole), orgID).Return(nil)
			deps.policyMgr.EXPECT().AddGroupingPolicy(userID, string(models.OrgMemberRole), orgID).Return(nil)

			deps.clusterSvc.EXPECT().InternalUpsertDefaultCluster(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, spec *models.Cluster) (*models.Cluster, *errors.ServiceError) {
					Expect(spec.Name).To(Equal(clusterName))
					Expect(spec.OrganisationID).To(Equal(orgID))
					Expect(spec.Default).To(BeTrue())
					Expect(spec.ClusterURL).To(Equal(clusterURL))
					Expect(spec.ClusterCAData).To(Equal(clusterCA))
					Expect(spec.Token).To(Equal(clusterToken))
					return &models.Cluster{ID: clusterID}, nil
				})

			deps.registrySvc.EXPECT().GetForOrg(gomock.Any(), orgID).
				Return(nil, errors.NotFound("registry not found"))
			deps.registrySvc.EXPECT().Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError) {
					Expect(spec.Name).To(Equal(registryName))
					Expect(spec.ClusterID).To(Equal(clusterID))
					Expect(spec.OrganisationID).To(Equal(orgID))
					Expect(spec.BackendStorageSize).To(Equal(storageSize))
					Expect(spec.BackendStorageClass).To(Equal(storageClass))
					return &models.ClusterImageRegistry{ID: "reg-1"}, nil
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
		It("is idempotent and does not update the password", func() {
			existingUser := &models.User{ID: userID, Email: adminEmail, OrganisationID: orgID, Password: hashOf(adminPassword)}
			deps.userSvc.EXPECT().InternalGetByEmail(gomock.Any(), adminEmail).Return(existingUser, nil)
			deps.orgSvc.EXPECT().InternalGetDefaultOrg(gomock.Any()).
				Return(&models.Organisation{ID: orgID, Name: orgName, Default: true}, nil)

			deps.clusterSvc.EXPECT().InternalUpsertDefaultCluster(gomock.Any(), gomock.Any()).
				Return(&models.Cluster{ID: clusterID}, nil)

			deps.registrySvc.EXPECT().GetForOrg(gomock.Any(), orgID).
				Return(&models.ClusterImageRegistry{ID: "reg-1"}, nil)
			deps.domainSvc.EXPECT().GetDefaultDomainForOrganisation(gomock.Any(), orgID).
				Return(&models.OrganisationDomain{Domain: baseDomain}, nil)

			svc := deps.service(fullBootstrapConfig(), setClusterConfig())
			Expect(svc.Run(ctx)).To(Succeed())
		})
	})

	When("the admin password changed", func() {
		It("rotates the password exactly once", func() {
			existingUser := &models.User{ID: userID, Email: adminEmail, OrganisationID: orgID, Password: hashOf("stale-password")}
			deps.userSvc.EXPECT().InternalGetByEmail(gomock.Any(), adminEmail).Return(existingUser, nil)
			deps.orgSvc.EXPECT().InternalGetDefaultOrg(gomock.Any()).
				Return(&models.Organisation{ID: orgID, Default: true}, nil)

			deps.userSvc.EXPECT().InternalUpdatePassword(gomock.Any(), userID, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, hashed string) *errors.ServiceError {
					Expect(bcrypt.CompareHashAndPassword([]byte(hashed), []byte(adminPassword))).To(Succeed())
					return nil
				})

			deps.clusterSvc.EXPECT().InternalUpsertDefaultCluster(gomock.Any(), gomock.Any()).
				Return(&models.Cluster{ID: clusterID}, nil)
			deps.registrySvc.EXPECT().GetForOrg(gomock.Any(), orgID).
				Return(&models.ClusterImageRegistry{ID: "reg-1"}, nil)
			deps.domainSvc.EXPECT().GetDefaultDomainForOrganisation(gomock.Any(), orgID).
				Return(&models.OrganisationDomain{Domain: baseDomain}, nil)

			svc := deps.service(fullBootstrapConfig(), setClusterConfig())
			Expect(svc.Run(ctx)).To(Succeed())
		})
	})

	When("no registry name is configured", func() {
		It("skips registry provisioning entirely", func() {
			existingUser := &models.User{ID: userID, Email: adminEmail, OrganisationID: orgID, Password: hashOf(adminPassword)}
			deps.userSvc.EXPECT().InternalGetByEmail(gomock.Any(), adminEmail).Return(existingUser, nil)
			deps.orgSvc.EXPECT().InternalGetDefaultOrg(gomock.Any()).
				Return(&models.Organisation{ID: orgID, Default: true}, nil)

			deps.clusterSvc.EXPECT().InternalUpsertDefaultCluster(gomock.Any(), gomock.Any()).
				Return(&models.Cluster{ID: clusterID}, nil)
			deps.domainSvc.EXPECT().GetDefaultDomainForOrganisation(gomock.Any(), orgID).
				Return(&models.OrganisationDomain{Domain: baseDomain}, nil)

			cfg := fullBootstrapConfig()
			cfg.RegistryName = ""
			svc := deps.service(cfg, setClusterConfig())
			Expect(svc.Run(ctx)).To(Succeed())
		})
	})
})
