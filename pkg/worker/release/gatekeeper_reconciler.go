package release

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type gatekeeperReconciler struct {
	releaseService releaseService
	logger         logger.Logger
}

func newGatekeeperReconciler(spec ReleaseWorkerSpec) *gatekeeperReconciler {
	return &gatekeeperReconciler{
		releaseService: spec.ReleaseService,
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
		reason := fmt.Sprintf("superseded by release #%d", latest.Sequence)
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
	}

	return resultNil, nil
}
