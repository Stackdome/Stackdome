package preview

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
)

type deprovisionReconciler struct {
	previewStackStore previewStackStore
	stackService      stackService
	commentService    previewCommentService
	stackfileCache    *sync.Map
	previewCacheKeys  *sync.Map
	logger            logger.Logger
}

func newDeprovisionReconciler(spec PreviewWorkerSpec, stackfileCache, previewCacheKeys *sync.Map) *deprovisionReconciler {
	return &deprovisionReconciler{
		previewStackStore: spec.PreviewStackStore,
		stackService:      spec.StackService,
		commentService:    spec.CommentService,
		stackfileCache:    stackfileCache,
		previewCacheKeys:  previewCacheKeys,
		logger:            logger.NewLoggerWithPrefix(context.Background(), "preview-deprovision"),
	}
}

func (r *deprovisionReconciler) Name() string { return "deprovision" }

func (r *deprovisionReconciler) Reconcile(ctx context.Context, preview *models.PreviewStack) (subReconcilerResult, error) {
	if preview.DeletionTimestamp == nil {
		return resultNil, nil
	}

	if preview.StackID != nil {
		stack, sErr := r.stackService.InternalGetStack(ctx, *preview.StackID)
		if sErr != nil {
			if sErr.Is404() {
				return r.deletePreviewRecord(ctx, preview)
			}
			return resultNil, fmt.Errorf("failed to get stack %s: %w", *preview.StackID, sErr)
		}

		_, sErr = r.stackService.InternalDeleteStack(ctx, stack)
		if sErr != nil {
			if sErr.Is404() {
				return r.deletePreviewRecord(ctx, preview)
			}
			return resultNil, fmt.Errorf("failed to delete stack %s: %w", *preview.StackID, sErr)
		}

		// Stack deletion is async — requeue to wait for it to be fully removed.
		return resultRequeueAfter(10 * time.Second), nil
	}

	// No stack was ever created — just clean up the preview record.
	return r.deletePreviewRecord(ctx, preview)
}

func (r *deprovisionReconciler) deletePreviewRecord(ctx context.Context, preview *models.PreviewStack) (subReconcilerResult, error) {
	if key, ok := r.previewCacheKeys.LoadAndDelete(preview.ID); ok {
		r.stackfileCache.Delete(key.(string))
	}
	r.logger.Info(ctx, "preview %s: cleaning up record", preview.ID)
	if err := r.commentService.InternalUpsertComment(ctx, preview); err != nil {
		r.logger.Warn(ctx, "preview %s: failed to mark PR comment deleted: %v", preview.ID, err)
	}
	if sErr := r.previewStackStore.Delete(ctx, preview.ID); sErr != nil {
		return resultNil, fmt.Errorf("failed to delete preview record %s: %w", preview.ID, sErr)
	}
	r.logger.Info(ctx, "preview %s deprovisioned", preview.ID)
	return resultStop, nil
}
