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
		if sErr != nil && !sErr.Is404() {
			return resultNil, fmt.Errorf("failed to get stack %s: %w", *preview.StackID, sErr)
		}
		_, sErr = r.stackService.InternalDeleteStack(ctx, stack)
		if sErr != nil && !sErr.Is404() {
			return resultNil, fmt.Errorf("failed to delete stack %s: %w", *preview.StackID, sErr)
		}
		if sErr.Is404() {
			r.logger.Infof("preview %s: stack %s already deleted", preview.ID, *preview.StackID)
			if sErr := r.previewStackStore.Delete(ctx, preview.ID); sErr != nil && !sErr.Is404() {
				return resultNil, fmt.Errorf("failed to delete preview record %s: %w", preview.ID, sErr)
			}
			r.logger.Infof("preview %s deprovisioned", preview.ID)
			return resultStop, nil
		}
		return resultRequeueAfter(10 * time.Second), nil
	}

	return resultStop, nil
}
