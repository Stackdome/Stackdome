package workspaceresource

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/controllers"
	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

const (
	controllerName = "stack-resource-controller"
)

type stackResourceReconciler struct {
	client               client.Client
	stackResourceService services.StackResourceService
	stackService         services.StackService
	logger               logger.Logger
	env                  string
}

type StackResourceReconcilerSpec struct {
	StackResourceService services.StackResourceService
	StackService         services.StackService
	Log                  logger.Logger
	Env                  string
}

func NewStackResourceReconciler(spec StackResourceReconcilerSpec) *stackResourceReconciler {
	return &stackResourceReconciler{
		client:               nil,
		stackResourceService: spec.StackResourceService,
		stackService:         spec.StackService,
		logger:               spec.Log,
		env:                  spec.Env,
	}
}

// AddToManager adds the reconciler to the manager
func (w *stackResourceReconciler) AddToManager(manager manager.Manager) error {
	w.client = manager.GetClient()

	controller, err := controller.New(controllerName, manager, controller.Options{
		Reconciler: w,
	})
	if err != nil {
		return err
	}

	src := source.Kind(
		manager.GetCache(),
		&corev1alpha1.StackResource{},
		&handler.TypedEnqueueRequestForObject[*corev1alpha1.StackResource]{},
		controllers.StackIDLabelPresentPredicate[*corev1alpha1.StackResource](),
	)

	return controller.Watch(src)
}

func (w *stackResourceReconciler) Name() string {
	return controllerName
}

func (w *stackResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	w.logger.Infof("Reconciling stack resource: %v", req.NamespacedName)

	stackResourceCr := &corev1alpha1.StackResource{}
	if err := w.client.Get(ctx, req.NamespacedName, stackResourceCr); err != nil {
		if errors.IsNotFound(err) {
			w.logger.Infof("StackResource %v not found", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	stackID, ok := stackResourceCr.Labels[corev1alpha1.LabelStackID]
	if !ok {
		// How are we here? The predicate should have prevented this.
		w.logger.Errorf("StackResource %s does not have a stack id label", stackResourceCr.Name)
		return ctrl.Result{}, nil
	}

	dbStackResource, serr := w.stackResourceService.InternalGetByStackIDAndResourceName(ctx, stackID, stackResourceCr.Name)
	if serr != nil {
		if serr.Code == apperrors.ErrorNotFound {
			w.logger.Infof("StackResource %s not found in DB", stackResourceCr.Name)
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get stack resource from db: %v", serr)
	}

	if dbStackResource.Status == nil || dbStackResource.Status.LastObservedStatusHash != stackResourceCr.Status.StatusHash {
		dbStackResource.Status = computeStatusRewrite(dbStackResource.Status, stackResourceCr)
		if serr := w.stackResourceService.UpdateStatus(ctx, dbStackResource.ID, dbStackResource.Status); serr != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update stack resource status: %v", serr)
		}
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

// computeStatusRewrite rebuilds the whole server-side status from the cluster
// CR. Build failures are owned by the imagebuild controller — the workload CR
// never reports them — so when the CR yields no failure of its own, an
// existing build_failure is carried over instead of being clobbered with nil.
// A CR-derived (runtime) failure still overwrites it, and clearing a build
// failure stays with the imagebuild controller, which clears it on build
// success.
func computeStatusRewrite(current *models.StackResourceStatus, clusterInstance *corev1alpha1.StackResource) *models.StackResourceStatus {
	updated := mapClusterStatusToServerStatus(clusterInstance)
	if updated.LastFailure == nil && current != nil && current.LastFailure != nil &&
		current.LastFailure.Type == models.FailureTypeBuildFailure {
		updated.LastFailure = current.LastFailure
	}
	return updated
}

func mapClusterStatusToServerStatus(clusterInstance *corev1alpha1.StackResource) *models.StackResourceStatus {
	res := &models.StackResourceStatus{
		LastObservedStatusHash: clusterInstance.Status.StatusHash,
		State:                  mapStackResourceState(clusterInstance.Status.Phase),
		Conditions:             models.ConvertConditions(clusterInstance.Status.Conditions),
		PublicIngresses:        mapToPublicIngresses(clusterInstance.Status.ExternalAddress),
		ObservedCrRevision:     clusterInstance.Status.ObservedRevision,
		InternalServiceName:    clusterInstance.Status.InternalAddress,
		LastFailure:            controllers.MapLastFailureDetails(clusterInstance.Name, clusterInstance.Status.LastFailureDetails),
		Replicas:               clusterInstance.Status.Replicas,
		AvailableReplicas:      clusterInstance.Status.AvailableReplicas,
		UpdatedReplicas:        clusterInstance.Status.UpdatedReplicas,
		LastRunSucceeded:       clusterInstance.Status.LastRunSucceeded,
	}
	if clusterInstance.Status.LastRestartRequestProcessedAt != nil {
		res.LastRestartRequestProcessedAt = ptr.To(clusterInstance.Status.LastRestartRequestProcessedAt.UTC())
	}
	if clusterInstance.Status.LastRunTime != nil {
		t := clusterInstance.Status.LastRunTime.Time
		res.LastRunTime = &t
	}
	return res
}

func mapStackResourceState(in corev1alpha1.StackResourcePhase) models.StackResourceState {
	switch in {
	case corev1alpha1.StackResourcePhasePending:
		return models.StackResourcePhasePending
	case corev1alpha1.StackResourcePhaseReady:
		return models.StackResourcePhaseReady
	case corev1alpha1.StackResourcePhaseFailed:
		return models.StackResourcePhaseFailed
	default:
		return models.StackResourcePhasePending
	}
}

func mapToPublicIngresses(externalAddresses []corev1alpha1.ExternalAddress) []models.Ingress {
	var publicIngresses []models.Ingress
	for _, externalAddress := range externalAddresses {
		publicIngresses = append(publicIngresses, models.Ingress{
			URL:        externalAddress.Address,
			TargetPort: int(externalAddress.TargetPort),
		})
	}
	return publicIngresses
}
