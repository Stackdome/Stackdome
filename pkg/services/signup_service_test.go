package services

import (
	"context"
	"regexp"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

// Suite bootstrapped by TestServices in services_suite_test.go.

var _ = Describe("SignupService default-infra seeding", func() {
	const (
		orgName        = "Acme Inc"
		orgID          = "11112222-3333-4444-5555-666677778888"
		platformOrgID  = "platform-org"
		baseDomain     = "apps.example.com"
		expectedSlug   = "acme-inc"
		expectedDomain = expectedSlug + "." + baseDomain
		registryName   = expectedSlug + "-11112222"
		storageSize    = "50Gi"
		storageClass   = "standard"
	)

	var (
		ctrl         *gomock.Controller
		userSvc      *mocks.MockUserService
		orgSvc       *mocks.MockOrganisationService
		projectSvc   *mocks.MockProjectService
		clusterSvc   *mocks.MockClusterService
		domainSvc    *mocks.MockOrganisationDomainsService
		registrySvc  *mocks.MockImageRegistryService
		policyMgr    *mocks.MockResourceAccessPolicyManager
		refreshStore *mocks.MockRefreshTokenStore
		svc          SignupService
		ctx          context.Context
	)

	newUser := func() *models.User {
		return &models.User{
			Email:        "founder@acme.test",
			Password:     "supersecret",
			Organisation: &models.Organisation{Name: orgName},
		}
	}

	expectUserAndOrgCreation := func() {
		userSvc.EXPECT().InternalGetByEmail(gomock.Any(), "founder@acme.test").
			Return(nil, errors.NotFound("user not found"))
		orgSvc.EXPECT().InternalCreate(gomock.Any(), gomock.Any()).
			Return(&models.Organisation{ID: orgID, Name: orgName}, nil)
		projectSvc.EXPECT().InternalCreateDefaultProject(gomock.Any(), orgID).
			Return(&models.Project{}, nil)
		userSvc.EXPECT().InternalCreate(gomock.Any(), gomock.Any()).
			Return(&models.User{ID: "user-1", OrganisationID: orgID}, nil)
		policyMgr.EXPECT().AddGroupingPolicy("user-1", string(models.OrgAdminRole), orgID).Return(nil)
		policyMgr.EXPECT().AddGroupingPolicy("user-1", string(models.OrgMemberRole), orgID).Return(nil)
	}

	expectResponseBuilt := func() {
		refreshStore.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&models.RefreshToken{}, nil)
	}

	expectDefaultClusterAndPlatform := func() {
		clusterSvc.EXPECT().GetDefaultCluster(gomock.Any()).Return(&models.Cluster{ID: "cluster-1"}, nil)
		orgSvc.EXPECT().InternalGetDefaultOrg(gomock.Any()).Return(&models.Organisation{ID: platformOrgID}, nil)
		domainSvc.EXPECT().GetDefaultDomainForOrganisation(gomock.Any(), platformOrgID).
			Return(&models.OrganisationDomain{Domain: baseDomain}, nil)
	}

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		userSvc = mocks.NewMockUserService(ctrl)
		orgSvc = mocks.NewMockOrganisationService(ctrl)
		projectSvc = mocks.NewMockProjectService(ctrl)
		clusterSvc = mocks.NewMockClusterService(ctrl)
		domainSvc = mocks.NewMockOrganisationDomainsService(ctrl)
		registrySvc = mocks.NewMockImageRegistryService(ctrl)
		policyMgr = mocks.NewMockResourceAccessPolicyManager(ctrl)
		refreshStore = mocks.NewMockRefreshTokenStore(ctrl)

		svc = NewSignupService(SignupServiceSpec{
			UserService:               userSvc,
			OrganisationService:       orgSvc,
			ProjectService:            projectSvc,
			ClusterService:            clusterSvc,
			OrganisationDomainService: domainSvc,
			ImageRegistryService:      registrySvc,
			PolicyManager:             policyMgr,
			RefreshTokenStore:         refreshStore,
			JWTSecretKey:              "test-secret",
			JWTClaimsBuilder:          auth.NewJWTClaimsBuilder(),
			RegistryStorageSize:       storageSize,
			RegistryStorageClass:      storageClass,
			Logger:                    logger.NewLogger(),
		})
		ctx = context.Background()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("skips seeding and signs up when no default cluster exists", func() {
		expectUserAndOrgCreation()
		clusterSvc.EXPECT().GetDefaultCluster(gomock.Any()).
			Return(nil, errors.NotFound("default cluster not found"))
		expectResponseBuilt()

		resp, err := svc.Signup(ctx, newUser(), "")
		Expect(err).To(BeNil())
		Expect(resp).ToNot(BeNil())
	})

	It("seeds the org domain and registry on the default cluster", func() {
		expectUserAndOrgCreation()
		expectDefaultClusterAndPlatform()
		domainSvc.EXPECT().Create(gomock.Any(), &models.OrganisationDomain{
			OrganisationID: orgID,
			Domain:         expectedDomain,
		}).Return(&models.OrganisationDomain{}, nil)
		registrySvc.EXPECT().Create(gomock.Any(), &models.ClusterImageRegistry{
			ClusterID:           "cluster-1",
			OrganisationID:      orgID,
			Name:                registryName,
			BackendStorageSize:  storageSize,
			BackendStorageClass: storageClass,
		}).Return(&models.ClusterImageRegistry{}, nil)
		expectResponseBuilt()

		resp, err := svc.Signup(ctx, newUser(), "")
		Expect(err).To(BeNil())
		Expect(resp).ToNot(BeNil())
	})

	It("retries the domain with a random suffix when the first attempt conflicts", func() {
		expectUserAndOrgCreation()
		expectDefaultClusterAndPlatform()
		suffixed := regexp.MustCompile(`^` + expectedSlug + `-[0-9a-f]{6}\.` + regexp.QuoteMeta(baseDomain) + `$`)
		gomock.InOrder(
			domainSvc.EXPECT().Create(gomock.Any(), &models.OrganisationDomain{
				OrganisationID: orgID,
				Domain:         expectedDomain,
			}).Return(nil, errors.Conflict("domain already exists")),
			domainSvc.EXPECT().Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, spec *models.OrganisationDomain) (*models.OrganisationDomain, *errors.ServiceError) {
					Expect(spec.Domain).To(MatchRegexp(suffixed.String()))
					return &models.OrganisationDomain{}, nil
				}),
		)
		registrySvc.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&models.ClusterImageRegistry{}, nil)
		expectResponseBuilt()

		resp, err := svc.Signup(ctx, newUser(), "")
		Expect(err).To(BeNil())
		Expect(resp).ToNot(BeNil())
	})

	It("fails signup when domain creation returns a non-conflict error", func() {
		expectUserAndOrgCreation()
		expectDefaultClusterAndPlatform()
		boom := errors.GeneralError("domain store unavailable")
		domainSvc.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, boom)

		resp, err := svc.Signup(ctx, newUser(), "")
		Expect(resp).To(BeNil())
		Expect(err).To(Equal(boom))
	})

	It("fails signup when registry creation fails", func() {
		expectUserAndOrgCreation()
		expectDefaultClusterAndPlatform()
		domainSvc.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&models.OrganisationDomain{}, nil)
		boom := errors.GeneralError("cluster unreachable")
		registrySvc.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, boom)

		resp, err := svc.Signup(ctx, newUser(), "")
		Expect(resp).To(BeNil())
		Expect(err).To(Equal(boom))
	})
})
