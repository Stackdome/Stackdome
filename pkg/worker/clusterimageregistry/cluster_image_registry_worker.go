package clusterimageregistry

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services/clusterresource"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/worker"
)

const ClusterImageRegistryWorkerName = "cluster-image-registry-worker"

type ClusterImageRegistryWorkerSpec struct {
	ClusterStore       stores.ClusterStore
	ImageRegistryStore stores.ClusterImageRegistryStore
	ClusterManager     clustermanager.ClusterManager
	ClusterResource    clusterresource.ClusterImageRegistryService
	Env                string
}

type clusterImageRegistryWorker struct {
	clusterStore       stores.ClusterStore
	imageRegistryStore stores.ClusterImageRegistryStore
	clusterManager     clustermanager.ClusterManager
	clusterResource    clusterresource.ClusterImageRegistryService
	worker.BaseWorker
}

var _ worker.Worker = (*clusterImageRegistryWorker)(nil)
var _ worker.PeriodicReconcilable = (*clusterImageRegistryWorker)(nil)

func NewClusterImageRegistryWorker(spec ClusterImageRegistryWorkerSpec) worker.Worker {
	return &clusterImageRegistryWorker{
		clusterStore:       spec.ClusterStore,
		imageRegistryStore: spec.ImageRegistryStore,
		clusterManager:     spec.ClusterManager,
		clusterResource:    spec.ClusterResource,
		BaseWorker:         worker.NewBaseWorker(ClusterImageRegistryWorkerName, spec.Env),
	}
}

func (w *clusterImageRegistryWorker) Interval() time.Duration {
	return 30 * time.Second
}

func (w *clusterImageRegistryWorker) GetInput(ctx context.Context) ([]worker.Operand, *errors.ServiceError) {
	clusterIDs, err := w.clusterStore.ListIDsForImageRegistryReconciliation(ctx)
	if err != nil {
		return nil, err
	}
	operands := make([]worker.Operand, len(clusterIDs))
	for index, clusterID := range clusterIDs {
		operands[index] = models.ClusterImageRegistryOperand{ClusterID: clusterID}
	}
	return operands, nil
}

func (w *clusterImageRegistryWorker) Execute(ctx context.Context, operand worker.Operand) (worker.Result, *errors.ServiceError) {
	request, ok := operand.(models.ClusterImageRegistryOperand)
	if !ok || request.ClusterID == "" {
		return worker.Result{}, w.WorkerError.NewError("invalid operand, expected models.ClusterImageRegistryOperand with a cluster ID")
	}

	cluster, err := w.clusterStore.Get(ctx, request.ClusterID)
	if err != nil {
		if err.Is404() {
			return worker.Result{}, nil
		}
		return worker.Result{}, err
	}
	registries, err := w.imageRegistryStore.ListByClusterIDInternal(ctx, cluster.ID)
	if err != nil {
		return worker.Result{}, err
	}

	deletingCluster := cluster.DeletionTimestamp != nil
	registryDeletionPending := false
	for _, registry := range registries {
		if registry == nil {
			return worker.Result{}, w.WorkerError.NewError("cluster %q has a nil image registry", cluster.ID)
		}
		deletingRegistry := deletingCluster || (registry.Status != nil && registry.Status.State == models.RegistryStateDeleting)
		if deletingRegistry {
			absent, resourceErr := w.clusterResource.DeleteImageRegistryInCluster(ctx, registry)
			if resourceErr != nil {
				return worker.Result{}, w.WorkerError.NewError("delete image registry %q: %v", registry.ID, resourceErr)
			}
			if !absent {
				registryDeletionPending = true
				continue
			}
			if deleteErr := w.imageRegistryStore.Delete(ctx, registry.ID); deleteErr != nil {
				return worker.Result{}, deleteErr
			}
			continue
		}
		// Shared-compute registries are seeded in the database at signup and
		// created lazily by the first release that needs one. A cluster-wide
		// reconciliation must not provision every tenant's pending seed.
		if cluster.SharedCompute {
			continue
		}

		if registry.Status != nil && registry.Status.State == models.RegistryStateRunning {
			continue
		}
		if resourceErr := w.clusterResource.EnsureImageRegistryInCluster(ctx, registry); resourceErr != nil {
			status := &models.ClusterImageRegistryStatus{State: models.RegistryStateError, Conditions: []models.Condition{}}
			if statusErr := w.imageRegistryStore.UpdateStatus(ctx, registry.ID, status); statusErr != nil {
				return worker.Result{}, statusErr
			}
			return worker.Result{}, w.WorkerError.NewError("ensure image registry %q: %v", registry.ID, resourceErr)
		}
		if registry.Status != nil && registry.Status.State == models.RegistryStateError {
			status := &models.ClusterImageRegistryStatus{State: models.RegistryStatePending, Conditions: []models.Condition{}}
			if statusErr := w.imageRegistryStore.UpdateStatus(ctx, registry.ID, status); statusErr != nil {
				return worker.Result{}, statusErr
			}
		}
	}

	if registryDeletionPending {
		return worker.Result{RequeueAfter: 2 * time.Second}, nil
	}
	if !deletingCluster {
		return worker.Result{}, nil
	}
	if err := w.clusterManager.UnregisterCluster(cluster.ID); err != nil {
		return worker.Result{}, w.WorkerError.NewError("unregister cluster %q: %v", cluster.ID, err)
	}
	if err := w.clusterStore.Delete(ctx, cluster.ID); err != nil {
		return worker.Result{}, err
	}
	return worker.Result{}, nil
}
