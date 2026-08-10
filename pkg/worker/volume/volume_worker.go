package volume

import (
	"context"
	"fmt"
	"time"

	"github.com/Stackdome/stackdome/pkg/builders"
	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/Stackdome/stackdome/pkg/worker"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
	storagev1alpha1 "stackdome.io/cluster-agent/api/storage/v1alpha1"
)

const VolumeWorkerName = "volume-worker"

type VolumeWorkerSpec struct {
	VolumeService    volumeService
	StackService     stackService
	StackVolumeStore stackVolumeStore
	ClusterManager   clustermanager.ClusterManager
	VolumeCrBuilder  builders.ClusterResourceBuilder
	Env              string
	RuntimePolicy    services.RuntimePolicy
	ReleaseService   releaseService
}

type volumeWorker struct {
	volumeService    volumeService
	stackService     stackService
	stackVolumeStore stackVolumeStore
	clusterManager   clustermanager.ClusterManager
	volumeCRbuilder  builders.ClusterResourceBuilder
	runtimePolicy    services.RuntimePolicy
	releaseService   releaseService
	worker.BaseWorker
}

var _ worker.Worker = (*volumeWorker)(nil)

func NewVolumeWorker(spec VolumeWorkerSpec) worker.Worker {
	return &volumeWorker{
		volumeService:    spec.VolumeService,
		stackService:     spec.StackService,
		stackVolumeStore: spec.StackVolumeStore,
		clusterManager:   spec.ClusterManager,
		volumeCRbuilder:  spec.VolumeCrBuilder,
		runtimePolicy:    spec.RuntimePolicy,
		releaseService:   spec.ReleaseService,
		BaseWorker:       worker.NewBaseWorker(VolumeWorkerName, spec.Env),
	}
}

func (w *volumeWorker) Execute(ctx context.Context, operand worker.Operand) (worker.Result, *errors.ServiceError) {
	volumeRef, ok := operand.(models.VolumeOperand)
	if !ok {
		return worker.Result{}, w.WorkerError.NewError("invalid operand type, expected models.VolumeOperand")
	}

	var stack *models.Stack
	var vol *models.Volume
	if w.runtimePolicy.DraftProvisioningMode() == services.ProvisioningModeDatabaseOnly {
		var authorized bool
		var resolveErr *errors.ServiceError
		vol, stack, authorized, resolveErr = w.resolveCloudDesiredVolume(ctx, volumeRef)
		if resolveErr != nil {
			return worker.Result{}, resolveErr
		}
		if !authorized {
			return worker.Result{}, nil
		}
		if admissionErr := w.runtimePolicy.RequireActiveAllocation(ctx, stack.OrganisationID); admissionErr != nil {
			if admissionErr.Reason == errors.ErrorCodeTrialInactive {
				return worker.Result{}, nil
			}
			return worker.Result{}, admissionErr
		}
	} else {
		var serr *errors.ServiceError
		vol, serr = w.volumeService.InternalGet(ctx, volumeRef.ID)
		if serr != nil {
			if serr.Is404() {
				w.Logger().Info(ctx, "volume %s not found, skipping", volumeRef.ID)
				return worker.Result{}, nil
			}
			return worker.Result{}, serr
		}
	}

	w.Logger().Info(ctx, "processing volume: %s", vol.ID)

	if err := w.resolveGitRevision(ctx, vol); err != nil {
		return worker.Result{}, w.WorkerError.NewError("failed to resolve git revision for volume '%s': %v", vol.ID, err)
	}

	// Resolve cluster ID through the stack-volume association.
	if stack == nil {
		resolvedStack, stackErr := w.resolveStack(ctx, vol.ID)
		if stackErr != nil {
			return worker.Result{}, w.WorkerError.NewError("failed to resolve cluster for volume '%s': %v", vol.ID, stackErr)
		}
		stack = resolvedStack
	}

	clusterClient, cerr := w.clusterManager.GetClient(stack.ClusterID)
	if cerr != nil {
		return worker.Result{}, w.WorkerError.NewError("failed to get cluster client for volume '%s': %v", vol.ID, cerr)
	}

	volumeCR, buildErr := w.volumeCRbuilder.BuildVolumeCR(ctx, vol)
	if buildErr != nil {
		return worker.Result{}, w.WorkerError.NewError("failed to build volume CR for '%s': %v", vol.ID, buildErr)
	}

	existingVolumeCR := &storagev1alpha1.Volume{}
	if getErr := clusterClient.Get(ctx, client.ObjectKey{Name: volumeCR.Name, Namespace: volumeCR.Namespace}, existingVolumeCR); getErr != nil {
		if k8sapierrors.IsNotFound(getErr) {
			if createErr := clusterClient.Create(ctx, volumeCR); createErr != nil {
				return worker.Result{}, w.WorkerError.NewError("failed to create volume CR for '%s': %v", vol.ID, createErr)
			}
			return worker.Result{}, nil
		}
		return worker.Result{}, w.WorkerError.NewError("failed to get volume CR for '%s': %v", vol.ID, getErr)
	}

	// Update source revisions on existing CRs.
	if vol.VolumeSource == nil {
		return worker.Result{}, nil
	}

	if vol.VolumeSource.GitRepoSource != nil {
		updatedRevision := buildGitRevision(vol.VolumeSource.GitRepoSource.Revision)
		if existingVolumeCR.Spec.Source.GitRepo.Revision != updatedRevision {
			existingVolumeCR.Spec.Source.GitRepo.Revision = updatedRevision
			if updateErr := clusterClient.Update(ctx, existingVolumeCR); updateErr != nil {
				return worker.Result{}, w.WorkerError.NewError("failed to update volume CR git revision for '%s': %v", vol.ID, updateErr)
			}
		}
		return worker.Result{}, nil
	}

	if vol.VolumeSource.RemoteDirSource != nil {
		if existingVolumeCR.Spec.Source.RemoteDir.CurrentDirectoryHash != vol.VolumeSource.RemoteDirSource.CurrentDirectoryHash {
			existingVolumeCR.Spec.Source.RemoteDir.CurrentDirectoryHash = vol.VolumeSource.RemoteDirSource.CurrentDirectoryHash
			if updateErr := clusterClient.Update(ctx, existingVolumeCR); updateErr != nil {
				return worker.Result{}, w.WorkerError.NewError("failed to update volume CR remote dir hash for '%s': %v", vol.ID, updateErr)
			}
		}
	}

	return worker.Result{}, nil
}

