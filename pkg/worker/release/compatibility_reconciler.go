package release

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
)

const incompatibleReleaseMessage = "release is incompatible with cloud reconciliation"

type compatibilityReconciler struct {
	runtimePolicy  runtimePolicy
	releaseService releaseService
}

func newCompatibilityReconciler(spec ReleaseWorkerSpec) *compatibilityReconciler {
	return &compatibilityReconciler{
		runtimePolicy:  spec.RuntimePolicy,
		releaseService: spec.ReleaseService,
	}
}

func (*compatibilityReconciler) Name() string { return "compatibility" }

func (r *compatibilityReconciler) Reconcile(ctx context.Context, release *models.StackRelease) (subReconcilerResult, error) {
	if r.runtimePolicy.DraftProvisioningMode() == services.ProvisioningModeEager {
		return resultNil, nil
	}
	if err := models.ValidatePinnedVolumeGitRevisions(release.Snapshot); err != nil {
		message := fmt.Sprintf("%s: %v", incompatibleReleaseMessage, err)
		updated, serr := r.releaseService.MarkFailed(ctx, release.ID, message, nil)
		if serr != nil {
			return resultNil, fmt.Errorf("mark incompatible release failed: %w", serr)
		}
		if updated {
			release.State = models.ReleaseStateFailed
			release.Message = message
		}
		return resultStop, nil
	}
	return resultNil, nil
}
