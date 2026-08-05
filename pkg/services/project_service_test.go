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
