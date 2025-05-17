package clusterimageregistry

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/controllers"
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
	registryv1alpha1 "stackdome.io/cluster-agent/api/registry/v1alpha1"
)

const (
	controllerName = "cluster-image-registry-controller"
)

type clusterImageRegistryReconciler struct {
	Client                 client.Client
	DBImageRegistryService services.ClusterImageRegistryService
	Logger                 logger.Logger
}

type ClusterImageRegistryReconcilerSpec struct {
	DBImageRegistryService services.ClusterImageRegistryService
	Logger                 logger.Logger
}

func NewClusterImageRegistryReconciler(spec ClusterImageRegistryReconcilerSpec) *clusterImageRegistryReconciler {
	return &clusterImageRegistryReconciler{
		DBImageRegistryService: spec.DBImageRegistryService,
		Logger:                 spec.Logger,
	}
}

// AddToManager adds the reconciler to the manager
func (r *clusterImageRegistryReconciler) AddToManager(manager manager.Manager) error {
	r.Client = manager.GetClient()
	controller, err := controller.New(controllerName, manager, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}

	src := source.Kind(
		manager.GetCache(),
		&registryv1alpha1.ClusterRegistry{},
		&handler.TypedEnqueueRequestForObject[*registryv1alpha1.ClusterRegistry]{},
		controllers.ClusterImageRegistryIDLabelPresentPredicate[*registryv1alpha1.ClusterRegistry](),
	)

	return controller.Watch(src)
}

func (r *clusterImageRegistryReconciler) Name() string {
	return controllerName
}

func (r *clusterImageRegistryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	registryCr := &registryv1alpha1.ClusterRegistry{}
	if err := r.Client.Get(ctx, req.NamespacedName, registryCr); err != nil {
		if errors.IsNotFound(err) {
			r.Logger.Infof("cluster registry %s not found", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	r.Logger.Infof("Reconciling cluster registry", "cluster_image_registry", req.NamespacedName)

	registryID, ok := registryCr.Labels[models.ImageRegistryIDLabel]
	if !ok {
		r.Logger.Errorf("ClusterRegistry %s does not have cluster registry ID label", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	dbImageRegistry, err := r.DBImageRegistryService.Get(ctx, registryID)
	if err != nil {
		r.Logger.Error(ctx, "Failed to get cluster image registry from DB", "error", err)
		return ctrl.Result{}, err
	}

	if dbImageRegistry.Status == nil ||
		string(dbImageRegistry.Status.State) != string(registryCr.Status.Phase) ||
		len(dbImageRegistry.Status.Conditions) != len(registryCr.Status.Conditions) || dbImageRegistry.Status.RegistryUrl == "" {
		dbImageRegistry.Status = mapClusterStatusToServerStatus(registryCr.Status)
		return ctrl.Result{}, r.DBImageRegistryService.UpdateStatus(ctx, dbImageRegistry.ID, dbImageRegistry.Status)
	}

	return ctrl.Result{}, nil
}

func mapClusterStatusToServerStatus(clusterStatus registryv1alpha1.RegistryStatus) *models.ClusterImageRegistryStatus {
	return &models.ClusterImageRegistryStatus{
		Conditions:  models.ConvertConditions(clusterStatus.Conditions),
		State:       mapClusterPhaseToServerState(clusterStatus.Phase),
		RegistryUrl: clusterStatus.InternalURL,
	}
}

func mapClusterPhaseToServerState(phase registryv1alpha1.RegistryPhase) models.RegistryState {
	switch phase {
	case registryv1alpha1.RegistryPhasePending:
		return models.RegistryStatePending
	case registryv1alpha1.RegistryPhaseRunning:
		return models.RegistryStateRunning
	case registryv1alpha1.RegistryPhaseFailed:
		return models.RegistryStateError
	default:
		return models.RegistryStatePending
	}
}
