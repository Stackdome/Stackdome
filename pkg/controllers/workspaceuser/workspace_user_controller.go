package workspaceuser

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/controllers"
	apperrors "github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	workspacev1alpha1 "soradev.io/cluster-agent/api/v1alpha1"
)

type WorkspaceUserReconciler struct {
	Client               client.Client
	Log                  logger.Logger
	WorkspaceUserService services.WorkspaceUserService
	ClusterService       services.ClusterService
	Env                  string
}

type WorkspaceUserReconcilerSpec struct {
	Log                  logger.Logger
	WorkspaceUserService services.WorkspaceUserService
	ClusterService       services.ClusterService
	Env                  string
}

func NewWorkspaceUserReconciler(spec WorkspaceUserReconcilerSpec) *WorkspaceUserReconciler {
	return &WorkspaceUserReconciler{
		Log:                  spec.Log,
		WorkspaceUserService: spec.WorkspaceUserService,
		ClusterService:       spec.ClusterService,
		Env:                  spec.Env,
	}
}

func (w *WorkspaceUserReconciler) AddToManager(manager manager.Manager) error {
	w.Client = manager.GetClient()
	controller, err := controller.New("workspace-user-controller", manager, controller.Options{
		Reconciler: w,
	})
	if err != nil {
		return err
	}
	src := source.Kind(
		manager.GetCache(),
		&workspacev1alpha1.WorkspaceUser{},
		&handler.TypedEnqueueRequestForObject[*workspacev1alpha1.WorkspaceUser]{},
		controllers.DBObjectIDPresentPredicate[*workspacev1alpha1.WorkspaceUser](models.WorkspaceUserIDLabel),
	)

	return controller.Watch(src)
}

func (r *WorkspaceUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Infof("reconciling workspace user: %s in namespace %s", req.Name, req.Namespace)
	clusterInstance := &workspacev1alpha1.WorkspaceUser{}
	if err := r.Client.Get(ctx, req.NamespacedName, clusterInstance); err != nil {
		r.Log.Errorf("failed to get workspace user from cluster: %v", err)
		return ctrl.Result{}, nil
	}

	workspaceUserID, ok := clusterInstance.Labels[models.WorkspaceUserIDLabel]
	if !ok {
		r.Log.Errorf("workspace user ID not found in workspaceuser labels")
		return ctrl.Result{}, nil
	}

	// TODO: When to garbage collect the workspace user in cluster if it is not found in db?
	workspaceuser, serr := r.WorkspaceUserService.GetByID(ctx, workspaceUserID)
	if serr != nil {
		if serr.Code == apperrors.ErrorNotFound {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get workspace user from db: %v", serr)
	}

	cluster, serr := r.ClusterService.Get(ctx, workspaceuser.ClusterID)
	if serr != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get cluster from db: %v", serr)
	}

	// status changed
	if workspaceuser.Status.ClusterStatusHash != clusterInstance.Status.StatusHash {
		workspaceuser.Status = mapToDBStatusAndState(clusterInstance, cluster)
		serr := r.WorkspaceUserService.UpdateStatus(ctx, workspaceuser.ID, workspaceuser)
		if serr != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update workspace user status in db: %v", serr)
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *WorkspaceUserReconciler) deleteWorkspaceUserFromCluster(ctx context.Context, inClusterInstance *workspacev1alpha1.WorkspaceUser) error {
	r.Log.Infof("deleting workspace user: %s in namespace %s", inClusterInstance.Name, inClusterInstance.Namespace)
	if err := r.Client.Delete(ctx, inClusterInstance); client.IgnoreNotFound(err) != nil {
		r.Log.Errorf("failed to delete workspace volume: %v", err)
		return err
	}
	return nil
}

// TODO: CorrelationID b/w db object and object in cluster.
func mapToDBStatusAndState(
	clusterObject *workspacev1alpha1.WorkspaceUser,
	cluster *models.Cluster) *models.WorkspaceUserStatus {

	status := &models.WorkspaceUserStatus{
		ObservedVersion:       clusterObject.Status.ObservedStackdomeServerObjectGeneration,
		ServiceAccountName:    clusterObject.Status.ServiceAccountName,
		ServiceAccountToken:   clusterObject.Status.ServiceAccountToken,
		ProvisionedNamespaces: clusterObject.Status.Namespaces,
		ClusterCACert:         cluster.ClusterCAData,
		ClusterUrl:            cluster.ClusterURL,
		ClusterStatusHash:     clusterObject.Status.StatusHash,
		Conditions:            models.ConvertConditions(clusterObject.Status.Conditions),
	}
	availableCondition := meta.FindStatusCondition(clusterObject.Status.Conditions, workspacev1alpha1.WorkspaceUserAvailable)
	switch {
	case availableCondition == nil:
		status.State = models.WorkspaceUserProvisionPending
		status.Message = "WorkspaceUser provision pending"
	case availableCondition.ObservedGeneration != clusterObject.Generation:
		status.State = models.WorkspaceUserProvisionPending
		status.Message = "WorkspaceUser not uptodate"
		// Invalidate the version.
		status.ObservedVersion = clusterObject.Status.ObservedStackdomeServerObjectGeneration - 1
	case availableCondition.Status == v1.ConditionFalse:
		status.State = models.WorkspaceUserProvisionPending
		status.Message = availableCondition.Message
	case availableCondition.Status == v1.ConditionTrue:
		status.State = models.WorkspaceUserProvisionCompleted
		status.Message = "WorkspaceUser provision completed"
	}
	return status
}

func invalidatedDBGenerationHash(hash string) string {
	return hash + "X"
}
