package workspacevolume

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
	workspacev1alpha1 "soradev.io/cluster-agent/api/v1alpha1"
)

type WorkspaceVolumeReconciler struct {
	Client                  client.Client
	Log                     logger.Logger
	WorkspaceStorageService services.WorkspaceStorageService
	VolumeService           services.VolumeService
	Env                     string
}

type WorkspaceVolumeReconcilerSpec struct {
	Client                  client.Client
	Log                     logger.Logger
	WorkspaceStorageService services.WorkspaceStorageService
	VolumeService           services.VolumeService
	Env                     string
}

func NewWorkspaceVolumeReconciler(spec WorkspaceVolumeReconcilerSpec) *WorkspaceVolumeReconciler {
	return &WorkspaceVolumeReconciler{
		Client:                  spec.Client,
		Log:                     spec.Log,
		WorkspaceStorageService: spec.WorkspaceStorageService,
		VolumeService:           spec.VolumeService,
		Env:                     spec.Env,
	}
}

func (w *WorkspaceVolumeReconciler) AddToManager(manager manager.Manager) error {
	controller, err := controller.New("workspace-volume-controller", manager, controller.Options{
		Reconciler: w,
	})
	if err != nil {
		return err
	}
	src := source.Kind(
		manager.GetCache(),
		&workspacev1alpha1.WorkspaceVolume{},
		&handler.TypedEnqueueRequestForObject[*workspacev1alpha1.WorkspaceVolume]{},
		controllers.WorkspaceStorageIdLabelPresentPredicate[*workspacev1alpha1.WorkspaceVolume](),
	)

	return controller.Watch(src)
}

func (r *WorkspaceVolumeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Infof("reconciling workspace volume: %s in namespace %s", req.Name, req.Namespace)
	clusterInstance := &workspacev1alpha1.WorkspaceVolume{}
	if err := r.Client.Get(ctx, req.NamespacedName, clusterInstance); err != nil {
		r.Log.Errorf("failed to get workspace volume from cluster: %v", err)
		return ctrl.Result{}, nil
	}

	workspaceStorageID, ok := clusterInstance.Labels[models.WorkspaceStorageIDLabel]
	if !ok {
		r.Log.Errorf("workspace storage ID not found in workspace volume labels")
		return ctrl.Result{}, nil
	}

	if _, serr := r.WorkspaceStorageService.Get(ctx, workspaceStorageID); serr != nil {
		r.Log.Errorf("failed to get workspace storage from db: %v", serr)
		return ctrl.Result{}, nil
	}

	dbInstance, serr := r.VolumeService.Get(ctx, clusterInstance.Name, workspaceStorageID)
	if serr != nil {
		if serr.Code == apperrors.ErrorNotFound {
			return ctrl.Result{}, r.deleteWorkspaceVolumeFromCluster(ctx, clusterInstance)
		}
		return ctrl.Result{}, fmt.Errorf("failed to get workspace volume from db: %v", serr)
	}

	// status changed
	if dbInstance.VolumeStatus.LastObservedStatusHash != clusterInstance.Status.StatusHash {
		dbInstance.VolumeStatus = mapToVolumeStatus(clusterInstance.Status)
		_, serr := r.VolumeService.Update(ctx, clusterInstance.Name, workspaceStorageID, dbInstance)
		if serr != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update workspace volume in db: %v", serr)
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *WorkspaceVolumeReconciler) deleteWorkspaceVolumeFromCluster(ctx context.Context, inClusterInstance *workspacev1alpha1.WorkspaceVolume) error {
	r.Log.Infof("deleting workspace volume: %s in namespace %s", inClusterInstance.Name, inClusterInstance.Namespace)
	if err := r.Client.Delete(ctx, inClusterInstance); client.IgnoreNotFound(err) != nil {
		r.Log.Errorf("failed to delete workspace volume: %v", err)
		return err
	}
	return nil
}

func mapToVolumeStatus(clusterStatus workspacev1alpha1.WorkspaceVolumeStatus) *models.VolumeStatus {
	return &models.VolumeStatus{
		ObservedGeneration:     clusterStatus.ObservedGeneration,
		Conditions:             models.ConvertConditions(clusterStatus.Conditions),
		Phase:                  string(clusterStatus.Phase),
		BuildArtifactSyncs:     mapToBuildArtifactSyncInfo(clusterStatus.BuildArtifactSyncs),
		LastObservedStatusHash: clusterStatus.StatusHash,
	}
}

func mapToBuildArtifactSyncInfo(clusterBuildArtifactSyncInfoMap map[workspacev1alpha1.ResourceRef]workspacev1alpha1.BuildArtifactSyncInfo) []models.BuildArtifactSyncInfo {
	res := make([]models.BuildArtifactSyncInfo, 0)
	for resourceRef, syncInfo := range clusterBuildArtifactSyncInfoMap {
		res = append(res, models.BuildArtifactSyncInfo{
			ResourceName: resourceRef.String(),
			BuildID:      syncInfo.BuildID,
			Status:       string(syncInfo.Status),
		})
	}
	return res
}
