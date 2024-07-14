package workspacestorage

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/controllers"
	apperrors "github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	workspacev1alpha1 "soradev.io/cluster-agent/api/v1alpha1"
)

type WorskspaceStorageReconciler struct {
	Client                  client.Client
	Log                     logger.Logger
	WorkspaceStorageService services.WorkspaceStorageService
	VolumeService           services.VolumeService
	Env                     string
}

type WorskspaceStorageReconcilerSpec struct {
	Client                  client.Client
	Log                     logger.Logger
	WorkspaceStorageService services.WorkspaceStorageService
	VolumeService           services.VolumeService
	Env                     string
}

func NewWorskspaceStorageReconciler(spec WorskspaceStorageReconcilerSpec) *WorskspaceStorageReconciler {
	return &WorskspaceStorageReconciler{
		Client:                  spec.Client,
		Log:                     spec.Log,
		WorkspaceStorageService: spec.WorkspaceStorageService,
		VolumeService:           spec.VolumeService,
		Env:                     spec.Env,
	}
}

func (w *WorskspaceStorageReconciler) AddToManager(manager manager.Manager) error {
	controller, err := controller.New("workspace-storage-controller", manager, controller.Options{
		Reconciler: w,
	})
	if err != nil {
		return err
	}

	src := source.Kind(
		manager.GetCache(),
		&workspacev1alpha1.WorkspaceStorage{},
		&handler.TypedEnqueueRequestForObject[*workspacev1alpha1.WorkspaceStorage]{},
		controllers.WorkspaceStorageIdLabelPresentPredicate[*workspacev1alpha1.WorkspaceStorage](),
	)

	return controller.Watch(src)
}

func (r *WorskspaceStorageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Infof("reconciling workspace storage: %s in namespace %s", req.Name, req.Namespace)

	clusterInstance := &workspacev1alpha1.WorkspaceStorage{}
	err := r.Client.Get(ctx, req.NamespacedName, clusterInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	workspaceStorageID, ok := clusterInstance.Labels[models.WorkspaceStorageIDLabel]
	if !ok {
		r.Log.Errorf("workspace storage %s in namespace %s does not have a workspace storage id label", clusterInstance.Name, clusterInstance.Namespace)
		return ctrl.Result{}, nil
	}
	dbInstance, serr := r.WorkspaceStorageService.Get(ctx, workspaceStorageID)
	if err != nil {
		if serr.Code == apperrors.ErrorNotFound {
			return ctrl.Result{}, r.deleteWorkspaceStorageFromCluster(ctx, clusterInstance)
		}
		return ctrl.Result{}, fmt.Errorf("failed to get workspace storage %s in namespace %s: %w from DB", clusterInstance.Name, clusterInstance.Namespace, serr)
	}

	stateChanged := mapClusterObjStateToDBObjState(clusterInstance, dbInstance)
	objectHashChanged := clusterInstance.Status.StatusHash != dbInstance.Status.LastObservedStatusHash
	if objectHashChanged {
		dbInstance.Status = mapClusterObjStatusToDBObj(clusterInstance)
	}
	if objectHashChanged || stateChanged {
		_, serr = r.WorkspaceStorageService.Update(ctx, workspaceStorageID, dbInstance)
		if serr != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update workspace storage %s in namespace %s: %w from DB", clusterInstance.Name, clusterInstance.Namespace, serr)
		}
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

func mapClusterObjStatusToDBObj(clusterInstance *workspacev1alpha1.WorkspaceStorage) *models.WorkspaceStorageStatus {
	clusterObjectStatus := clusterInstance.Status
	return &models.WorkspaceStorageStatus{
		Phase:                    string(clusterObjectStatus.Phase),
		Conditions:               models.ConvertConditions(clusterObjectStatus.Conditions),
		StorageServerServiceName: clusterObjectStatus.ServiceName,
		LastObservedStatusHash:   clusterObjectStatus.StatusHash,
	}
}

func mapClusterObjStateToDBObjState(clusterInstance *workspacev1alpha1.WorkspaceStorage, dbInstance *models.WorkspaceStorage) bool {
	availableCondition := meta.FindStatusCondition(clusterInstance.Status.Conditions, string(workspacev1alpha1.WorkspaceStorageAvailable))
	failedCondition := meta.FindStatusCondition(clusterInstance.Status.Conditions, string(workspacev1alpha1.WorkspaceStorageFailed))
	switch {
	case availableCondition == nil:
		return dbInstance.SetState(models.WorkspaceStorageStatePending)
	case availableCondition.Status == metav1.ConditionTrue:
		return dbInstance.SetState(models.WorkspaceStorageStateReady)
	case availableCondition.Status == metav1.ConditionFalse:
		return dbInstance.SetState(models.WorkspaceStorageStateCreating)
	case failedCondition != nil && failedCondition.Status == metav1.ConditionTrue:
		return dbInstance.SetState(models.WorkspaceStorageStateFailed)
	}
	return false
}

func (r *WorskspaceStorageReconciler) deleteWorkspaceStorageFromCluster(ctx context.Context, instance *workspacev1alpha1.WorkspaceStorage) error {
	r.Log.Infof("deleting workspace storage: %s in namespace %s", instance.Name, instance.Namespace)
	err := r.Client.Delete(ctx, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete workspace storage %s in namespace %s: %w from cluster", instance.Name, instance.Namespace, err)
	}
	return nil
}

func (r *WorskspaceStorageReconciler) deleteWorkspaceStorageFromDB(ctx context.Context, instance *workspacev1alpha1.WorkspaceStorage) error {
	r.Log.Infof("deleting workspace storage: %s in namespace %s", instance.Name, instance.Namespace)
	id, found := instance.Labels[models.WorkspaceStorageIDLabel]
	if !found {
		r.Log.Errorf("workspace storage %s in namespace %s does not have a workspace storage id label", instance.Name, instance.Namespace)
		return nil
	}
	serr := r.WorkspaceStorageService.Delete(ctx, id)
	if serr != nil {
		if serr.Code == apperrors.ErrorNotFound {
			return nil
		}
		return fmt.Errorf("failed to delete workspace storage %s in namespace %s: %w from DB", instance.Name, instance.Namespace, serr)
	}
	return nil
}
