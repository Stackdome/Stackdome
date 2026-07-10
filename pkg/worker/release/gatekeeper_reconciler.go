package release

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
)

// supersededEventMessageFmt is the user-facing message recorded on the
// release_superseded terminal event (distinct from the internal CAS reason).
const supersededEventMessageFmt = "Release superseded by release #%d"

type gatekeeperReconciler struct {
	releaseService releaseService
	eventRecorder  eventRecorder
	logger         logger.Logger
}

func newGatekeeperReconciler(spec ReleaseWorkerSpec) *gatekeeperReconciler {
	return &gatekeeperReconciler{
		releaseService: spec.ReleaseService,
		eventRecorder:  spec.EventRecorder,
		logger:         logger.NewLoggerWithPrefix(context.Background(), "release-gatekeeper"),
	}
}

func (r *gatekeeperReconciler) Name() string { return "gatekeeper" }

func (r *gatekeeperReconciler) Reconcile(ctx context.Context, release *models.StackRelease) (subReconcilerResult, error) {
	latest, serr := r.releaseService.InternalGetActiveByStackID(ctx, release.StackID)
	if serr != nil {
		return resultNil, fmt.Errorf("failed to get active release: %w", serr)
	}
	if latest != nil && latest.ID != release.ID && latest.Sequence > release.Sequence {
		reason := fmt.Sprintf(supersededEventMessageFmt, latest.Sequence)
		if _, err := r.releaseService.MarkSuperseded(ctx, release.ID, reason); err != nil {
			return resultNil, fmt.Errorf("failed to mark release superseded: %w", err)
		}
		r.logger.Infof("release %s: %s", release.ID, reason)
		return resultStop, nil
	}

	if release.State == models.ReleaseStatePending {
		ok, serr := r.releaseService.MarkInProgress(ctx, release.ID)
		if serr != nil {
			return resultNil, fmt.Errorf("failed to mark in progress: %w", serr)
		}
		if !ok {
			r.logger.Infof("release %s: CAS Pending→InProgress failed", release.ID)
			return resultStop, nil
		}
		release.State = models.ReleaseStateInProgress
		if recErr := r.eventRecorder.RecordReleaseStarted(ctx, release); recErr != nil {
			r.logger.Errorf("release %s: failed to record release_started event: %v", release.ID, recErr)
		}
	}

	return resultNil, nil
}
