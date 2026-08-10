package volume

import (
	"context"
	"fmt"
	"time"

	"github.com/Stackdome/stackdome/pkg/builders"
	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
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
	ClusterWrites    *worker.ClusterMutationCoordinator
}

type volumeWorker struct {
	volumeService    volumeService
	stackService     stackService
	stackVolumeStore stackVolumeStore
	clusterManager   clustermanager.ClusterManager
	volumeCRbuilder  builders.ClusterResourceBuilder
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

	vol, serr := w.volumeService.InternalGet(ctx, volumeRef.ID)
	if serr != nil {
		if serr.Is404() {
			w.Logger().Info(ctx, "volume %s not found, skipping", volumeRef.ID)
			return worker.Result{}, nil
		}
		return worker.Result{}, serr
	}

	w.Logger().Info(ctx, "processing volume: %s", vol.ID)

	// Resolve cluster ID through the stack-volume association.
	stack, stackErr := w.resolveStack(ctx, vol.ID)
	if stackErr != nil {
		return worker.Result{}, w.WorkerError.NewError("failed to resolve cluster for volume '%s': %v", vol.ID, stackErr)
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

	metadataChanged := mergeDesiredMetadata(&existingVolumeCR.Labels, volumeCR.Labels)
	metadataChanged = mergeDesiredMetadata(&existingVolumeCR.Annotations, volumeCR.Annotations) || metadataChanged
	if !apiequality.Semantic.DeepEqual(existingVolumeCR.Spec, volumeCR.Spec) || metadataChanged {
		existingVolumeCR.Spec = volumeCR.Spec
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
