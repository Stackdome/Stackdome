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
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

const (
	controllerName = "stack-controller"
)

type stackReconciler struct {
	Client       client.Client
	StackService services.StackService
	Log          logger.Logger
	Env          string
}

type StackReconcilerSpec struct {
	Log          logger.Logger
	StackService services.StackService
	Env          string
}

func NewStackReconciler(spec StackReconcilerSpec) *stackReconciler {
	return &stackReconciler{
		Client:       nil,
		StackService: spec.StackService,
		Log:          spec.Log,
		Env:          spec.Env,
	}
}

// AddToManager adds the reconciler to the manager
func (w *stackReconciler) AddToManager(manager manager.Manager) error {
	w.Client = manager.GetClient()
	controller, err := controller.New(controllerName, manager, controller.Options{
		Reconciler: w,
	})

	if err != nil {
		return err
	}

	src := source.Kind(
		manager.GetCache(),
		&corev1alpha1.Stack{},
		&handler.TypedEnqueueRequestForObject[*corev1alpha1.Stack]{},
		controllers.StackIDLabelPresentPredicate[*corev1alpha1.Stack](),
	)

	return controller.Watch(src)
}

func (w *stackReconciler) Name() string {
	return controllerName
}

// Reconcile reconciles the workspace storage resource
func (w *stackReconciler) Reconcile(ctx context.Context, req reconcile.Request) (ctrl.Result, error) {
	w.Log.Infof("Reconciling stack %s", req.NamespacedName)
	stackCr := &corev1alpha1.Stack{}
	if err := w.Client.Get(ctx, req.NamespacedName, stackCr); err != nil {
		if errors.IsNotFound(err) {
			w.Log.Infof("Stack %s not found", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	stackID, ok := stackCr.Labels[models.StackIDLabel]
	if !ok {
		// How are we here? The predicate should have prevented this.
		w.Log.Errorf("Stack %s does not have a workspace id label", stackCr.Name)
		return ctrl.Result{}, nil
	}
	dbStack, serr := w.StackService.GetStack(ctx, stackID)
	if serr != nil {
		if serr.Code == apperrors.ErrorNotFound {
			w.Log.Infof("Stack %s not found in DB", stackCr.Name)
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, serr
	}

	if dbStack.Status == nil || stackCr.Status.StatusHash != dbStack.Status.LastObservedStatusHash {
		dbStack.Status = mapClusterObjStatusToDBObjStatus(stackCr)
		serr = w.StackService.UpdateStatus(ctx, stackID, dbStack.Status)
		if serr != nil {
			w.Log.Errorf("Failed to update stack '%s' status : %s", dbStack.ID, serr)
			return ctrl.Result{}, serr
		}
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

func mapClusterObjStatusToDBObjStatus(clusterInstance *corev1alpha1.Stack) *models.StackStatus {
	return &models.StackStatus{
		State:                  mapStackState(clusterInstance.Status.Phase),
		ObservedVersion:        clusterInstance.Status.ObservedStackdomeServerObjectGeneration,
		Conditions:             models.ConvertConditions(clusterInstance.Status.Conditions),
		LastObservedStatusHash: clusterInstance.Status.StatusHash,
	}
}

func mapStackState(in corev1alpha1.StackPhase) models.StackState {
	switch in {
	case corev1alpha1.StackPending:
		return models.StackPending
	case corev1alpha1.StackReady:
		return models.StackReady
	case corev1alpha1.StackFailed:
		return models.StackFailed
	default:
		return models.StackPending
	}
}

// TODO:
// Add ObservedStackdomeServerObjectGeneration to the Workspace CR
