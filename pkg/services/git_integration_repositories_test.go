package services

import (
	"context"
	"errors"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/clients/githubapp"
	svcerrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

var _ = Describe("GitIntegrationService.ListRepositories", func() {
	var (
		ctrl          *gomock.Controller
		store         *mocks.MockGitIntegrationStore
		installations *mocks.MockGitInstallationStore
		appClient     *mocks.MockGitHubAppClient
		svc           GitIntegrationService
		ctx           context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.Background()

		encryption, err := NewAESEncryptionService(EncryptionServiceSpec{Masterkey: strings.Repeat("k", 64)})
		Expect(err).ToNot(HaveOccurred())

		store = mocks.NewMockGitIntegrationStore(ctrl)
		installations = mocks.NewMockGitInstallationStore(ctrl)
		appClient = mocks.NewMockGitHubAppClient(ctrl)

		permissions := mocks.NewMockPermissionService(ctrl)
		permissions.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		organisations := mocks.NewMockOrganisationStore(ctrl)
		organisations.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&models.Organisation{ID: "org-1"}, nil).AnyTimes()
		atomic := mocks.NewMockAtomicExecutor(ctrl)
		atomic.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(context.Context) *svcerrors.ServiceError) *svcerrors.ServiceError {
				return fn(ctx)
			}).AnyTimes()

		// A sealed, installed GitHub App integration the service can unseal.
		app := models.GitHubAppCredentials{AppID: 4242, Slug: "stackdome-test", PEM: "PEM"}
		authBlob, mErr := (models.GitIntegrationAuth{GitHubApp: &app}).Marshal()
		Expect(mErr).ToNot(HaveOccurred())
		encrypted, eErr := encryption.EncryptData(authBlob)
		Expect(eErr).ToNot(HaveOccurred())
		integration := &models.GitIntegration{
			ID:             "gi-app",
			OrganisationID: "org-1",
			Type:           models.GitIntegrationTypeGitHubApp,
			Host:           "github.com",
			Status:         models.GitIntegrationStatusInstalled,
			EncryptedAuth:  encrypted,
			DataHash:       gitIntegrationDataHash(authBlob),
		}
		store.EXPECT().GetByID(gomock.Any(), "gi-app").Return(integration, nil).AnyTimes()

		svc = NewGitIntegrationService(GitIntegrationServiceSpec{
			Store:             store,
			InstallationStore: installations,
			OAuthStateStore:   mocks.NewMockOAuthStateStore(ctrl),
			OrganisationStore: organisations,
			AtomicExecutor:    atomic,
			GitHubAppClient:   appClient,
			EncryptionService: encryption,
			Permissions:       permissions,
			Logger:            logger.NewLogger(),
			ExternalURL:       "https://hub.example.com",
		})
	})

	Context("when an installation UUID is given", func() {
		It("resolves the UUID and lists that installation's page", func() {
			installations.EXPECT().GetByIntegrationAndID(gomock.Any(), "gi-app", "inst-uuid-1").
				Return(&models.GitInstallation{ID: "inst-uuid-1", InstallationID: 77}, nil)
			appClient.EXPECT().ListInstallationRepos(gomock.Any(), gomock.Any(), int64(77), 2).
				Return(&githubapp.RepoPage{Page: 2, TotalCount: 3, HasNext: true, Repos: []githubapp.Repo{{FullName: "acme/a"}}}, nil)

			page, serr := svc.ListRepositories(ctx, "gi-app", 2, "inst-uuid-1")
			Expect(serr).To(BeNil())
			Expect(page.Repos).To(HaveLen(1))
			Expect(page.TotalCount).To(Equal(3))
			Expect(page.HasNext).To(BeTrue())
		})

		It("returns NotFound when the UUID does not belong to the integration", func() {
			installations.EXPECT().GetByIntegrationAndID(gomock.Any(), "gi-app", "missing").
				Return(nil, svcerrors.NotFound("no installation 'missing' for this integration"))

			page, serr := svc.ListRepositories(ctx, "gi-app", 1, "missing")
			Expect(page).To(BeNil())
			Expect(serr).ToNot(BeNil())
			Expect(serr.Is404()).To(BeTrue())
		})
	})

	Context("when no installation UUID is given", func() {
		It("aggregates page N across every installation in order", func() {
			installations.EXPECT().ListByIntegrationID(gomock.Any(), "gi-app").Return([]*models.GitInstallation{
				{ID: "u-a", InstallationID: 11},
				{ID: "u-b", InstallationID: 22},
			}, nil)
			appClient.EXPECT().ListInstallationRepos(gomock.Any(), gomock.Any(), int64(11), 1).
				Return(&githubapp.RepoPage{TotalCount: 5, HasNext: true, Repos: []githubapp.Repo{{FullName: "acme/a"}}}, nil)
			appClient.EXPECT().ListInstallationRepos(gomock.Any(), gomock.Any(), int64(22), 1).
				Return(&githubapp.RepoPage{TotalCount: 2, HasNext: false, Repos: []githubapp.Repo{{FullName: "bob/b"}}}, nil)

			page, serr := svc.ListRepositories(ctx, "gi-app", 1, "")
			Expect(serr).To(BeNil())
			Expect([]string{page.Repos[0].FullName, page.Repos[1].FullName}).To(Equal([]string{"acme/a", "bob/b"}))
			Expect(page.TotalCount).To(Equal(7))
			Expect(page.HasNext).To(BeTrue())
			Expect(page.Page).To(Equal(1))
		})

		It("returns an empty page when there are no installations", func() {
			installations.EXPECT().ListByIntegrationID(gomock.Any(), "gi-app").Return(nil, nil)

			page, serr := svc.ListRepositories(ctx, "gi-app", 1, "")
			Expect(serr).To(BeNil())
			Expect(page.Repos).To(BeEmpty())
			Expect(page.TotalCount).To(BeZero())
			Expect(page.HasNext).To(BeFalse())
		})

		It("propagates an error when one installation fetch fails", func() {
			installations.EXPECT().ListByIntegrationID(gomock.Any(), "gi-app").Return([]*models.GitInstallation{
				{ID: "u-a", InstallationID: 11},
				{ID: "u-b", InstallationID: 22},
			}, nil)
			appClient.EXPECT().ListInstallationRepos(gomock.Any(), gomock.Any(), int64(11), 1).
				Return(&githubapp.RepoPage{Repos: []githubapp.Repo{{FullName: "acme/a"}}}, nil).AnyTimes()
			appClient.EXPECT().ListInstallationRepos(gomock.Any(), gomock.Any(), int64(22), 1).
				Return(nil, errors.New("github boom")).AnyTimes()

			page, serr := svc.ListRepositories(ctx, "gi-app", 1, "")
			Expect(page).To(BeNil())
			Expect(serr).ToNot(BeNil())
			Expect(serr.Code).To(Equal(svcerrors.ErrorBadRequest))
		})
	})
})
