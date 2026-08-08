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

var _ = Describe("previewStackService InternalCreateFromWebhook", func() {
	var (
		ctrl     *gomock.Controller
		store    *mocks.MockPreviewStackStore
		enqueuer *mocks.MockBackgroundJobEnqueuer
		svc      *previewStackService
		ctx      context.Context
		config   *models.StackPreviewConfig
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		store = mocks.NewMockPreviewStackStore(ctrl)
		enqueuer = mocks.NewMockBackgroundJobEnqueuer(ctrl)
		svc = &previewStackService{
			store: store,
			BackgroundJobEnqueuerDep: BackgroundJobEnqueuerDep{
				BackgroundJobEnqueuer: enqueuer,
			},
		}
		ctx = context.Background()

		config = &models.StackPreviewConfig{
			ID:                "config-1",
			OrganisationID:    "org-1",
			ProjectID:         "project-1",
			UserID:            "user-1",
			Name:              "web",
			MaxActivePreviews: 2,
			GitRepository:     models.PreviewGitRepository{RepoURL: "https://github.com/acme/app.git", BaseBranch: "main"},
		}
	})

	It("provisions a preview sourced from the webhook using the supplied head SHA", func() {
		store.EXPECT().GetByConfigAndPR(ctx, "config-1", "7").Return(nil, errors.NotFound("not found"))
		store.EXPECT().CountActiveByConfigID(ctx, "config-1").Return(int64(0), nil)
		store.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, preview *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
				Expect(preview.Source).To(Equal(models.PreviewStackSourceWebhook))
				Expect(preview.CommitSHA).To(Equal("abc123"))
				Expect(preview.PRNumber).To(Equal("7"))
				Expect(preview.Branch).To(Equal("feature"))
				Expect(preview.UserID).To(Equal("user-1"))
				preview.ID = "preview-1"
				return preview, nil
			},
		)
		enqueuer.EXPECT().Enqueue(models.PreviewStackOperand{ID: "preview-1"}).Return(nil)

		created, sErr := svc.InternalCreateFromWebhook(ctx, config, "7", "feature", "abc123")
		Expect(sErr).To(BeNil())
		Expect(created.ID).To(Equal("preview-1"))
	})

	It("rejects creation once the config's active preview cap is reached", func() {
		store.EXPECT().GetByConfigAndPR(ctx, "config-1", "7").Return(nil, errors.NotFound("not found"))
		store.EXPECT().CountActiveByConfigID(ctx, "config-1").Return(int64(2), nil)

		_, sErr := svc.InternalCreateFromWebhook(ctx, config, "7", "feature", "abc123")
		Expect(sErr).ToNot(BeNil())
		Expect(sErr.Code).To(Equal(errors.ErrorTooManyRequests))
	})
})
