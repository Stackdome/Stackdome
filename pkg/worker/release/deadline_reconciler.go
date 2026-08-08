package release

import (
	"context"
	"fmt"
	"time"

	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
)

type deadlineReconciler struct {
	releaseService    releaseService
	imageBuildService imageBuildService
	logger            logger.Logger
}

func newDeadlineReconciler(spec ReleaseWorkerSpec) *deadlineReconciler {
	return &deadlineReconciler{
		releaseService:    spec.ReleaseService,
		imageBuildService: spec.ImageBuildService,
		logger:            logger.NewLoggerWithPrefix(context.Background(), "release-deadline"),
	}
}

func (r *deadlineReconciler) Name() string { return "deadline" }

// Times out a release that is stuck. Two separate clocks:
//
//   - Lifetime cap (releaseLifetimeCap, from CreatedAt): absolute ceiling.
//     Catches everything the converge clock cannot see — a build stuck
//     Pending forever, a build CR that never appears, a dead agent.
//
//   - Converge clock (convergenceTimeout, from WorkerStatus stamp): the
//     deploy/converge phase only. The clock is stamped when the release has
//     no builds left to wait for, and cleared again while a build is
//     running, so a slow build never eats the deploy phase's window.
//
// Build failures are NOT handled here: a Failed or Cancelled build fails the
// release immediately via the converge reconciler's fail-fast.
//
// Runs after the gatekeeper so supersede/CAS decisions happen first.
func (r *deadlineReconciler) Reconcile(ctx context.Context, release *models.StackRelease) (subReconcilerResult, error) {
	if release.State != models.ReleaseStateInProgress {
		return resultNil, nil
	}

	if time.Since(release.CreatedAt) > releaseLifetimeCap {
		return r.fail(ctx, release,
			fmt.Sprintf("release exceeded maximum lifetime of %d hours", int(releaseLifetimeCap.Hours())))
	}

	phase, err := r.buildPhase(ctx, release)
	if err != nil {
		return resultNil, fmt.Errorf("failed to check builds for deadline: %w", err)
	}

	switch phase {
	case buildsRunning:
		// Can't converge-timeout mid-build. Clear a stale stamp so the
		// deploy phase gets a fresh window once the build finishes.
		if release.ConvergeClockStartedAt() != nil {
			if serr := r.releaseService.SetConvergeClockStartedAt(ctx, release.ID, nil); serr != nil {
				return resultNil, fmt.Errorf("failed to clear converge clock: %w", serr)
			}
		}
		return resultNil, nil

	case buildsNotAppearedYet:
		// The stack builds from source but the agent has not created the
		// build CR yet. Starting the clock now would count build time
		// against the deploy window. The lifetime cap covers a build that
		// never appears.
		return resultNil, nil

	case buildsDone:
		startedAt := release.ConvergeClockStartedAt()
		if startedAt == nil {
			now := time.Now().UTC()
			if serr := r.releaseService.SetConvergeClockStartedAt(ctx, release.ID, &now); serr != nil {
				return resultNil, fmt.Errorf("failed to start converge clock: %w", serr)
			}
			return resultNil, nil
		}
		if time.Since(*startedAt) <= convergenceTimeout {
			return resultNil, nil
		}
		msg := fmt.Sprintf("release did not converge within %d minutes", int(convergenceTimeout.Minutes()))
		if releaseExpectsBuilds(release) {
			msg += " of build completion"
		}
		return r.fail(ctx, release, msg)
	}
	return resultNil, nil
}

func (r *deadlineReconciler) fail(ctx context.Context, release *models.StackRelease, msg string) (subReconcilerResult, error) {
	if _, serr := r.releaseService.MarkFailed(ctx, release.ID, msg, nil); serr != nil {
		return resultNil, fmt.Errorf("failed to mark release failed on deadline: %w", serr)
	}
	r.logger.Info(ctx, "release %s: %s", release.ID, msg)
	return resultStop, nil
}

// releaseBuildPhase summarises where this release's builds are, as far as the
// deadline clocks care.
type releaseBuildPhase int

const (
	// buildsDone: nothing to wait for — no builds expected, or every build
	// attributed to this release is terminal. The converge clock may run.
	buildsDone releaseBuildPhase = iota
	// buildsRunning: at least one attributed build is not terminal yet.
	buildsRunning
	// buildsNotAppearedYet: the snapshot says the stack builds from source,
	// but no build attributed to this release exists yet.
	buildsNotAppearedYet
)

// buildPhase classifies this release's builds. Attribution is by
// Status.ReleaseID, same as the converge reconciler's fail-fast. A build in
// an unknown state counts as running: better to wait than to cancel a
// release mid-build.
func (r *deadlineReconciler) buildPhase(ctx context.Context, release *models.StackRelease) (releaseBuildPhase, error) {
	if !releaseExpectsBuilds(release) {
		return buildsDone, nil
	}

	builds, serr := r.imageBuildService.ListByStackID(ctx, release.StackID)
	if serr != nil {
		return buildsDone, serr
	}

	seen := false
	for _, b := range builds {
		if b.Status == nil || b.Status.ReleaseID != release.ID {
			continue
		}
		seen = true
		switch b.Status.State {
		case string(buildsv1alpha1.BuildPhaseSuccess),
			string(buildsv1alpha1.BuildPhaseFailed),
			string(buildsv1alpha1.BuildPhaseCancelled):
			continue
		}
		return buildsRunning, nil
	}
	if !seen {
		return buildsNotAppearedYet, nil
	}
	return buildsDone, nil
}

// releaseExpectsBuilds reports whether any resource in the release snapshot
// builds from source (and therefore produces an ImageBuild).
func releaseExpectsBuilds(release *models.StackRelease) bool {
	for _, res := range release.Snapshot.Resources {
		if res.BuildConfig != nil {
			return true
		}
	}
	return false
}
