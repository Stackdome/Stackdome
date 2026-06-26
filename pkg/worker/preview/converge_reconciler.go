package preview

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type convergeReconciler struct {
	releaseService    releaseService
	stackService      stackService
	previewStackStore previewStackStore
	logger            logger.Logger
}

func newConvergeReconciler(spec PreviewWorkerSpec) *convergeReconciler {
	return &convergeReconciler{
		releaseService:    spec.ReleaseService,
		stackService:      spec.StackService,
		previewStackStore: spec.PreviewStackStore,
		logger:            logger.NewLoggerWithPrefix(context.Background(), "preview-converge"),
	}
}

func (r *convergeReconciler) Name() string { return "converge" }

func (r *convergeReconciler) Reconcile(ctx context.Context, preview *models.PreviewStack) (subReconcilerResult, error) {
	if preview.Status.Phase != models.PreviewStackPhaseDeploying {
		return resultNil, nil
	}

	if preview.ActiveReleaseID == nil {
		// How did this happen? We should have a release by now.
		r.logger.Errorf("preview %s: no active release found", preview.ID)
		return resultStop, nil
	}

	release, sErr := r.releaseService.InternalGet(ctx, *preview.ActiveReleaseID)
	if sErr != nil {
		return resultNil, fmt.Errorf("failed to get release %s: %w", *preview.ActiveReleaseID, sErr)
	}

	switch {
	case release.State == models.ReleaseStateReleased:
		stack, sErr := r.stackService.InternalGetStack(ctx, *preview.StackID)
		if sErr != nil {
			return resultNil, fmt.Errorf("failed to get stack: %w", sErr)
		}

		outputs := buildOutputs(stack, preview.CommitSHA)
		preview.Status = models.PreviewStackStatus{
			Phase:   models.PreviewStackPhaseReady,
			Reason:  "ReleaseConverged",
			Outputs: &outputs,
		}
		if _, sErr := r.previewStackStore.Update(ctx, preview); sErr != nil {
			return resultNil, fmt.Errorf("failed to update preview status: %w", sErr)
		}
		r.logger.Infof("preview %s is ready", preview.ID)
		return resultStop, nil

	case release.State == models.ReleaseStateSuperseded:
		// A newer sync triggered a new release that superseded this one.
		// Requeue — the provision reconciler will update ActiveReleaseID.
		return resultRequeueAfter(convergePollInterval), nil

	case release.State == models.ReleaseStateFailed:
		preview.Status = models.PreviewStackStatus{
			Phase:   models.PreviewStackPhaseFailed,
			Reason:  "ReleaseFailed",
			Message: release.Message,
		}
		if _, sErr := r.previewStackStore.Update(ctx, preview); sErr != nil {
			return resultNil, fmt.Errorf("failed to update preview status: %w", sErr)
		}
		r.logger.Infof("preview %s failed: %s", preview.ID, release.Message)
		return resultStop, nil

	case release.State == models.ReleaseStateCancelled:
		preview.Status = models.PreviewStackStatus{
			Phase:   models.PreviewStackPhaseFailed,
			Reason:  "ReleaseCancelled",
			Message: "release was cancelled",
		}
		if _, sErr := r.previewStackStore.Update(ctx, preview); sErr != nil {
			return resultNil, fmt.Errorf("failed to update preview status: %w", sErr)
		}
		r.logger.Infof("preview %s cancelled", preview.ID)
		return resultStop, nil

	default:
		return resultRequeueAfter(convergePollInterval), nil
	}
}

func buildOutputs(stack *models.Stack, commitSHA string) models.PreviewStackOutputs {
	outputs := models.PreviewStackOutputs{
		CommitSHA: commitSHA,
	}
	for _, res := range stack.StackResources {
		if res.Status == nil {
			continue
		}
		for _, ingress := range res.Status.PublicIngresses {
			outputs.URLs = append(outputs.URLs, models.PreviewURL{
				Resource: res.Name,
				URL:      ingress.URL,
			})
		}
	}
	return outputs
}
