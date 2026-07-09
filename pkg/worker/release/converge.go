package release

import (
	"context"
	"fmt"
	"time"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
)

// releaseLiveMessage is the user-facing message recorded on the
// release_released terminal event.
const releaseLiveMessage = "Release is live"

type convergeReconciler struct {
	releaseService releaseService
	stackService   stackService
	eventRecorder  eventRecorder
	logger         logger.Logger
}

func newConvergeReconciler(spec ReleaseWorkerSpec) *convergeReconciler {
	return &convergeReconciler{
		releaseService: spec.ReleaseService,
		stackService:   spec.StackService,
		eventRecorder:  spec.EventRecorder,
		logger:         logger.NewLoggerWithPrefix(context.Background(), "release-converge"),
	}
}

func (r *convergeReconciler) Name() string { return "converge" }

func (r *convergeReconciler) Reconcile(ctx context.Context, release *models.StackRelease) (subReconcilerResult, error) {
	if release.Manifest == nil {
		return resultNil, nil
	}

	latest, serr := r.stackService.InternalGetStack(ctx, release.StackID)
	if serr != nil {
		if serr.Is404() {
			failRelease(ctx, r.releaseService, r.eventRecorder, r.logger, release, "stack not found")
			return resultStop, nil
		}
		return resultNil, fmt.Errorf("failed to get stack: %w", serr)
	}

	if latest.Status != nil && latest.Status.LastConverged != nil {
		converged := latest.Status.LastConverged
		if converged.ReleaseID == release.ID && converged.Revision == release.ManifestRevision {
			outcome := buildOutcome(latest, release)
			ok, markErr := r.releaseService.MarkReleased(ctx, release.ID, outcome)
			if markErr != nil {
				return resultNil, fmt.Errorf("failed to mark released: %w", markErr)
			}
			if ok {
				r.logger.Infof("release %s converged", release.ID)
				release.State = models.ReleaseStateReleased
				if recErr := r.eventRecorder.RecordReleaseTerminal(ctx, release, models.ReleaseStateReleased, releaseLiveMessage); recErr != nil {
					r.logger.Errorf("release %s: failed to record release_released event: %v", release.ID, recErr)
				}
			}
			return resultStop, nil
		}
	}

	deployTimeout := time.Duration(latest.EffectiveSettings().DeployTimeoutMinutes) * time.Minute
	if release.RenderedAt != nil && time.Since(*release.RenderedAt) > deployTimeout {
		msg := fmt.Sprintf("timed out waiting for convergence after %s", deployTimeout)
		outcome := buildOutcome(latest, release)
		won, markErr := r.releaseService.MarkFailed(ctx, release.ID, msg, &outcome)
		if markErr != nil {
			return resultNil, fmt.Errorf("failed to mark failed: %w", markErr)
		}
		r.logger.Errorf("release %s: %s", release.ID, msg)
		if won {
			release.State = models.ReleaseStateFailed
			if recErr := r.eventRecorder.RecordReleaseTerminal(ctx, release, models.ReleaseStateFailed, msg); recErr != nil {
				r.logger.Errorf("release %s: failed to record release_failed event: %v", release.ID, recErr)
			}
		}
		return resultStop, nil
	}

	return resultRequeueAfter(convergencePollInterval), nil
}

func buildOutcome(stack *models.Stack, release *models.StackRelease) models.ReleaseOutcome {
	outcome := models.ReleaseOutcome{
		Resources: make(map[string]models.ResourceOutcome),
	}
	if release.RenderedAt != nil {
		outcome.Duration = time.Since(*release.RenderedAt)
	}
	if stack.Status == nil {
		return outcome
	}
	for _, rs := range stack.Status.Resources {
		outcome.Resources[rs.Name] = models.ResourceOutcome{
			Phase:         rs.Phase,
			ReadyReplicas: rs.AvailableReplicas,
			Replicas:      rs.Replicas,
			Message:       rs.Message,
		}
	}
	return outcome
}
