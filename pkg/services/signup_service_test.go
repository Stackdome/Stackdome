package services

import (
	"context"

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

var _ = Describe("SignupService", func() {
	const (
		orgName = "Acme Inc"
		orgID   = "11112222-3333-4444-5555-666677778888"
	)

	var (
		ctrl         *gomock.Controller
		userSvc      *mocks.MockUserService
		orgSvc       *mocks.MockOrganisationService
		projectSvc   *mocks.MockProjectService
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

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		userSvc = mocks.NewMockUserService(ctrl)
		orgSvc = mocks.NewMockOrganisationService(ctrl)
		projectSvc = mocks.NewMockProjectService(ctrl)
		policyMgr = mocks.NewMockResourceAccessPolicyManager(ctrl)
		refreshStore = mocks.NewMockRefreshTokenStore(ctrl)

		svc = NewSignupService(SignupServiceSpec{
			UserService:         userSvc,
			OrganisationService: orgSvc,
			ProjectService:      projectSvc,
			PolicyManager:       policyMgr,
			RefreshTokenStore:   refreshStore,
			JWTSecretKey:        "test-secret",
			JWTClaimsBuilder:    auth.NewJWTClaimsBuilder(),
			Logger:              logger.NewLogger(),
		})
		ctx = context.Background()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("creates org, default project, user and role policies", func() {
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
		refreshStore.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&models.RefreshToken{}, nil)

		resp, err := svc.Signup(ctx, newUser(), "")
		Expect(err).To(BeNil())
		Expect(resp).ToNot(BeNil())
	})

	It("fails signup when org creation (incl. platform-infra seeding) fails", func() {
		userSvc.EXPECT().InternalGetByEmail(gomock.Any(), "founder@acme.test").
			Return(nil, errors.NotFound("user not found"))
		orgSvc.EXPECT().InternalCreate(gomock.Any(), gomock.Any()).
			Return(nil, errors.GeneralError("seeding failed"))

		resp, err := svc.Signup(ctx, newUser(), "")
		Expect(resp).To(BeNil())
		Expect(err).ToNot(BeNil())
	})
})
