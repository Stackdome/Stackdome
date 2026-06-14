package release

import (
	"context"
	"fmt"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/worker"
)

// checkConverged inspects the stack's live status to determine whether the
// release has converged. The stack controller watches CRs and writes status
// back to the DB, including LastConverged which tells us when the cluster
// matched a given revision.
func (w *releaseWorker) checkConverged(ctx context.Context, release *models.StackRelease, stack *models.Stack) (worker.Result, *errors.ServiceError) {
	// Re-fetch the stack to get the latest status.
	latest, serr := w.stackService.InternalGetStack(ctx, release.StackID)
	if serr != nil {
		return worker.Result{}, serr
	}

	if latest.Status != nil && latest.Status.LastConverged != nil {
		converged := latest.Status.LastConverged
		if converged.ReleaseID == release.ID || converged.Revision == release.ManifestRevision {
			outcome := buildOutcome(latest, release)
			if _, markErr := w.releaseService.MarkReleased(ctx, release.ID, outcome); markErr != nil {
				return worker.Result{}, markErr
			}
			w.Logger().Infof("release %s converged", release.ID)
			return worker.Result{}, nil
		}
	}

	// Check for timeout.
	if release.RenderedAt != nil && time.Since(*release.RenderedAt) > convergenceTimeout {
		msg := fmt.Sprintf("timed out waiting for convergence after %s", convergenceTimeout)
		outcome := buildOutcome(latest, release)
		if _, markErr := w.releaseService.MarkFailed(ctx, release.ID, msg, &outcome); markErr != nil {
			return worker.Result{}, markErr
		}
		w.Logger().Errorf("release %s: %s", release.ID, msg)
		return worker.Result{}, nil
	}

	// Not converged yet; requeue.
	return worker.Result{RequeueAfter: convergencePollInterval}, nil
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
