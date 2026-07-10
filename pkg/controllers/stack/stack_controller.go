package workspace

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/controllers"
	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/Stackdome/stackdome/pkg/worker/workermanager"
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

// releaseActiveChecker is a narrow interface to check for active releases without
// importing the full StackReleaseService.
type releaseActiveChecker interface {
	InternalGetActiveByStackID(ctx context.Context, stackID string) (*models.StackRelease, *apperrors.ServiceError)
}

type stackReconciler struct {
	Client         client.Client
	StackService   services.StackService
	Log            logger.Logger
	Env            string
	releaseChecker releaseActiveChecker
	enqueuer       workermanager.BackgroundJobEnqueuer
}

type StackReconcilerSpec struct {
	Log            logger.Logger
	StackService   services.StackService
	Env            string
	ReleaseChecker releaseActiveChecker
	Enqueuer       workermanager.BackgroundJobEnqueuer
}

func NewStackReconciler(spec StackReconcilerSpec) *stackReconciler {
	return &stackReconciler{
		Client:         nil,
		StackService:   spec.StackService,
		Log:            spec.Log,
		Env:            spec.Env,
		releaseChecker: spec.ReleaseChecker,
		enqueuer:       spec.Enqueuer,
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
	w.Log.Infof("Reconciling stack %v", req.NamespacedName)
	stackCr := &corev1alpha1.Stack{}
	if err := w.Client.Get(ctx, req.NamespacedName, stackCr); err != nil {
		if errors.IsNotFound(err) {
			w.Log.Infof("Stack %s not found", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	stackID, ok := stackCr.Labels[corev1alpha1.LabelStackID]
	if !ok {
		// How are we here? The predicate should have prevented this.
		w.Log.Errorf("Stack %s does not have a workspace id label", stackCr.Name)
		return ctrl.Result{}, nil
	}
	dbStack, serr := w.StackService.InternalGetStack(ctx, stackID)
	if serr != nil {
		if serr.Code == apperrors.ErrorNotFound {
			w.Log.Infof("Stack %s not found in DB", stackCr.Name)
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get stack from db: %w", serr)
	}

	if dbStack.Status == nil {
		dbStack.Status = &models.StackStatus{}
	}
	currentStackState := dbStack.Status.State
	currentStackCrState := mapStackState(stackCr.Status.Phase)
	if stackCr.Status.StatusHash != dbStack.Status.LastObservedStatusHash || currentStackState != currentStackCrState {
		lastValidationRun := dbStack.Status.LastValidationRun
		dbStack.Status = mapClusterObjStatusToDBObjStatus(stackCr)
		dbStack.Status.LastValidationRun = lastValidationRun
		serr = w.StackService.UpdateStatus(ctx, stackID, dbStack.Status)
		if serr != nil {
			w.Log.Errorf("Failed to update stack '%s' status : %s", dbStack.ID, serr)
			return ctrl.Result{}, fmt.Errorf("failed to update stack status: %w", serr)
		}

		// Kick the release worker so it can observe the new status quickly.
		w.enqueueActiveRelease(ctx, stackID)

		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

func (w *stackReconciler) enqueueActiveRelease(ctx context.Context, stackID string) {
	if w.releaseChecker == nil || w.enqueuer == nil {
		return
	}
	active, serr := w.releaseChecker.InternalGetActiveByStackID(ctx, stackID)
	if serr != nil {
		w.Log.Errorf("Failed to check active release for stack '%s': %s", stackID, serr)
		return
	}
	if active == nil {
		return
	}
	if err := w.enqueuer.Enqueue(&models.StackRelease{ID: active.ID}); err != nil {
		w.Log.Errorf("Failed to enqueue release '%s': %v", active.ID, err)
	}
}

func mapClusterObjStatusToDBObjStatus(clusterInstance *corev1alpha1.Stack) *models.StackStatus {
	status := &models.StackStatus{
		State:                  mapStackState(clusterInstance.Status.Phase),
		TargetRevision:         clusterInstance.Status.TargetRevision,
		Conditions:             models.ConvertConditions(clusterInstance.Status.Conditions),
		LastObservedStatusHash: clusterInstance.Status.StatusHash,
	}

	if clusterInstance.Status.LastConverged != nil {
		status.LastConverged = &models.StackConvergenceRecord{
			Revision:  clusterInstance.Status.LastConverged.Revision,
			ReleaseID: clusterInstance.Status.LastConverged.ReleaseID,
			At:        clusterInstance.Status.LastConverged.At.Time,
		}
	}

	if len(clusterInstance.Status.Resources) > 0 {
		status.Resources = make([]models.StackResourceSummary, len(clusterInstance.Status.Resources))
		for i, r := range clusterInstance.Status.Resources {
			summary := models.StackResourceSummary{
				Name:              r.Name,
				Phase:             models.StackResourcePhase(r.Phase),
				ObservedRevision:  r.ObservedRevision,
				ConvergedRevision: r.ConvergedRevision,
				AvailableReplicas: r.AvailableReplicas,
				UpdatedReplicas:   r.UpdatedReplicas,
				Replicas:          r.Replicas,
				Missing:           r.Missing,
				Message:           r.Message,
			}
			if r.LastConverged != nil {
				summary.LastConverged = &models.StackResourceConvergenceRecord{
					Revision: r.LastConverged.Revision,
					At:       r.LastConverged.At.Time,
				}
			}
			status.Resources[i] = summary
		}
	}

	return status
}

func mapStackState(in corev1alpha1.StackPhase) models.StackState {
	switch in {
	case corev1alpha1.StackPending:
		return models.StackPending
	case corev1alpha1.StackReady:
		return models.StackReady
	case corev1alpha1.StackFailed:
		return models.StackFailed
	case corev1alpha1.StackDegraded:
		return models.StackDegraded
	case corev1alpha1.StackProgressing:
		return models.StackProgressing
	default:
		return models.StackPending
	}
}
