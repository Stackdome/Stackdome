package services

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/clients/githubapp"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

var _ = Describe("PreviewCommentService", func() {
	var (
		ctrl            *gomock.Controller
		configs         *mocks.MockStackPreviewConfigStore
		gitIntegrations *mocks.MockGitIntegrationService
		commenter       *githubapp.MockPullRequestCommenter
		svc             PreviewCommentService
		ctx             context.Context
		preview         *models.PreviewStack
		config          *models.StackPreviewConfig
		mint            *models.GitHubAppMintResult
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		configs = mocks.NewMockStackPreviewConfigStore(ctrl)
		gitIntegrations = mocks.NewMockGitIntegrationService(ctrl)
		commenter = githubapp.NewMockPullRequestCommenter(ctrl)
		svc = NewPreviewCommentService(PreviewCommentServiceSpec{
			ConfigStore:     configs,
			GitIntegrations: gitIntegrations,
			Commenter:       commenter,
			Logger:          logger.NewLogger(),
		})
		ctx = context.Background()

		config = &models.StackPreviewConfig{
			ID:             "config-1",
			OrganisationID: "org-1",
			GitRepository:  models.PreviewGitRepository{RepoURL: "https://github.com/acme/app.git", BaseBranch: "main"},
		}
		preview = &models.PreviewStack{
			ID:                   "preview-1",
			StackPreviewConfigID: "config-1",
			PRNumber:             "7",
			Status: models.PreviewStackStatus{
				Phase: models.PreviewStackPhaseReady,
				Outputs: &models.PreviewStackOutputs{
					CommitSHA: "abc123",
					URLs:      []models.PreviewURL{{Resource: "web", URL: "https://pr-7-web.example.dev"}},
				},
			},
		}
		mint = &models.GitHubAppMintResult{Token: "tok"}
	})

	It("creates the comment and sets its id when none exists", func() {
		configs.EXPECT().GetByID(ctx, "config-1").Return(config, nil)
		gitIntegrations.EXPECT().InternalMintForRepo(ctx, "org-1", config.GitRepository.RepoURL).Return(mint, nil)
		commenter.EXPECT().CreateComment(ctx, "tok", "acme", "app", 7, gomock.Any()).
			DoAndReturn(func(_ context.Context, _, _, _ string, _ int, body string) (int64, error) {
				Expect(body).To(ContainSubstring("Preview environment is live"))
				Expect(body).To(ContainSubstring("`abc123`"))
				Expect(body).To(ContainSubstring("https://pr-7-web.example.dev"))
				return 4242, nil
			})

		Expect(svc.InternalUpsertComment(ctx, preview)).To(Succeed())
		Expect(preview.GitHubCommentID).To(Equal(int64(4242)))
	})

	It("edits by stored id when the comment already exists", func() {
		preview.GitHubCommentID = 4242
		configs.EXPECT().GetByID(ctx, "config-1").Return(config, nil)
		gitIntegrations.EXPECT().InternalMintForRepo(ctx, "org-1", config.GitRepository.RepoURL).Return(mint, nil)
		commenter.EXPECT().EditComment(ctx, "tok", "acme", "app", int64(4242), gomock.Any()).Return(nil)

		Expect(svc.InternalUpsertComment(ctx, preview)).To(Succeed())
	})

	It("re-creates and sets a new id when the edit target is gone", func() {
		preview.GitHubCommentID = 4242
		configs.EXPECT().GetByID(ctx, "config-1").Return(config, nil)
		gitIntegrations.EXPECT().InternalMintForRepo(ctx, "org-1", config.GitRepository.RepoURL).Return(mint, nil)
		commenter.EXPECT().EditComment(ctx, "tok", "acme", "app", int64(4242), gomock.Any()).Return(githubapp.ErrCommentNotFound)
		commenter.EXPECT().CreateComment(ctx, "tok", "acme", "app", 7, gomock.Any()).Return(int64(5000), nil)

		Expect(svc.InternalUpsertComment(ctx, preview)).To(Succeed())
		Expect(preview.GitHubCommentID).To(Equal(int64(5000)))
	})

	It("skips silently when no App installation covers the repo", func() {
		configs.EXPECT().GetByID(ctx, "config-1").Return(config, nil)
		gitIntegrations.EXPECT().InternalMintForRepo(ctx, "org-1", config.GitRepository.RepoURL).
			Return(nil, errors.NotFound("no app"))

		Expect(svc.InternalUpsertComment(ctx, preview)).To(Succeed())
	})

	It("skips silently for non-GitHub repo URLs", func() {
		config.GitRepository.RepoURL = "https://gitlab.com/acme/app.git"
		configs.EXPECT().GetByID(ctx, "config-1").Return(config, nil)

		Expect(svc.InternalUpsertComment(ctx, preview)).To(Succeed())
	})

	It("skips silently when the PR number is not numeric", func() {
		preview.PRNumber = "not-a-number"
		configs.EXPECT().GetByID(ctx, "config-1").Return(config, nil)

		Expect(svc.InternalUpsertComment(ctx, preview)).To(Succeed())
	})

	It("skips torn-down previews that were never announced", func() {
		now := time.Now()
		preview.DeletionTimestamp = &now
		preview.GitHubCommentID = 0

		Expect(svc.InternalUpsertComment(ctx, preview)).To(Succeed())
	})

	Describe("renderPreviewComment", func() {
		It("renders the deleted state without a URL table", func() {
			now := time.Now()
			preview.DeletionTimestamp = &now
			body := renderPreviewComment(preview)
			Expect(body).To(ContainSubstring("Preview environment deleted"))
			Expect(body).ToNot(ContainSubstring("| Resource |"))
			Expect(body).To(ContainSubstring(previewCommentMarker))
		})

		It("renders the failed state with the status message and last known URLs", func() {
			preview.Status.Phase = models.PreviewStackPhaseFailed
			preview.Status.Message = "release failed: image pull backoff"
			body := renderPreviewComment(preview)
			Expect(body).To(ContainSubstring("Preview deploy failed"))
			Expect(body).To(ContainSubstring("image pull backoff"))
			Expect(body).To(ContainSubstring("https://pr-7-web.example.dev"))
		})

		It("renders live without outputs when none exist yet", func() {
			preview.Status.Outputs = nil
			body := renderPreviewComment(preview)
			Expect(body).To(ContainSubstring("Preview environment is live"))
			Expect(body).ToNot(ContainSubstring("Deployed commit"))
		})

		It("notes that URLs are still being provisioned when none have landed", func() {
			preview.Status.Outputs.URLs = nil
			preview.Status.Outputs.URLsPending = true
			body := renderPreviewComment(preview)
			Expect(body).To(ContainSubstring("Preview environment is live"))
			Expect(body).To(ContainSubstring("`abc123`"))
			Expect(body).To(ContainSubstring(urlsPendingNote))
		})

		It("drops the pending note once the URLs land", func() {
			body := renderPreviewComment(preview)
			Expect(body).To(ContainSubstring("https://pr-7-web.example.dev"))
			Expect(body).ToNot(ContainSubstring(urlsPendingNote))
		})
	})
})
