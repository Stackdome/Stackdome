package preview

import (
	"context"
	"fmt"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type deprovisionReconciler struct {
	previewStackStore previewStackStore
	stackService      stackService
	logger            logger.Logger
}

func newDeprovisionReconciler(spec PreviewWorkerSpec) *deprovisionReconciler {
	return &deprovisionReconciler{
		previewStackStore: spec.PreviewStackStore,
		stackService:      spec.StackService,
		logger:            logger.NewLoggerWithPrefix(context.Background(), "preview-deprovision"),
	}
}

func (r *deprovisionReconciler) Name() string { return "deprovision" }

func (r *deprovisionReconciler) Reconcile(ctx context.Context, preview *models.PreviewStack) (subReconcilerResult, error) {
	if preview.Status.Phase != models.PreviewStackPhaseDeleting {
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
	r.logger.Infof("preview %s: stack already deleted, cleaning up record", preview.ID)
	if sErr := r.previewStackStore.Delete(ctx, preview.ID); sErr != nil {
		return resultNil, fmt.Errorf("failed to delete preview record %s: %w", preview.ID, sErr)
	}
	r.logger.Infof("preview %s deprovisioned", preview.ID)
	return resultStop, nil
}
