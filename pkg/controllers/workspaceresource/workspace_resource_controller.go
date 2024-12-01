package workspaceresource

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

type workspaceResourceReconciler struct {
	client                   client.Client
	workspaceResourceService services.WorkspaceResourceService
	workspaceService         services.WorkspaceService
	logger                   logger.Logger
	env                      string
}

type WorkspaceResourceReconcilerSpec struct {
	WorkspaceResourceService services.WorkspaceResourceService
	WorkspaceService         services.WorkspaceService
	Log                      logger.Logger
	Env                      string
}

func NewWorkspaceResourceReconciler(spec WorkspaceResourceReconcilerSpec) *workspaceResourceReconciler {
	return &workspaceResourceReconciler{
		client:                   nil,
		workspaceResourceService: spec.WorkspaceResourceService,
		workspaceService:         spec.WorkspaceService,
		logger:                   spec.Log,
		env:                      spec.Env,
	}
}

// AddToManager adds the reconciler to the manager
func (w *workspaceResourceReconciler) AddToManager(manager manager.Manager) error {
	w.client = manager.GetClient()

	controller, err := controller.New("workspace-resource-controller", manager, controller.Options{
		Reconciler: w,
	})
	if err != nil {
		return err
	}

	src := source.Kind(
		manager.GetCache(),
		&workspacev1alpha1.WorkspaceResource{},
		&handler.TypedEnqueueRequestForObject[*workspacev1alpha1.WorkspaceResource]{},
		controllers.WorkspaceIDLabelPresentPredicate[*workspacev1alpha1.WorkspaceResource](),
	)

	return controller.Watch(src)
}

func (w *workspaceResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	w.logger.Infof("Reconciling workspace resource", "workspace_resource", req.NamespacedName)

	workspaceResource := &workspacev1alpha1.WorkspaceResource{}
	if err := w.client.Get(ctx, req.NamespacedName, workspaceResource); err != nil {
		if errors.IsNotFound(err) {
			w.logger.Infof("WorkspaceResource %s not found", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	workspaceID, ok := workspaceResource.Labels[models.WorkspaceIDLabel]
	if !ok {
		// How are we here? The predicate should have prevented this.
		w.logger.Errorf("WorkspaceResource %s does not have a workspace id label", workspaceResource.Name)
		return ctrl.Result{}, nil
	}

	dbWorkspaceResource, serr := w.workspaceResourceService.GetByWorkspaceIDAndResourceName(ctx, workspaceID, workspaceResource.Name)
	if serr != nil {
		if serr.Code == apperrors.ErrorNotFound {
			w.logger.Infof("WorkspaceResource %s not found in DB", workspaceResource.Name)
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, serr
	}

	if dbWorkspaceResource.Status == nil || dbWorkspaceResource.Status.LastObservedStatusHash != workspaceResource.Status.StatusHash {
		dbWorkspaceResource.Status = mapClusterStatusToServerStatus(workspaceResource)
		return ctrl.Result{}, w.workspaceResourceService.UpdateStatus(ctx, dbWorkspaceResource.ID, dbWorkspaceResource.Status)
	}
	return ctrl.Result{}, nil
}

func mapClusterStatusToServerStatus(clusterInstance *workspacev1alpha1.WorkspaceResource) *models.WorkspaceResourceStatus {
	return &models.WorkspaceResourceStatus{
		LastObservedStatusHash: clusterInstance.Status.StatusHash,
		State:                  string(clusterInstance.Status.Phase),
		Conditions:             models.ConvertConditions(clusterInstance.Status.Conditions),
		PublicIngresses:        mapToPublicIngresses(clusterInstance.Status.ExternalAddress),
		ObservedVersion:        clusterInstance.Status.ObservedStackdomeServerObjectGeneration,
		InternalServiceName:    clusterInstance.Status.InternalAddress,
	}
}

func mapToPublicIngresses(externalAddresses []workspacev1alpha1.ExternalAddress) []models.Ingress {
	var publicIngresses []models.Ingress
	for _, externalAddress := range externalAddresses {
		publicIngresses = append(publicIngresses, models.Ingress{
			URL:        externalAddress.Address,
			TargetPort: int(externalAddress.TargetPort),
		})
	}
	return publicIngresses
}
