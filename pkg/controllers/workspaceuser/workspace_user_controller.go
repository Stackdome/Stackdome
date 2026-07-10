package workspaceuser

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/controllers"
	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	usersv1alpha1 "stackdome.io/cluster-agent/api/users/v1alpha1"
)

const (
	controllerName = "workspace-user-controller"
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
	controller, err := controller.New(controllerName, manager, controller.Options{
		Reconciler: w,
	})
	if err != nil {
		return err
	}
	src := source.Kind(
		manager.GetCache(),
		&usersv1alpha1.User{},
		&handler.TypedEnqueueRequestForObject[*usersv1alpha1.User]{},
		controllers.DBObjectIDPresentPredicate[*usersv1alpha1.User](models.WorkspaceUserIDLabel),
	)

	return controller.Watch(src)
}

func (w *WorkspaceUserReconciler) Name() string {
	return controllerName
}

func (r *WorkspaceUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Infof("reconciling workspace user: %s in namespace %s", req.Name, req.Namespace)
	clusterInstance := &usersv1alpha1.User{}
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
	workspaceuser, serr := r.WorkspaceUserService.InternalGetByID(ctx, workspaceUserID)
	if serr != nil {
		if serr.Code == apperrors.ErrorNotFound {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get workspace user from db: %w", serr)
	}

	cluster, serr := r.ClusterService.InternalGet(ctx, workspaceuser.ClusterID)
	if serr != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get cluster from db: %w", serr)
	}

	// status changed
	if workspaceuser.Status.ClusterStatusHash != clusterInstance.Status.StatusHash {
		workspaceuser.Status = mapToDBStatusAndState(clusterInstance, cluster)
		serr := r.WorkspaceUserService.UpdateStatus(ctx, workspaceuser.ID, workspaceuser)
		if serr != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update workspace user status in db: %w", serr)
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// TODO: CorrelationID b/w db object and object in cluster.
func mapToDBStatusAndState(
	clusterObject *usersv1alpha1.User,
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
	availableCondition := meta.FindStatusCondition(clusterObject.Status.Conditions, usersv1alpha1.UserAvailable)
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
