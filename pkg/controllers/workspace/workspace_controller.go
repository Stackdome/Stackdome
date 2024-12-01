package workspace

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/controllers"
	apperrors "github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	workspacev1alpha1 "soradev.io/cluster-agent/api/v1alpha1"
)

type workspaceReconciler struct {
	Client           client.Client
	WorkspaceService services.WorkspaceService
	Log              logger.Logger
	Env              string
}

type WorkspaceReconcilerSpec struct {
	Log              logger.Logger
	WorkspaceService services.WorkspaceService
	Env              string
}

func NewWorkspaceReconciler(spec WorkspaceReconcilerSpec) *workspaceReconciler {
	return &workspaceReconciler{
		Client:           nil,
		WorkspaceService: spec.WorkspaceService,
		Log:              spec.Log,
		Env:              spec.Env,
	}
}

// AddToManager adds the reconciler to the manager
func (w *workspaceReconciler) AddToManager(manager manager.Manager) error {
	w.Client = manager.GetClient()
	controller, err := controller.New("workspace-controller", manager, controller.Options{
		Reconciler: w,
	})
	if err != nil {
		return err
	}

	src := source.Kind(
		manager.GetCache(),
		&workspacev1alpha1.Workspace{},
		&handler.TypedEnqueueRequestForObject[*workspacev1alpha1.Workspace]{},
		controllers.WorkspaceIDLabelPresentPredicate[*workspacev1alpha1.Workspace](),
	)

	return controller.Watch(src)
}

// Reconcile reconciles the workspace storage resource
func (w *workspaceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (ctrl.Result, error) {
	w.Log.Infof("Reconciling workspace %s", req.NamespacedName)
	workspace := &workspacev1alpha1.Workspace{}
	if err := w.Client.Get(ctx, req.NamespacedName, workspace); err != nil {
		if errors.IsNotFound(err) {
			w.Log.Infof("Workspace %s not found", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	workspaceID, ok := workspace.Labels[models.WorkspaceIDLabel]
	if !ok {
		// How are we here? The predicate should have prevented this.
		w.Log.Errorf("Workspace %s does not have a workspace id label", workspace.Name)
		return ctrl.Result{}, nil
	}
	dbWorkspace, serr := w.WorkspaceService.GetWorkspace(ctx, workspaceID)
	if serr != nil {
		if serr.Code == apperrors.ErrorNotFound {
			w.Log.Infof("Workspace %s not found in DB", workspace.Name)
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, serr
	}

	if dbWorkspace.Status == nil || workspace.Status.StatusHash != dbWorkspace.Status.LastObservedStatusHash {
		dbWorkspace.Status = mapClusterObjStatusToDBObjStatus(workspace)
		serr = w.WorkspaceService.UpdateStatus(ctx, workspaceID, dbWorkspace.Status)
		if serr != nil {
			w.Log.Errorf("Failed to update workspace '%s' status : %s", dbWorkspace.ID, serr)
			return ctrl.Result{}, serr
		}
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

func mapClusterObjStatusToDBObjStatus(clusterInstance *workspacev1alpha1.Workspace) *models.WorkspaceStatus {
	return &models.WorkspaceStatus{
		State:                  string(clusterInstance.Status.Phase),
		ObservedVersion:        clusterInstance.Status.ObservedStackdomeServerObjectGeneration,
		Conditions:             models.ConvertConditions(clusterInstance.Status.Conditions),
		LastObservedStatusHash: clusterInstance.Status.StatusHash,
	}
}

// TODO:
// Add ObservedStackdomeServerObjectGeneration to the Workspace CR
