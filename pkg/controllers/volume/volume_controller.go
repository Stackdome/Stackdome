package volume

// Package volume contains the implementation of the Volume controller for managing volumes in a Kubernetes cluster.

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/controllers"
	apperrors "github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	storagev1alpha1 "stackdome.io/cluster-agent/api/storage/v1alpha1"
)

const (
	controllerName = "volume-controller"
)

type volumeReconciler struct {
	Client         client.Client
	Log            logger.Logger
	StorageService services.StackStorageService
	VolumeService  services.VolumeService
	Env            string
}

type VolumeReconcilerSpec struct {
	Log            logger.Logger
	StorageService services.StackStorageService
	VolumeService  services.VolumeService
	Env            string
}

func NewVolumeReconciler(spec VolumeReconcilerSpec) *volumeReconciler {
	return &volumeReconciler{
		Log:            spec.Log,
		StorageService: spec.StorageService,
		VolumeService:  spec.VolumeService,
		Env:            spec.Env,
	}
}

func (w *volumeReconciler) AddToManager(manager manager.Manager) error {
	w.Client = manager.GetClient()
	controller, err := controller.New(controllerName, manager, controller.Options{
		Reconciler: w,
	})
	if err != nil {
		return err
	}
	src := source.Kind(
		manager.GetCache(),
		&storagev1alpha1.Volume{},
		&handler.TypedEnqueueRequestForObject[*storagev1alpha1.Volume]{},
		controllers.VolumeIdLabelPresentPredicate[*storagev1alpha1.Volume](),
	)

	return controller.Watch(src)
}

func (w *volumeReconciler) Name() string {
	return controllerName
}

func (r *volumeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Infof("reconciling volume: %s in namespace %s", req.Name, req.Namespace)
	clusterInstance := &storagev1alpha1.Volume{}
	if err := r.Client.Get(ctx, req.NamespacedName, clusterInstance); err != nil {
		r.Log.Errorf("failed to get volume from cluster: %v", err)
		return ctrl.Result{}, nil
	}

	volumeID, ok := clusterInstance.Labels[models.VolumeIDLabel]
	if !ok {
		r.Log.Errorf("storage ID not found in volumeCR labels")
		return ctrl.Result{}, nil
	}

	dbInstance, serr := r.VolumeService.Get(ctx, volumeID)
	if serr != nil {
		if serr.Code == apperrors.ErrorNotFound {
			r.Log.Infof("volume %s in namespace %s not found in DB", clusterInstance.Name, clusterInstance.Namespace)
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get volume from db: %v", serr)
	}

	// status changed
	if dbInstance.Status == nil || dbInstance.Status.LastObservedStatusHash != clusterInstance.Status.StatusHash {
		dbInstance.Status = mapToVolumeStatus(clusterInstance.Status)
		serr := r.VolumeService.UpdateStatus(ctx, dbInstance.ID, dbInstance.Status)
		if serr != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update  volume status in db: %v", serr)
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func mapToVolumeStatus(clusterStatus storagev1alpha1.VolumeStatus) *models.VolumeStatus {
	return &models.VolumeStatus{
		ObservedGeneration:     clusterStatus.ObservedGeneration,
		Conditions:             models.ConvertConditions(clusterStatus.Conditions),
		Phase:                  string(clusterStatus.Phase),
		BuildArtifactSyncs:     mapToBuildArtifactSyncInfo(clusterStatus.BuildArtifactSyncs),
		LastObservedStatusHash: clusterStatus.StatusHash,
		LastSyncedGitRevision:  clusterStatus.LastSyncedGitReference,
		LastRemoteDirSyncHash:  clusterStatus.LastRemoteSyncHash,
	}
}

func mapToBuildArtifactSyncInfo(clusterBuildArtifactSyncInfoMap map[string]storagev1alpha1.BuildArtifactSyncInfo) []models.BuildArtifactSyncInfo {
	res := make([]models.BuildArtifactSyncInfo, 0)
	for resourceRef, syncInfo := range clusterBuildArtifactSyncInfoMap {
		res = append(res, models.BuildArtifactSyncInfo{
			ResourceName: resourceRef,
			BuildID:      syncInfo.BuildID,
			Status:       string(syncInfo.Status),
		})
	}
	return res
}
