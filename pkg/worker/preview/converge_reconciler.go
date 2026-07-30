package preview

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
)

type convergeReconciler struct {
	releaseService    releaseService
	stackService      stackService
	previewStackStore previewStackStore
	commentService    previewCommentService
	logger            logger.Logger
}

func newConvergeReconciler(spec PreviewWorkerSpec) *convergeReconciler {
	return &convergeReconciler{
		releaseService:    spec.ReleaseService,
		stackService:      spec.StackService,
		previewStackStore: spec.PreviewStackStore,
		commentService:    spec.CommentService,
		logger:            logger.NewLoggerWithPrefix(context.Background(), "preview-converge"),
	}
}

func (r *convergeReconciler) Name() string { return "converge" }

func (r *convergeReconciler) Reconcile(ctx context.Context, preview *models.PreviewStack) (subReconcilerResult, error) {
	if preview.ActiveReleaseID == nil {
		return resultNil, nil
	}

	release, sErr := r.releaseService.InternalGet(ctx, *preview.ActiveReleaseID)
	if sErr != nil {
		return resultNil, fmt.Errorf("failed to get release %s: %w", *preview.ActiveReleaseID, sErr)
	}

	switch release.State {
	case models.ReleaseStateReleased:
		if preview.Status.Phase == models.PreviewStackPhaseReady {
			return r.retryPendingComment(ctx, preview)
		}
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
		preview.GitHubCommentPending = !r.upsertComment(ctx, preview)
		if _, sErr := r.previewStackStore.Update(ctx, preview); sErr != nil {
			return resultNil, fmt.Errorf("failed to update preview status: %w", sErr)
		}

		r.logger.Info(ctx, "preview %s is ready", preview.ID)
		return resultStop, nil

	case models.ReleaseStateSuperseded:
		// A newer sync triggered a new release that superseded this one.
		// Requeue — the provision reconciler will update ActiveReleaseID.
		return resultRequeueAfter(convergePollInterval), nil

	case models.ReleaseStateFailed:
		preview.Status = models.PreviewStackStatus{
			Phase:   models.PreviewStackPhaseFailed,
			Reason:  "ReleaseFailed",
			Message: release.Message,
			Outputs: preview.Status.Outputs, // keep last-known URLs for the failed comment
		}
		preview.GitHubCommentPending = !r.upsertComment(ctx, preview)
		if _, sErr := r.previewStackStore.Update(ctx, preview); sErr != nil {
			return resultNil, fmt.Errorf("failed to update preview status: %w", sErr)
		}
		r.logger.Info(ctx, "preview %s failed: %s", preview.ID, release.Message)
		return resultStop, nil

	case models.ReleaseStateCancelled:
		preview.Status = models.PreviewStackStatus{
			Phase:   models.PreviewStackPhaseFailed,
			Reason:  "ReleaseCancelled",
			Message: "release was cancelled",
		}
		if _, sErr := r.previewStackStore.Update(ctx, preview); sErr != nil {
			return resultNil, fmt.Errorf("failed to update preview status: %w", sErr)
		}
		r.logger.Info(ctx, "preview %s cancelled", preview.ID)
		return resultStop, nil

	default:
		return resultRequeueAfter(convergePollInterval), nil
	}
}

// upsertComment posts the terminal-state comment; a failure is logged and
// retried on a later pass via GitHubCommentPending.
func (r *convergeReconciler) upsertComment(ctx context.Context, preview *models.PreviewStack) (posted bool) {
	if err := r.commentService.InternalUpsertComment(ctx, preview); err != nil {
		r.logger.Warn(ctx, "preview %s: failed to upsert PR comment, will retry: %v", preview.ID, err)
		return false
	}
	return true
}

// retryPendingComment re-attempts a PR comment that failed after the preview
// already went Ready. Status is already persisted; only the flag changes.
func (r *convergeReconciler) retryPendingComment(ctx context.Context, preview *models.PreviewStack) (subReconcilerResult, error) {
	if !preview.GitHubCommentPending {
		return resultNil, nil
	}
	if !r.upsertComment(ctx, preview) {
		return resultStop, nil // still failing; the periodic scan retries
	}
	preview.GitHubCommentPending = false
	if _, sErr := r.previewStackStore.Update(ctx, preview); sErr != nil {
		return resultNil, fmt.Errorf("failed to update preview status: %w", sErr)
	}
	return resultStop, nil
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
