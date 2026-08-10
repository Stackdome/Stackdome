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
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	ClusterWrites    *worker.ClusterMutationCoordinator
	ReferenceService referenceService
}

type volumeWorker struct {
	volumeService    volumeService
	stackService     stackService
	stackVolumeStore stackVolumeStore
	clusterManager   clustermanager.ClusterManager
	volumeCRbuilder  builders.ClusterResourceBuilder
	runtimePolicy    services.RuntimePolicy
	releaseService   releaseService
	referenceService referenceService
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
		referenceService: spec.ReferenceService,
		BaseWorker: worker.NewBaseWorkerWithClusterMutationCoordinator(
			VolumeWorkerName, spec.Env, spec.ClusterWrites,
		),
	}
}

func (w *volumeWorker) Execute(ctx context.Context, operand worker.Operand) (worker.Result, *errors.ServiceError) {
	volumeRef, ok := operand.(models.VolumeOperand)
	if !ok {
		return worker.Result{}, w.WorkerError.NewError("invalid operand type, expected models.VolumeOperand")
	}
	unlock := w.LockResource(volumeRef.ID)
	defer unlock()

	var stack *models.Stack
	var vol *models.Volume
	var releaseID string
	if w.runtimePolicy.DraftProvisioningMode() == services.ProvisioningModeDatabaseOnly {
		var authorized bool
		var resolveErr *errors.ServiceError
		vol, stack, releaseID, authorized, resolveErr = w.resolveCloudDesiredVolume(ctx, volumeRef)
		if resolveErr != nil {
			return worker.Result{}, resolveErr
		}
		if !authorized {
			return worker.Result{}, nil
		}
		if admissionErr := w.runtimePolicy.RequireComputeAccess(ctx, stack.OrganisationID); admissionErr != nil {
			if admissionErr.Reason == errors.ErrorCodeComputeAccessInactive {
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

	if releaseID != "" {
		if err := models.ValidatePinnedVolumeGitRevisions(models.StackSnapshot{Volumes: []*models.Volume{vol}}); err != nil {
			w.Logger().Info(ctx, "skipping incompatible release volume %s: %v", vol.ID, err)
			return worker.Result{}, nil
		}
	} else if err := w.resolveGitRevision(ctx, vol); err != nil {
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
	unlockCluster := w.LockClusterNamespace(stack.ClusterID, stack.Namespace)
	defer unlockCluster()

	clusterClient, cerr := w.clusterManager.GetClient(stack.ClusterID)
	if cerr != nil {
		return worker.Result{}, w.WorkerError.NewError("failed to get cluster client for volume '%s': %v", vol.ID, cerr)
	}

	volumeCR, buildErr := w.volumeCRbuilder.BuildVolumeCR(ctx, vol)
	if buildErr != nil {
		return worker.Result{}, w.WorkerError.NewError("failed to build volume CR for '%s': %v", vol.ID, buildErr)
	}
	if releaseID != "" {
		setVolumeReleaseID(volumeCR, releaseID)
		authorized, authorizationErr := w.releaseRemainsAuthoritative(ctx, releaseID, stack.ID)
		if authorizationErr != nil {
			return worker.Result{}, authorizationErr
		}
		if !authorized {
			return worker.Result{}, nil
		}
	}

	existingVolumeCR := &storagev1alpha1.Volume{}
	if getErr := clusterClient.Get(ctx, client.ObjectKey{Name: volumeCR.Name, Namespace: volumeCR.Namespace}, existingVolumeCR); getErr != nil {
		if k8sapierrors.IsNotFound(getErr) {
			if releaseID != "" {
				authorized, authorizationErr := w.releaseRemainsAuthoritative(ctx, releaseID, stack.ID)
				if authorizationErr != nil {
					return worker.Result{}, authorizationErr
				}
				if !authorized {
					return worker.Result{}, nil
				}
			}
			if createErr := clusterClient.Create(ctx, volumeCR); createErr != nil {
				return worker.Result{}, w.WorkerError.NewError("failed to create volume CR for '%s': %v", vol.ID, createErr)
			}
			return worker.Result{}, nil
		}
		return worker.Result{}, w.WorkerError.NewError("failed to get volume CR for '%s': %v", vol.ID, getErr)
	}

	metadataChanged := mergeDesiredMetadata(&existingVolumeCR.Labels, volumeCR.Labels)
	metadataChanged = mergeDesiredMetadata(&existingVolumeCR.Annotations, volumeCR.Annotations) || metadataChanged
	if !apiequality.Semantic.DeepEqual(existingVolumeCR.Spec, volumeCR.Spec) || metadataChanged {
		existingVolumeCR.Spec = volumeCR.Spec
		if releaseID != "" {
			authorized, authorizationErr := w.releaseRemainsAuthoritative(ctx, releaseID, stack.ID)
			if authorizationErr != nil {
				return worker.Result{}, authorizationErr
			}
			if !authorized {
				return worker.Result{}, nil
			}
		}
		if updateErr := clusterClient.Update(ctx, existingVolumeCR); updateErr != nil {
			return worker.Result{}, w.WorkerError.NewError("failed to update volume CR for '%s': %v", vol.ID, updateErr)
		}
	}

	return worker.Result{}, nil
}

func mergeDesiredMetadata(existing *map[string]string, desired map[string]string) bool {
	if *existing == nil {
		*existing = make(map[string]string, len(desired))
	}
	changed := false
	for key, value := range desired {
		if (*existing)[key] == value {
			continue
		}
		(*existing)[key] = value
		changed = true
	}
	return changed
}

// resolveCloudDesiredVolume returns the immutable volume spec from the release
// that authorizes this reconciliation. Periodic operands without a release ID
// may only reconcile the stack's last converged release.
func (w *volumeWorker) resolveCloudDesiredVolume(ctx context.Context, ref models.VolumeOperand) (*models.Volume, *models.Stack, string, bool, *errors.ServiceError) {
	releaseID := ref.ReleaseID
	var stack *models.Stack
	if releaseID == "" {
		current, serr := w.volumeService.InternalGet(ctx, ref.ID)
		if serr != nil {
			if serr.Is404() {
				return nil, nil, "", false, nil
			}
			return nil, nil, "", false, serr
		}
		var err error
		stack, err = w.resolveStack(ctx, current.ID)
		if err != nil {
			return nil, nil, "", false, w.WorkerError.NewError("failed to resolve stack for volume '%s': %v", current.ID, err)
		}
		releaseID = stack.GetConvergedReleaseID()
		if releaseID == "" {
			return nil, nil, "", false, nil
		}
	}

	release, serr := w.releaseService.InternalGet(ctx, releaseID)
	if serr != nil {
		return nil, nil, "", false, serr
	}
	if stack == nil {
		stack, serr = w.stackService.InternalGetStack(ctx, release.StackID)
		if serr != nil {
			return nil, nil, "", false, serr
		}
	}
	authoritativeRelease, authorizationErr := w.releaseService.InternalResolveAuthoritativeWorkloadRelease(ctx, stack)
	if authorizationErr != nil {
		return nil, nil, "", false, authorizationErr
	}
	if authoritativeRelease == nil || authoritativeRelease.ID != release.ID {
		return nil, nil, "", false, nil
	}
	for _, volume := range release.Snapshot.Volumes {
		if volume == nil || volume.ID != ref.ID {
			continue
		}
		if volume.OrganisationID != stack.OrganisationID || volume.ProjectID != stack.ProjectID ||
			volume.NamespaceID != stack.NamespaceID || volume.Namespace != stack.Namespace {
			return nil, nil, "", false, w.WorkerError.NewError("release %s does not authorize volume %s", release.ID, ref.ID)
		}
		return volume, stack, release.ID, true, nil
	}

	if ref.ReleaseID != "" {
		return nil, nil, "", false, w.WorkerError.NewError("release %s does not contain volume %s", release.ID, ref.ID)
	}
	return nil, nil, "", false, nil
}

func (w *volumeWorker) releaseRemainsAuthoritative(ctx context.Context, releaseID, stackID string) (bool, *errors.ServiceError) {
	stack, serr := w.stackService.InternalGetStack(ctx, stackID)
	if serr != nil {
		return false, serr
	}
	if stack.DeletionTimestamp != nil {
		return false, nil
	}
	authoritativeRelease, serr := w.releaseService.InternalResolveAuthoritativeWorkloadRelease(ctx, stack)
	if serr != nil {
		return false, serr
	}
	if authoritativeRelease == nil || authoritativeRelease.ID != releaseID {
		return false, nil
	}
	if admissionErr := w.runtimePolicy.RequireComputeAccess(ctx, stack.OrganisationID); admissionErr != nil {
		if admissionErr.Reason == errors.ErrorCodeComputeAccessInactive {
			return false, nil
		}
		return false, admissionErr
	}
	return true, nil
}

func setVolumeReleaseID(volume *storagev1alpha1.Volume, releaseID string) {
	if volume.Annotations == nil {
		volume.Annotations = make(map[string]string)
	}
	volume.Annotations[models.VolumeReleaseIDAnnotation] = releaseID
}

func (w *volumeWorker) Interval() time.Duration {
	return 30 * time.Second
}

func (w *volumeWorker) GetInput(ctx context.Context) ([]worker.Operand, *errors.ServiceError) {
	if w.runtimePolicy.DraftProvisioningMode() == services.ProvisioningModeDatabaseOnly {
		return w.getCloudInput(ctx)
	}
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

func (w *volumeWorker) getCloudInput(ctx context.Context) ([]worker.Operand, *errors.ServiceError) {
	scan, serr := w.releaseService.InternalListAuthoritativeWorkload(ctx)
	if serr != nil {
		return nil, w.WorkerError.NewError("failed to list authoritative volume releases: %v", serr)
	}
	releaseIDs := make([]string, len(scan.Releases))
	for i, release := range scan.Releases {
		releaseIDs[i] = release.ReleaseID
	}
	refs, serr := w.referenceService.InternalListReleaseReferents(ctx, releaseIDs, models.ReferentVolume)
	if serr != nil {
		return nil, w.WorkerError.NewError("failed to list authoritative volume references: %v", serr)
	}
	operands := make([]worker.Operand, 0, len(refs))
	seen := make(map[string]struct{}, len(refs))
	for _, ref := range refs {
		if ref.ReleaseID == nil {
			continue
		}
		key := ref.ReferentID + "\x00" + *ref.ReleaseID
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		operands = append(operands, models.VolumeOperand{ID: ref.ReferentID, ReleaseID: *ref.ReleaseID})
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
