package services

import (
	"context"
	stderrors "errors"
	"strings"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("UserService OAuth signup", func() {
	const (
		orgID    = "org-1"
		userID   = "user-1"
		email    = "founder@acme.test"
		githubID = "gh-1"
		avatar   = "https://example.test/a.png"
	)

	var (
		ctrl       *gomock.Controller
		userStore  *mocks.MockUserStore
		orgSvc     *mocks.MockOrganisationService
		projectSvc *mocks.MockProjectService
		policyMgr  *mocks.MockResourceAccessPolicyManager
		atomicExec *mocks.MockAtomicExecutor
		rolledBack bool
		svc        usersService
		ctx        context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		userStore = mocks.NewMockUserStore(ctrl)
		orgSvc = mocks.NewMockOrganisationService(ctrl)
		projectSvc = mocks.NewMockProjectService(ctrl)
		policyMgr = mocks.NewMockResourceAccessPolicyManager(ctrl)

		rolledBack = false
		atomicExec = mocks.NewMockAtomicExecutor(ctrl)
		atomicExec.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(txCtx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				serr := fn(txCtx)
				rolledBack = serr != nil
				return serr
			}).AnyTimes()

		svc = usersService{
			userStore:               userStore,
			organisationService:     orgSvc,
			projectService:          projectSvc,
			resourceAccessPolicyMgr: policyMgr,
			atomicExecutor:          atomicExec,
			logger:                  logger.NewLogger(),
		}
		ctx = context.Background()
	})

	AfterEach(func() { ctrl.Finish() })

	It("creates org, default project, OrgAdmin user and role policies", func() {
		orgSvc.EXPECT().InternalCreate(gomock.Any(), gomock.Any()).
			Return(&models.Organisation{ID: orgID, Name: models.UserOrgNameFromOauth("Ada")}, nil)
		projectSvc.EXPECT().InternalCreateDefaultProject(gomock.Any(), orgID).Return(&models.Project{}, nil)
		userStore.EXPECT().Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, u *models.User) (*models.User, *errors.ServiceError) {
				Expect(u.Role).To(Equal(models.OrgAdminRole))
				Expect(u.OrganisationID).To(Equal(orgID))
				return &models.User{ID: userID, OrganisationID: orgID, Role: models.OrgAdminRole}, nil
			})
		policyMgr.EXPECT().AddGroupingPolicy(userID, string(models.OrgAdminRole), orgID).Return(nil)
		policyMgr.EXPECT().AddGroupingPolicy(userID, string(models.OrgMemberRole), orgID).Return(nil)

		user, err := svc.InternalCreateOAuthUser(ctx, email, "Ada", githubID, avatar)
		Expect(err).ToNot(HaveOccurred())
		Expect(user.Role).To(Equal(models.OrgAdminRole))
		Expect(rolledBack).To(BeFalse())
	})

	It("rolls back when the default project cannot be created", func() {
		orgSvc.EXPECT().InternalCreate(gomock.Any(), gomock.Any()).
			Return(&models.Organisation{ID: orgID}, nil)
		projectSvc.EXPECT().InternalCreateDefaultProject(gomock.Any(), orgID).
			Return(nil, errors.GeneralError("boom"))

		user, err := svc.InternalCreateOAuthUser(ctx, email, "Ada", githubID, avatar)
		Expect(user).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(rolledBack).To(BeTrue())
	})

	It("rolls back and revokes the role grants when a policy write fails", func() {
		orgSvc.EXPECT().InternalCreate(gomock.Any(), gomock.Any()).
			Return(&models.Organisation{ID: orgID}, nil)
		projectSvc.EXPECT().InternalCreateDefaultProject(gomock.Any(), orgID).Return(&models.Project{}, nil)
		userStore.EXPECT().Create(gomock.Any(), gomock.Any()).
			Return(&models.User{ID: userID, OrganisationID: orgID}, nil)
		policyMgr.EXPECT().AddGroupingPolicy(userID, string(models.OrgAdminRole), orgID).
			Return(stderrors.New("casbin down"))
		policyMgr.EXPECT().RemoveGroupingPolicy(userID, string(models.OrgAdminRole), orgID).Return(nil)
		policyMgr.EXPECT().RemoveGroupingPolicy(userID, string(models.OrgMemberRole), orgID).Return(nil)

		user, err := svc.InternalCreateOAuthUser(ctx, email, "Ada", githubID, avatar)
		Expect(user).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(rolledBack).To(BeTrue())
	})

	It("rejects an oversized organisation name before touching the database", func() {
		user, err := svc.InternalCreateOAuthUser(ctx, email, strings.Repeat("a", maxOrgNameLength+1), githubID, avatar)
		Expect(user).To(BeNil())
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("validateOrganisationName", func() {
	DescribeTable("rejects",
		func(name string) {
			Expect(validateOrganisationName(name)).ToNot(BeNil())
		},
		Entry("empty", ""),
		Entry("whitespace only", "  \t "),
		Entry("punctuation only", "!!!"),
		Entry("too long", strings.Repeat("a", maxOrgNameLength+1)),
	)

	DescribeTable("accepts",
		func(name string) {
			Expect(validateOrganisationName(name)).To(BeNil())
		},
		Entry("normal name", "Acme Labs"),
		Entry("digits", "42"),
		Entry("oauth derived name", models.UserOrgNameFromOauth("Ada Lovelace")),
		Entry("at the cap", strings.Repeat("a", maxOrgNameLength)),
	)
})
