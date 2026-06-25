package preview

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type provisionReconciler struct {
	previewStackService previewStackService
	previewStackStore   previewStackStore
	configStore         configStore
	stackService        stackService
	releaseService      releaseService
	logger              logger.Logger
}

func newProvisionReconciler(spec PreviewWorkerSpec) *provisionReconciler {
	return &provisionReconciler{
		previewStackService: spec.PreviewStackService,
		previewStackStore:   spec.PreviewStackStore,
		configStore:         spec.ConfigStore,
		stackService:        spec.StackService,
		releaseService:      spec.ReleaseService,
		logger:              logger.NewLoggerWithPrefix(context.Background(), "preview-provision"),
	}
}

func (r *provisionReconciler) Name() string { return "provision" }

func (r *provisionReconciler) Reconcile(ctx context.Context, preview *models.PreviewStack) (subReconcilerResult, error) {
	if preview.Status.Phase != models.PreviewStackPhaseProvisioning {
		return resultNil, nil
	}

	config, sErr := r.configStore.GetByID(ctx, preview.StackPreviewConfigID)
	if sErr != nil {
		return resultNil, fmt.Errorf("failed to get config: %w", sErr)
	}

	content, hash, opErr := r.previewStackService.InternalFetchStackfile(ctx, config, preview.CommitSHA)
	if opErr != nil {
		if errors.IsRetryable(opErr) {
			return resultNil, opErr
		}
		return r.fail(ctx, preview, opErr.Reason, opErr.Message)
	}

	if preview.StackID != nil {
		// Sync path: stack already exists
		needsUpdate := hash != preview.StackfileHash || len(preview.ImageOverrides) > 0
		var model *models.Stack
		if needsUpdate {
			var buildErr *errors.OperationError
			model, buildErr = r.previewStackService.InternalBuildStackFromContent(ctx, config, preview, content)
			if buildErr != nil {
				if errors.IsRetryable(buildErr) {
					return resultNil, buildErr
				}
				return r.fail(ctx, preview, buildErr.Reason, buildErr.Message)
			}
		}

		if txErr := r.previewStackStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
			if needsUpdate {
				if _, sErr := r.stackService.InternalUpdateStack(txCtx, *preview.StackID, model); sErr != nil {
					return sErr
				}
			}

			release, sErr := r.releaseService.InternalCreateRelease(txCtx, *preview.StackID, models.ReleaseCause{
				Kind:   models.ReleaseCausePreviewSync,
				Detail: fmt.Sprintf("PR #%s synced at %s", preview.PRNumber, preview.CommitSHA[:7]),
			})
			if sErr != nil {
				return sErr
			}

			preview.StackfileHash = hash
			preview.ActiveReleaseID = &release.ID
			preview.Status = models.PreviewStackStatus{Phase: models.PreviewStackPhaseDeploying, Reason: "SyncTriggered"}
			if _, sErr := r.previewStackStore.Update(txCtx, preview); sErr != nil {
				return sErr
			}
			return nil
		}); txErr != nil {
			return resultNil, fmt.Errorf("sync transaction failed: %w", txErr)
		}

		return resultRequeueAfter(convergePollInterval), nil
	}

	// Create path: no stack yet
	model, opErr := r.previewStackService.InternalBuildStackFromContent(ctx, config, preview, content)
	if opErr != nil {
		if errors.IsRetryable(opErr) {
			return resultNil, opErr
		}
		return r.fail(ctx, preview, opErr.Reason, opErr.Message)
	}

	var created *models.Stack
	if txErr := r.previewStackStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		var sErr *errors.ServiceError
		created, sErr = r.stackService.InternalCreateStack(txCtx, model)
		if sErr != nil {
			return sErr
		}

		release, sErr := r.releaseService.InternalCreateRelease(txCtx, created.ID, models.ReleaseCause{
			Kind:   models.ReleaseCausePreviewSync,
			Detail: fmt.Sprintf("PR #%s opened at %s", preview.PRNumber, preview.CommitSHA[:7]),
		})
		if sErr != nil {
			return sErr
		}

		preview.StackID = &created.ID
		preview.ActiveReleaseID = &release.ID
		preview.StackfileHash = hash
		preview.Status = models.PreviewStackStatus{Phase: models.PreviewStackPhaseDeploying, Reason: "StackCreated"}
		if _, sErr := r.previewStackStore.Update(txCtx, preview); sErr != nil {
			return sErr
		}
		return nil
	}); txErr != nil {
		return resultNil, fmt.Errorf("provision transaction failed: %w", txErr)
	}

	r.logger.Infof("preview %s provisioned, stack %s created", preview.ID, created.ID)
	return resultRequeueAfter(convergePollInterval), nil
}

func (r *provisionReconciler) fail(ctx context.Context, preview *models.PreviewStack, reason, message string) (subReconcilerResult, error) {
	preview.Status = models.PreviewStackStatus{
		Phase:   models.PreviewStackPhaseFailed,
		Reason:  reason,
		Message: message,
	}
	if _, sErr := r.previewStackStore.Update(ctx, preview); sErr != nil {
		r.logger.Errorf("failed to update preview %s status: %v", preview.ID, sErr)
	}
	r.logger.Errorf("preview %s failed: %s: %s", preview.ID, reason, message)
	return resultStop, nil
}
