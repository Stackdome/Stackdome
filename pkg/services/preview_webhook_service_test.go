package services

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

var _ = Describe("PreviewWebhookService HandlePullRequest", func() {
	var (
		ctrl         *gomock.Controller
		configStore  *mocks.MockStackPreviewConfigStore
		previewStore *mocks.MockPreviewStackStore
		previewSvc   *MockPreviewStackService
		svc          PreviewWebhookService
		ctx          context.Context
		config       *models.StackPreviewConfig
		ev           PullRequestEvent
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		configStore = mocks.NewMockStackPreviewConfigStore(ctrl)
		previewStore = mocks.NewMockPreviewStackStore(ctrl)
		previewSvc = NewMockPreviewStackService(ctrl)
		svc = NewPreviewWebhookService(PreviewWebhookServiceSpec{
			ConfigStore:       configStore,
			PreviewStackStore: previewStore,
			PreviewService:    previewSvc,
		})
		ctx = context.Background()

		config = &models.StackPreviewConfig{
			ID:             "config-1",
			OrganisationID: "org-1",
			GitRepository:  models.PreviewGitRepository{RepoURL: "https://github.com/acme/app.git", BaseBranch: "main"},
		}
		ev = PullRequestEvent{
			OrganisationID: "org-1",
			RepoURL:        "https://github.com/acme/app.git",
			Branch:         "feature",
			BaseBranch:     "main",
			HeadSHA:        "abc",
			PRNumber:       "7",
		}
	})

	It("creates a preview when opened and none exists", func() {
		ev.Action = PRActionOpened
		configStore.EXPECT().GetByOrgAndRepo(ctx, "org-1", models.NormalizeRepoURL(ev.RepoURL)).Return(config, nil)
		previewStore.EXPECT().GetByConfigAndPR(ctx, "config-1", "7").Return(nil, errors.NotFound("not found"))
		previewSvc.EXPECT().InternalCreateFromWebhook(ctx, config, "7", "feature", "abc").Return(&models.PreviewStack{}, nil)

		Expect(svc.HandlePullRequest(ctx, ev)).To(BeNil())
	})

	It("syncs when opened and a preview already exists", func() {
		ev.Action = PRActionOpened
		existing := &models.PreviewStack{ID: "preview-1"}
		configStore.EXPECT().GetByOrgAndRepo(ctx, "org-1", models.NormalizeRepoURL(ev.RepoURL)).Return(config, nil)
		previewStore.EXPECT().GetByConfigAndPR(ctx, "config-1", "7").Return(existing, nil)
		previewSvc.EXPECT().InternalSyncFromWebhook(ctx, "preview-1", "abc").Return(nil)

		Expect(svc.HandlePullRequest(ctx, ev)).To(BeNil())
	})

	It("syncs on synchronize when a preview already exists", func() {
		ev.Action = PRActionSynchronize
		existing := &models.PreviewStack{ID: "preview-1"}
		configStore.EXPECT().GetByOrgAndRepo(ctx, "org-1", models.NormalizeRepoURL(ev.RepoURL)).Return(config, nil)
		previewStore.EXPECT().GetByConfigAndPR(ctx, "config-1", "7").Return(existing, nil)
		previewSvc.EXPECT().InternalSyncFromWebhook(ctx, "preview-1", "abc").Return(nil)

		Expect(svc.HandlePullRequest(ctx, ev)).To(BeNil())
	})

	It("creates on synchronize when no preview exists yet", func() {
		ev.Action = PRActionSynchronize
		configStore.EXPECT().GetByOrgAndRepo(ctx, "org-1", models.NormalizeRepoURL(ev.RepoURL)).Return(config, nil)
		previewStore.EXPECT().GetByConfigAndPR(ctx, "config-1", "7").Return(nil, errors.NotFound("not found"))
		previewSvc.EXPECT().InternalCreateFromWebhook(ctx, config, "7", "feature", "abc").Return(&models.PreviewStack{}, nil)

		Expect(svc.HandlePullRequest(ctx, ev)).To(BeNil())
	})

	It("deletes the preview when closed", func() {
		ev.Action = PRActionClosed
		existing := &models.PreviewStack{ID: "preview-1"}
		configStore.EXPECT().GetByOrgAndRepo(ctx, "org-1", models.NormalizeRepoURL(ev.RepoURL)).Return(config, nil)
		previewStore.EXPECT().GetByConfigAndPR(ctx, "config-1", "7").Return(existing, nil)
		previewSvc.EXPECT().InternalDeleteFromWebhook(ctx, "preview-1").Return(nil)

		Expect(svc.HandlePullRequest(ctx, ev)).To(BeNil())
	})

	It("does nothing when closed and no preview exists", func() {
		ev.Action = PRActionClosed
		configStore.EXPECT().GetByOrgAndRepo(ctx, "org-1", models.NormalizeRepoURL(ev.RepoURL)).Return(config, nil)
		previewStore.EXPECT().GetByConfigAndPR(ctx, "config-1", "7").Return(nil, errors.NotFound("not found"))

		Expect(svc.HandlePullRequest(ctx, ev)).To(BeNil())
	})

	It("does nothing when the PR base branch does not match the config", func() {
		ev.Action = PRActionOpened
		ev.BaseBranch = "dev"
		configStore.EXPECT().GetByOrgAndRepo(ctx, "org-1", models.NormalizeRepoURL(ev.RepoURL)).Return(config, nil)

		Expect(svc.HandlePullRequest(ctx, ev)).To(BeNil())
	})

	It("does nothing when no config is registered for the repo", func() {
		ev.Action = PRActionOpened
		configStore.EXPECT().GetByOrgAndRepo(ctx, "org-1", models.NormalizeRepoURL(ev.RepoURL)).Return(nil, errors.NotFound("not found"))

		Expect(svc.HandlePullRequest(ctx, ev)).To(BeNil())
	})
})
