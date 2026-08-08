package release

import (
	"context"
	"fmt"
	"time"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
)

type deadlineReconciler struct {
	releaseService releaseService
	logger         logger.Logger
}

func newDeadlineReconciler(spec ReleaseWorkerSpec) *deadlineReconciler {
	return &deadlineReconciler{
		releaseService: spec.ReleaseService,
		logger:         logger.NewLoggerWithPrefix(context.Background(), "release-deadline"),
	}
}

func (r *deadlineReconciler) Name() string { return "deadline" }

// Fails a release still in progress convergenceTimeout after creation.
// Runs after the gatekeeper so supersede/CAS decisions happen first; covers
// every later phase (render, apply, converge), not just convergence polling.
func (r *deadlineReconciler) Reconcile(ctx context.Context, release *models.StackRelease) (subReconcilerResult, error) {
	if release.State != models.ReleaseStateInProgress {
		return resultNil, nil
	}
	if time.Since(release.CreatedAt) <= convergenceTimeout {
		return resultNil, nil
	}

	msg := fmt.Sprintf("release did not converge within %d minutes", int(convergenceTimeout.Minutes()))
	if _, err := r.releaseService.MarkFailed(ctx, release.ID, msg, nil); err != nil {
		return resultNil, fmt.Errorf("failed to mark release failed on deadline: %w", err)
	}
	r.logger.Info(ctx, "release %s: %s", release.ID, msg)
	return resultStop, nil
}
