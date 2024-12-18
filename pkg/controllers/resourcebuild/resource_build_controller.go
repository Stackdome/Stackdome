package resourcebuild

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
	"sigs.k8s.io/controller-runtime/pkg/source"
	workspacev1alpha1 "soradev.io/cluster-agent/api/v1alpha1"
)

type ResourceBuildReconciler struct {
	Client                 client.Client
	DBResourceBuildService services.ResourceBuildService
	DBResourceService      services.WorkspaceResourceService
	Logger                 logger.Logger
}

type ResourceBuildReconcilerSpec struct {
	Client                 client.Client
	DBResourceBuildService services.ResourceBuildService
	DBResourceService      services.WorkspaceResourceService
	Log                    logger.Logger
}

func NewResourceBuildReconciler(spec ResourceBuildReconcilerSpec) *ResourceBuildReconciler {
	return &ResourceBuildReconciler{
		Client:                 spec.Client,
		DBResourceBuildService: spec.DBResourceBuildService,
		DBResourceService:      spec.DBResourceService,
		Logger:                 spec.Log,
	}
}

// AddToManager adds the reconciler to the manager
func (r *ResourceBuildReconciler) AddToManager(manager manager.Manager) error {
	r.Client = manager.GetClient()
	controller, err := controller.New("resource-build-controller", manager, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}

	src := source.Kind(
		manager.GetCache(),
		&workspacev1alpha1.WorkspaceApplicationBuild{},
		&handler.TypedEnqueueRequestForObject[*workspacev1alpha1.WorkspaceApplicationBuild]{},
		controllers.WorkspaceIDLabelPresentPredicate[*workspacev1alpha1.WorkspaceApplicationBuild](),
	)

	return controller.Watch(src)
}

func (r *ResourceBuildReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	resourceBuild := &workspacev1alpha1.WorkspaceApplicationBuild{}
	if err := r.Client.Get(ctx, req.NamespacedName, resourceBuild); err != nil {
		if errors.IsNotFound(err) {
			r.Logger.Infof("ResourceBuild %s not found", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	r.Logger.Infof("Reconciling resource build", "resource_build", req.NamespacedName)

	workspaceID, ok := resourceBuild.Labels[models.WorkspaceIDLabel]
	if !ok {
		r.Logger.Errorf("ResourceBuild %s does not have workspace ID label", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	dbWorkspaceResouce, err := r.DBResourceService.GetByWorkspaceIDAndResourceName(ctx, workspaceID, resourceBuild.Spec.ResourceName)
	if err != nil {
		r.Logger.Errorf("Failed to get workspace resource %s for build '%s'", resourceBuild.Spec.ResourceName, client.ObjectKeyFromObject(resourceBuild).String())
		return ctrl.Result{}, err
	}

	dbResourceBuild, err := r.DBResourceBuildService.GetByID(ctx, resourceBuild.Name)
	if err != nil {
		if err.Code == apperrors.ErrorNotFound {
			r.Logger.Infof("ResourceBuild %s not found in DB, creating a new build", resourceBuild.Name)
			return ctrl.Result{Requeue: true}, r.createResourceBuildInDB(ctx, resourceBuild, dbWorkspaceResouce)
		}
		return ctrl.Result{}, err
	}

	if dbResourceBuild.Status == nil || dbResourceBuild.Status.LastObservedStatusHash != resourceBuild.Status.StatusHash {
		dbResourceBuild.Status = mapClusterStatusToServerStatus(resourceBuild.Status)
		return ctrl.Result{}, r.DBResourceBuildService.UpdateStatus(ctx, dbResourceBuild.ID, dbResourceBuild.Status)
	}

	return ctrl.Result{}, nil
}

func (r *ResourceBuildReconciler) createResourceBuildInDB(ctx context.Context, resourceBuild *workspacev1alpha1.WorkspaceApplicationBuild, DBworkspaceResource *models.WorkspaceResource) error {
	dbResourceBuild := &models.WorkspaceResourceBuild{
		ID:                    resourceBuild.Name,
		WorkspaceResourceID:   DBworkspaceResource.ID,
		WorkspaceResourceName: DBworkspaceResource.Name,
		Namespace:             resourceBuild.Namespace,
		WorkspaceID:           DBworkspaceResource.WorkspaceID,
		BuildSourceHash:       resourceBuild.Spec.SourceHash,
		ImageRegistry:         resourceBuild.Spec.Registry,
		Status:                mapClusterStatusToServerStatus(resourceBuild.Status),
	}

	_, err := r.DBResourceBuildService.Create(ctx, dbResourceBuild)
	if err != nil {
		r.Logger.Errorf("Failed to create resource build '%s': %s", resourceBuild.Name, err)
		return err
	}
	return nil
}

func mapClusterStatusToServerStatus(clusterStatus workspacev1alpha1.WorkspaceApplicationBuildStatus) *models.WorkspaceResourceBuildStatus {
	return &models.WorkspaceResourceBuildStatus{
		Conditions:             models.ConvertConditions(clusterStatus.Conditions),
		State:                  string(clusterStatus.Phase),
		BuildSourceHash:        clusterStatus.BuildSourceHash,
		ImageURL:               clusterStatus.ImageUrl,
		LastObservedStatusHash: clusterStatus.StatusHash,
	}
}
