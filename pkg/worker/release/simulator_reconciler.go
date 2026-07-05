package release

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
)

type simulatorReconciler struct {
	releaseService releaseService
	env            string
	logger         logger.Logger
}

func newSimulatorReconciler(spec ReleaseWorkerSpec) *simulatorReconciler {
	return &simulatorReconciler{
		releaseService: spec.ReleaseService,
		env:            spec.Env,
		logger:         logger.NewLoggerWithPrefix(context.Background(), "release-simulator"),
	}
}

func (r *simulatorReconciler) Name() string { return "simulator" }

func (r *simulatorReconciler) Reconcile(ctx context.Context, release *models.StackRelease) (subReconcilerResult, error) {
	annotations := release.Snapshot.Stack.Annotations.ToMap()
	if annotations[models.SkipClusterProvisioningAnnotation] != "true" {
		return resultNil, nil
	}

	if r.env != config.EnvironmentTest {
		r.logger.Errorf("release %s: simulation annotations are only allowed in test environments, ignoring", release.ID)
		return resultNil, nil
	}

	targetState := annotations[models.SimulateReleaseStateAnnotation]
	if targetState == "" {
		targetState = string(models.ReleaseStateReleased)
	}

	switch models.StackReleaseState(targetState) {
	case models.ReleaseStateReleased:
		if _, err := r.releaseService.MarkReleased(ctx, release.ID, models.ReleaseOutcome{}); err != nil {
			return resultNil, fmt.Errorf("simulated Released: %w", err)
		}
	case models.ReleaseStateFailed:
		if _, err := r.releaseService.MarkFailed(ctx, release.ID, "simulated failure", nil); err != nil {
			return resultNil, fmt.Errorf("simulated Failed: %w", err)
		}
	case models.ReleaseStateSuperseded:
		if _, err := r.releaseService.MarkSuperseded(ctx, release.ID, "simulated supersession"); err != nil {
			return resultNil, fmt.Errorf("simulated Superseded: %w", err)
		}
	default:
		return resultNil, fmt.Errorf("unsupported simulated release state: %s", targetState)
	}

	r.logger.Infof("simulated release %s → %s (skip-cluster-provisioning)", release.ID, targetState)
	return resultStop, nil
}
