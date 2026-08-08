package release

import (
	"context"
	"fmt"
	"time"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"
)

type convergeReconciler struct {
	releaseService    releaseService
	stackService      stackService
	imageBuildService imageBuildService
	logger            logger.Logger
}

func newConvergeReconciler(spec ReleaseWorkerSpec) *convergeReconciler {
	return &convergeReconciler{
		releaseService:    spec.ReleaseService,
		stackService:      spec.StackService,
		imageBuildService: spec.ImageBuildService,
		logger:            logger.NewLoggerWithPrefix(context.Background(), "release-converge"),
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
			failRelease(ctx, r.releaseService, r.logger, release, "stack not found")
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
				r.logger.Info(ctx, "release %s converged", release.ID)
			}
			return resultStop, nil
		}
	}

	failedBuild, buildErr := r.failedReleaseBuild(ctx, release)
	if buildErr != nil {
		return resultNil, fmt.Errorf("failed to check builds: %w", buildErr)
	}
	if failedBuild != nil {
		msg := buildFailedMessage(failedBuild)
		outcome := buildOutcome(latest, release)
		if _, markErr := r.releaseService.MarkFailed(ctx, release.ID, msg, &outcome); markErr != nil {
			return resultNil, fmt.Errorf("failed to mark failed: %w", markErr)
		}
		r.logger.Error(ctx, "release %s: %s", release.ID, msg)
		return resultStop, nil
	}

	// Poll until converged, superseded, a failure signal (build failed,
	// apply error), or the deadline reconciler times the release out.
	return resultRequeueAfter(convergencePollInterval), nil
}

// failedReleaseBuild returns a terminally failed or cancelled build that this
// release triggered — Status.ReleaseID carries the release-id annotation the
// agent copies onto the ImageBuild at creation. Such a release can never
// converge: this is the fail-fast that replaced the deploy timeout. A prior
// release's failed build carries that release's ID, so it can never kill
// this one.
//
// ponytail: a release that reuses a prior release's already-failed build
// (identical revision and build config) matches nothing here and polls
// forever; needs the agent to re-stamp reused builds to close.
func (r *convergeReconciler) failedReleaseBuild(ctx context.Context, release *models.StackRelease) (*models.ImageBuild, error) {
	builds, serr := r.imageBuildService.ListByStackID(ctx, release.StackID)
	if serr != nil {
		return nil, serr
	}
	for _, b := range builds {
		if b.Status == nil {
			continue
		}
		switch b.Status.State {
		case string(buildsv1alpha1.BuildPhaseFailed), string(buildsv1alpha1.BuildPhaseCancelled):
		default:
			continue
		}
		if b.Status.ReleaseID != "" && b.Status.ReleaseID == release.ID {
			return b, nil
		}
	}
	return nil, nil
}

func buildFailedMessage(b *models.ImageBuild) string {
	if b.Status.State == string(buildsv1alpha1.BuildPhaseCancelled) {
		return fmt.Sprintf("build cancelled for %s", b.StackResourceName)
	}
	msg := fmt.Sprintf("build failed for %s", b.StackResourceName)
	if f := b.Status.LastBuildFailureDetail; f != nil {
		detail := f.Message
		if detail == "" {
			detail = f.Reason
		}
		if detail != "" {
			msg = fmt.Sprintf("%s: %s", msg, detail)
		}
	}
	return msg
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
