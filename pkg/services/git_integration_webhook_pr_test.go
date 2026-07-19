package services

import (
	"context"
	"encoding/json"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

var _ = Describe("ProcessGitHubWebhook pull_request dispatch", func() {
	const (
		appID      = 4242
		hookSecret = "hook-secret"
	)
	var (
		ctrl               *gomock.Controller
		svc                GitIntegrationService
		previewWebhook     *MockPreviewWebhookService
		installationsStore *mocks.MockGitInstallationStore
	)

	const baseRepoID = 1010

	// Real repository-event payloads carry only the slim installation
	// reference {id, node_id} — never app_id.
	prPayloadFromRepo := func(action string, headRepoID int) []byte {
		payload, err := json.Marshal(map[string]any{
			"action":       action,
			"number":       7,
			"installation": map[string]any{"id": 77},
			"repository":   map[string]any{"id": baseRepoID, "clone_url": "https://github.com/acme/app.git"},
			"pull_request": map[string]any{
				"head": map[string]any{"ref": "feature", "sha": "abc123", "repo": map[string]any{"id": headRepoID}},
				"base": map[string]any{"ref": "main"},
			},
		})
		Expect(err).ToNot(HaveOccurred())
		return payload
	}

	prPayload := func(action string) []byte {
		return prPayloadFromRepo(action, baseRepoID)
	}

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())

		encryption, err := NewAESEncryptionService(EncryptionServiceSpec{Masterkey: strings.Repeat("k", 64)})
		Expect(err).ToNot(HaveOccurred())

		app := models.GitHubAppCredentials{AppID: appID, Slug: "stackdome-test", WebhookSecret: hookSecret}
		blob, mErr := (models.GitIntegrationAuth{GitHubApp: &app}).Marshal()
		Expect(mErr).ToNot(HaveOccurred())
		encrypted, eErr := encryption.EncryptData(blob)
		Expect(eErr).ToNot(HaveOccurred())
		integration := &models.GitIntegration{
			ID:             "gi-app",
			OrganisationID: "org-1",
			Type:           models.GitIntegrationTypeGitHubApp,
			Host:           "github.com",
			Status:         models.GitIntegrationStatusInstalled,
			EncryptedAuth:  encrypted,
			DataHash:       gitIntegrationDataHash(blob),
		}

		store := mocks.NewMockGitIntegrationStore(ctrl)
		store.EXPECT().GetByID(gomock.Any(), "gi-app").Return(integration, nil).AnyTimes()
		installationsStore = mocks.NewMockGitInstallationStore(ctrl)
		installationsStore.EXPECT().GetByInstallationID(gomock.Any(), int64(77)).
			Return(&models.GitInstallation{GitIntegrationID: "gi-app", InstallationID: 77}, nil).AnyTimes()
		previewWebhook = NewMockPreviewWebhookService(ctrl)

		svc = NewGitIntegrationService(GitIntegrationServiceSpec{
			Store:             store,
			InstallationStore: installationsStore,
			OAuthStateStore:   mocks.NewMockOAuthStateStore(ctrl),
			OrganisationStore: mocks.NewMockOrganisationStore(ctrl),
			AtomicExecutor:    mocks.NewMockAtomicExecutor(ctrl),
			GitHubAppClient:   mocks.NewMockGitHubAppClient(ctrl),
			EncryptionService: encryption,
			Permissions:       mocks.NewMockPermissionService(ctrl),
			Logger:            logger.NewLogger(),
			ExternalURL:       "https://hub.example.com",
		})
		WirePreviewWebhook(svc, previewWebhook)
	})

	DescribeTable("routes the PR action with a populated event",
		func(action string, expected PullRequestAction) {
			var got PullRequestEvent
			previewWebhook.EXPECT().HandlePullRequest(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, ev PullRequestEvent) *errors.ServiceError {
					got = ev
					return nil
				})

			payload := prPayload(action)
			serr := svc.ProcessGitHubWebhook(context.Background(), GitHubEventPullRequest, payload, signWebhook(payload, hookSecret))
			Expect(serr).To(BeNil())
			Expect(got.Action).To(Equal(expected))
			Expect(got.OrganisationID).To(Equal("org-1"))
			Expect(got.RepoURL).To(Equal("https://github.com/acme/app.git"))
			Expect(got.Branch).To(Equal("feature"))
			Expect(got.BaseBranch).To(Equal("main"))
			Expect(got.HeadSHA).To(Equal("abc123"))
			Expect(got.PRNumber).To(Equal("7"))
		},
		Entry("opened", "opened", PRActionOpened),
		Entry("synchronize", "synchronize", PRActionSynchronize),
		Entry("closed", "closed", PRActionClosed),
	)

	It("drops fork PRs without dispatching", func() {
		// No HandlePullRequest expectation — must not be called.
		payload := prPayloadFromRepo("opened", baseRepoID+1)
		serr := svc.ProcessGitHubWebhook(context.Background(), GitHubEventPullRequest, payload, signWebhook(payload, hookSecret))
		Expect(serr).To(BeNil())
	})

	It("rejects deliveries for an unknown installation id", func() {
		// No HandlePullRequest expectation — must not be called.
		payload, err := json.Marshal(map[string]any{
			"action":       "opened",
			"number":       7,
			"installation": map[string]any{"id": 999},
			"repository":   map[string]any{"id": baseRepoID, "clone_url": "https://github.com/acme/app.git"},
			"pull_request": map[string]any{
				"head": map[string]any{"ref": "feature", "sha": "abc123", "repo": map[string]any{"id": baseRepoID}},
				"base": map[string]any{"ref": "main"},
			},
		})
		Expect(err).ToNot(HaveOccurred())
		installationsStore.EXPECT().GetByInstallationID(gomock.Any(), int64(999)).
			Return(nil, errors.NotFound("no installation with id %d", 999))

		serr := svc.ProcessGitHubWebhook(context.Background(), GitHubEventPullRequest, payload, signWebhook(payload, hookSecret))
		Expect(serr).ToNot(BeNil())
		Expect(serr.Is404()).To(BeTrue())
	})

	It("rejects a bad signature without dispatching", func() {
		// No HandlePullRequest expectation — must not be called.
		payload := prPayload("opened")
		serr := svc.ProcessGitHubWebhook(context.Background(), GitHubEventPullRequest, payload, signWebhook(payload, "wrong-secret"))
		Expect(serr).ToNot(BeNil())
		Expect(serr.IsForbidden()).To(BeTrue())
	})
})