// resolveCloudDesiredVolume returns the immutable volume spec from the release
// that authorizes this reconciliation. Periodic operands without a release ID
// may only reconcile the stack's last converged release.
func (w *volumeWorker) resolveCloudDesiredVolume(ctx context.Context, ref models.VolumeOperand) (*models.Volume, *models.Stack, bool, *errors.ServiceError) {
	releaseID := ref.ReleaseID
	expectedStackID := ""
	if releaseID == "" {
		current, serr := w.volumeService.InternalGet(ctx, ref.ID)
		if serr != nil {
			if serr.Is404() {
				return nil, nil, false, nil
			}
			return nil, nil, false, serr
		}
		stack, err := w.resolveStack(ctx, current.ID)
		if err != nil {
			return nil, nil, false, w.WorkerError.NewError("failed to resolve stack for volume '%s': %v", current.ID, err)
		}
		releaseID = stack.GetConvergedReleaseID()
		if releaseID == "" {
			return nil, nil, false, nil
		}
		expectedStackID = stack.ID
	}

	release, serr := w.releaseService.InternalGet(ctx, releaseID)
	if serr != nil {
		return nil, nil, false, serr
	}
	if !release.State.AllowsWorkloadReconciliation() {
		return nil, nil, false, nil
	}
	stack := release.Snapshot.ToStack()
	if release.StackID != stack.ID || stack.OrganisationID == "" || stack.ClusterID == "" {
		return nil, nil, false, w.WorkerError.NewError("release %s has an invalid stack identity", release.ID)
	}
	if expectedStackID != "" && release.StackID != expectedStackID {
		return nil, nil, false, w.WorkerError.NewError("release %s does not authorize stack %s", release.ID, expectedStackID)
	}
	for _, volume := range release.Snapshot.Volumes {
		if volume == nil || volume.ID != ref.ID {
			continue
		}
		if volume.OrganisationID != stack.OrganisationID || volume.ProjectID != stack.ProjectID ||
			volume.NamespaceID != stack.NamespaceID || volume.Namespace != stack.Namespace {
			return nil, nil, false, w.WorkerError.NewError("release %s does not authorize volume %s", release.ID, ref.ID)
		}
		return volume, stack, true, nil
	}

	if ref.ReleaseID != "" {
		return nil, nil, false, w.WorkerError.NewError("release %s does not contain volume %s", release.ID, ref.ID)
	}
	return nil, nil, false, nil
}

func (w *volumeWorker) Interval() time.Duration {
	return 30 * time.Second
}

func (w *volumeWorker) GetInput(ctx context.Context) ([]worker.Operand, *errors.ServiceError) {
	volumes, err := w.volumeService.InternalListNotReady(ctx)
	if err != nil {
		return nil, w.WorkerError.NewError("failed to list not-ready volumes: %v", err)
	}

	operands := make([]worker.Operand, len(volumes))
	for i, vol := range volumes {
		operands[i] = models.VolumeOperand{ID: vol.ID}
	}
	return operands, nil
}

func (w *volumeWorker) resolveStack(ctx context.Context, volumeID string) (*models.Stack, error) {
	sv, serr := w.stackVolumeStore.GetByVolumeID(ctx, volumeID)
	if serr != nil {
		return nil, fmt.Errorf("failed to find stack for volume '%s': %w", volumeID, serr)
	}
	stack, serr := w.stackService.InternalGetStack(ctx, sv.StackID)
	if serr != nil {
		return nil, fmt.Errorf("failed to get stack '%s': %w", sv.StackID, serr)
	}
	return stack, nil
}

func (w *volumeWorker) resolveGitRevision(ctx context.Context, vol *models.Volume) error {
	if vol.VolumeSource == nil || vol.VolumeSource.GitRepoSource == nil {
		return nil
	}

	src := vol.VolumeSource.GitRepoSource
	client, err := gitclient.NewGitClientForRepo(src.RepoUrl, gitclient.GitCredentials{})
	if err != nil {
		return fmt.Errorf("create git client: %w", err)
	}

	resolved, err := gitclient.ResolveGitRepoRevision(ctx, client, src.RepoUrl, src.Revision)
	if err != nil {
		return err
	}
	vol.VolumeSource.GitRepoSource.Revision = resolved
	return nil
}

func buildGitRevision(rev models.GitRepoRevision) corev1alpha1.GitRepoRevision {
	result := corev1alpha1.GitRepoRevision{}
	switch rev.Type() {
	case models.Branch:
		result.Branch = rev.Branch
		if rev.Commit != "" {
			result.Commit = rev.Commit
		}
	case models.Tag:
		result.Tag = rev.Tag
		if rev.Commit != "" {
			result.Commit = rev.Commit
		}
	case models.Commit:
		result.Commit = rev.Commit
	}
	return result
}
