package preview

import (
	"context"
	"sync"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"k8s.io/utils/ptr"
)

var _ = Describe("DeprovisionReconciler", func() {
	var (
		ctrl         *gomock.Controller
		previewStore *MockpreviewStackStore
		stackSvc     *MockstackService
		reconciler   *deprovisionReconciler
		ctx          context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		previewStore = NewMockpreviewStackStore(ctrl)
		stackSvc = NewMockstackService(ctrl)
		ctx = context.Background()

		cache := &sync.Map{}
		cacheKeys := &sync.Map{}
		reconciler = &deprovisionReconciler{
			previewStackStore: previewStore,
			stackService:      stackSvc,
			stackfileCache:    cache,
			previewCacheKeys:  cacheKeys,
			logger:            logger.NewLoggerWithPrefix(ctx, "test"),
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("noop", func() {
		It("returns resultNil when DeletionTimestamp is nil", func() {
			preview := &models.PreviewStack{
				ID:     "p-1",
				Status: models.PreviewStackStatus{Phase: models.PreviewStackPhaseReady},
			}
			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultNil))
		})
	})

	Context("when StackID is nil", func() {
		It("hard-deletes the preview record and returns resultStop", func() {
			preview := &models.PreviewStack{
				ID:                "p-2",
				DeletionTimestamp: ptr.To(time.Now()),
				Status:            models.PreviewStackStatus{Phase: models.PreviewStackPhaseDeleting},
			}
			previewStore.EXPECT().Delete(gomock.Any(), "p-2").Return(nil)

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultStop))
		})

		It("evicts stackfile cache entry when preview record is deleted", func() {
			preview := &models.PreviewStack{
				ID:                "p-cache",
				DeletionTimestamp: ptr.To(time.Now()),
				Status:            models.PreviewStackStatus{Phase: models.PreviewStackPhaseDeleting},
			}

			// Pre-populate cache
			cacheKey := "https://github.com/test/repo:abc123:stackfile.yaml"
			reconciler.stackfileCache.Store(cacheKey, "cached-content")
			reconciler.previewCacheKeys.Store("p-cache", cacheKey)

			previewStore.EXPECT().Delete(gomock.Any(), "p-cache").Return(nil)

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultStop))

			// Verify cache was evicted
			_, cacheHit := reconciler.stackfileCache.Load(cacheKey)
			Expect(cacheHit).To(BeFalse())
			_, keyHit := reconciler.previewCacheKeys.Load("p-cache")
			Expect(keyHit).To(BeFalse())
		})
	})

	Context("when stack exists", func() {
		var (
			preview *models.PreviewStack
			stackID string
		)

		BeforeEach(func() {
			stackID = "stack-1"
			preview = &models.PreviewStack{
				ID:                "p-3",
				StackID:           &stackID,
				DeletionTimestamp: ptr.To(time.Now()),
				Status:            models.PreviewStackStatus{Phase: models.PreviewStackPhaseDeleting},
			}
		})

		It("deletes the stack and requeues when delete succeeds with 404", func() {
			stack := &models.Stack{ID: stackID, Name: "preview-stack"}

			stackSvc.EXPECT().InternalGetStack(gomock.Any(), stackID).Return(stack, nil)
			stackSvc.EXPECT().InternalDeleteStack(gomock.Any(), stack).
				Return(nil, errors.NotFound("stack already deleted"))
			previewStore.EXPECT().Delete(gomock.Any(), "p-3").Return(nil)

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultStop))
		})
	})

	Context("when stack is already deleted (404 on GetStack)", func() {
		It("hard-deletes the preview record and returns resultStop", func() {
			stackID := "stack-gone"
			preview := &models.PreviewStack{
				ID:                "p-4",
				StackID:           &stackID,
				DeletionTimestamp: ptr.To(time.Now()),
				Status:            models.PreviewStackStatus{Phase: models.PreviewStackPhaseDeleting},
			}

			stackSvc.EXPECT().InternalGetStack(gomock.Any(), stackID).
				Return(nil, errors.NotFound("stack not found"))
			previewStore.EXPECT().Delete(gomock.Any(), "p-4").Return(nil)

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultStop))
		})
	})

	Context("when InternalGetStack returns transient error", func() {
		It("returns error for retry", func() {
			stackID := "stack-err"
			preview := &models.PreviewStack{
				ID:                "p-5",
				StackID:           &stackID,
				DeletionTimestamp: ptr.To(time.Now()),
				Status:            models.PreviewStackStatus{Phase: models.PreviewStackPhaseDeleting},
			}

			stackSvc.EXPECT().InternalGetStack(gomock.Any(), stackID).
				Return(nil, errors.GeneralError("connection refused"))

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get stack"))
			Expect(result).To(Equal(resultNil))
		})
	})

	Context("when InternalDeleteStack returns non-404 error", func() {
		It("returns error for retry", func() {
			stackID := "stack-del-err"
			preview := &models.PreviewStack{
				ID:                "p-6",
				StackID:           &stackID,
				DeletionTimestamp: ptr.To(time.Now()),
				Status:            models.PreviewStackStatus{Phase: models.PreviewStackPhaseDeleting},
			}
			stack := &models.Stack{ID: stackID, Name: "preview-stack"}

			stackSvc.EXPECT().InternalGetStack(gomock.Any(), stackID).Return(stack, nil)
			stackSvc.EXPECT().InternalDeleteStack(gomock.Any(), stack).
				Return(nil, errors.GeneralError("delete operation failed"))

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to delete stack"))
			Expect(result).To(Equal(resultNil))
		})
	})

	It("returns Name() as deprovision", func() {
		Expect(reconciler.Name()).To(Equal("deprovision"))
	})

	Describe("requeue after delete", func() {
		It("requeues after 10 seconds when delete is in progress", func() {
			preview := &models.PreviewStack{
				ID:                "p-5",
				DeletionTimestamp: ptr.To(time.Now()),
				Status:            models.PreviewStackStatus{Phase: models.PreviewStackPhaseDeleting},
			}
			stackID := "stack-5"
			preview.StackID = &stackID

			stack := &models.Stack{ID: stackID}
			stackSvc.EXPECT().InternalGetStack(gomock.Any(), stackID).Return(stack, nil)
			stackSvc.EXPECT().InternalDeleteStack(gomock.Any(), stack).Return(stack, nil)

			result, err := reconciler.Reconcile(context.Background(), preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.resultRequeueAfter).ToNot(BeNil())
			Expect(*result.resultRequeueAfter).To(Equal(10 * time.Second))
		})
	})
})
