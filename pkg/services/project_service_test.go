package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/resourceaccess"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

// Suite bootstrapped by TestAESEncryptionService in encryption_service_test.go.

var _ = Describe("ProjectService name resolution", func() {
	const (
		orgID       = "org-1"
		projectID   = "proj-1"
		projectName = "my-project"
		userID      = "user-1"
	)

	var (
		svc         *projectService
		tokenCtx    context.Context
		testProject *models.Project
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())
		testProject = &models.Project{ID: projectID, OrganisationID: orgID, Name: projectName}

		projectStore := mocks.NewMockProjectStore(ctrl)
		projectStore.EXPECT().GetByOrgAndName(gomock.Any(), orgID, projectName).Return(testProject, nil).AnyTimes()
		projectStore.EXPECT().GetByID(gomock.Any(), projectID).Return(testProject, nil).AnyTimes()

		logMock := mocks.NewMockLogger(ctrl)
		logMock.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
		logMock.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

		policyMgr, err := resourceaccess.NewInMemoryPolicyManager()
		Expect(err).ToNot(HaveOccurred())
		Expect(auth.LoadDefaultPolicies(policyMgr.AddPolicy)).To(Succeed())
		Expect(policyMgr.AddGroupingPolicy(userID, string(models.DeveloperRole), projectID)).To(Succeed())
		Expect(policyMgr.AddGroupingPolicy(userID, string(models.OrgMemberRole), orgID)).To(Succeed())

		svc = &projectService{
			projectStore: projectStore,
			policyMgr:    policyMgr,
			permissions: auth.NewPermissionService(auth.PermissionServiceSpec{
				PolicyManager: policyMgr,
				ProjectStore:  projectStore,
				Logger:        logMock,
			}),
			logger: logMock,
		}

		tokenCtx = auth.SetIdentityInContext(context.Background(), &auth.Identity{
			UserID:      userID,
			OrgID:       orgID,
			AuthMethod:  auth.AuthMethodAPIToken,
			TokenScopes: []string{"stacks:*"},
		})
	})

	It("resolves the project name for a stacks-scoped token", func() {
		project, serr := svc.InternalGetProjectByOrgAndName(tokenCtx, orgID, projectName)
		Expect(serr).To(BeNil())
		Expect(project.ID).To(Equal(projectID))
	})

	It("still denies the project resource itself to a stacks-scoped token", func() {
		_, serr := svc.GetProjectByOrgAndName(tokenCtx, orgID, projectName)
		Expect(serr).ToNot(BeNil())
		Expect(serr.Code).To(Equal(errors.ErrorForbidden))
	})

	It("allows the project resource for a projects-scoped token", func() {
		ctx := auth.SetIdentityInContext(context.Background(), &auth.Identity{
			UserID:      userID,
			OrgID:       orgID,
			AuthMethod:  auth.AuthMethodAPIToken,
			TokenScopes: []string{"projects:*"},
		})
		project, serr := svc.GetProjectByOrgAndName(ctx, orgID, projectName)
		Expect(serr).To(BeNil())
		Expect(project.ID).To(Equal(projectID))
	})
})

var _ = Describe("ProjectService launch tenancy invariant", func() {
	const (
		orgID     = "org-1"
		userID    = "user-1"
		projectID = "proj-1"
	)

	var (
		ctrl            *gomock.Controller
		projectStore    *mocks.MockProjectStore
		membershipStore *mocks.MockProjectMembershipStore
		userStore       *mocks.MockUserStore
		permissions     *mocks.MockPermissionService
		svc             *projectService
		ctx             context.Context
	)

	defaultProject := &models.Project{ID: projectID, OrganisationID: orgID, Name: models.DefaultProjectName, DefaultProject: true}

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		projectStore = mocks.NewMockProjectStore(ctrl)
		membershipStore = mocks.NewMockProjectMembershipStore(ctrl)
		userStore = mocks.NewMockUserStore(ctrl)
		permissions = mocks.NewMockPermissionService(ctrl)
		permissions.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		logMock := mocks.NewMockLogger(ctrl)
		logMock.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		svc = &projectService{
			projectStore:    projectStore,
			membershipStore: membershipStore,
			userStore:       userStore,
			permissions:     permissions,
			logger:          logMock,
		}
		ctx = context.Background()
	})

	AfterEach(func() { ctrl.Finish() })

	It("refuses a second project and writes nothing", func() {
		projectStore.EXPECT().ListByOrgID(gomock.Any(), orgID).Return([]*models.Project{defaultProject}, nil)

		created, serr := svc.CreateProject(ctx, orgID, &models.Project{Name: "staging"})
		Expect(created).To(BeNil())
		Expect(serr).ToNot(BeNil())
		Expect(serr.Code).To(Equal(errors.ErrorConflict))
	})

	It("creates the first project of an organisation", func() {
		projectStore.EXPECT().ListByOrgID(gomock.Any(), orgID).Return(nil, nil)
		projectStore.EXPECT().Create(gomock.Any(), gomock.Any()).Return(defaultProject, nil)

		created, serr := svc.CreateProject(ctx, orgID, &models.Project{Name: "staging"})
		Expect(serr).To(BeNil())
		Expect(created).To(Equal(defaultProject))
	})

	It("lists the org's projects for an org admin with no membership rows", func() {
		userStore.EXPECT().GetByID(gomock.Any(), userID).
			Return(&models.User{ID: userID, OrganisationID: orgID, Role: models.OrgAdminRole}, nil)
		projectStore.EXPECT().ListByOrgID(gomock.Any(), orgID).Return([]*models.Project{defaultProject}, nil)

		memberships, serr := svc.ListUserProjects(ctx, userID)
		Expect(serr).To(BeNil())
		Expect(memberships).To(HaveLen(1))
		Expect(memberships[0].ProjectID).To(Equal(projectID))
		Expect(memberships[0].UserID).To(Equal(userID))
		Expect(memberships[0].Role).To(Equal(models.DeveloperRole))
		Expect(memberships[0].Project.Name).To(Equal(models.DefaultProjectName))
	})

	It("keeps the membership-backed list for non-admins", func() {
		member := &models.ProjectMembership{ID: "m-1", ProjectID: projectID, UserID: userID, Role: models.DeveloperRole}
		userStore.EXPECT().GetByID(gomock.Any(), userID).
			Return(&models.User{ID: userID, OrganisationID: orgID, Role: models.NoRole}, nil)
		membershipStore.EXPECT().ListByUserIDAndOrgID(gomock.Any(), userID, orgID).
			Return([]*models.ProjectMembership{member}, nil)

		memberships, serr := svc.ListUserProjects(ctx, userID)
		Expect(serr).To(BeNil())
		Expect(memberships).To(Equal([]*models.ProjectMembership{member}))
	})
})
