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
	releaseChecker       releaseActiveChecker
	eventRecorder        resourceEventRecorder
	logger               logger.Logger
	env                  string
}

type StackResourceReconcilerSpec struct {
	StackResourceService services.StackResourceService
	StackService         services.StackService
	ReleaseChecker       releaseActiveChecker
	EventRecorder        resourceEventRecorder
	Log                  logger.Logger
	Env                  string
}

func NewStackResourceReconciler(spec StackResourceReconcilerSpec) *stackResourceReconciler {
	if spec.ReleaseChecker == nil {
		panic("StackResourceReconciler requires a ReleaseChecker")
	}
	if spec.EventRecorder == nil {
		panic("StackResourceReconciler requires an EventRecorder")
	}
	return &stackResourceReconciler{
		client:               nil,
		stackResourceService: spec.StackResourceService,
		stackService:         spec.StackService,
		releaseChecker:       spec.ReleaseChecker,
		eventRecorder:        spec.EventRecorder,
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
		w.recordResourceEvent(ctx, stackID, dbStackResource)
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

// recordResourceEvent records the observed resource state onto the active
// release timeline. It runs only inside the StatusHash-change gate, so each
// state transition is recorded at most once. A missing active release or a
// recorder failure is non-fatal: the status update has already been persisted,
// so both paths only log and return.
func (w *stackResourceReconciler) recordResourceEvent(ctx context.Context, stackID string, resource *models.StackResource) {
	if resource.Status == nil {
		return
	}
	eventType, reason, emit := resourceEventForState(resource.Status.State, resource.Status.Conditions, lastFailureReason(resource.Status.LastFailure))
	if !emit {
		return
	}
	active, serr := w.releaseChecker.InternalGetActiveByStackID(ctx, stackID)
	if serr != nil || active == nil {
		w.logger.Debugf("no active release for stack %s; skipping resource event for %s", stackID, resource.Name)
		return
	}
	if recErr := w.eventRecorder.RecordResourceEvent(ctx, active, resource.Name, eventType, reason); recErr != nil {
		w.logger.Errorf("failed to record resource event for %s: %v", resource.Name, recErr)
	}
}

// lastFailureReason extracts the most user-meaningful reason from a resource's
// last failure. It prefers the human-readable Message and falls back to the
// shorter Reason code, checking the main container first, then the init
// container, then a build failure. Returns "" when no failure detail carries a
// reason (a truthful empty reason).
func lastFailureReason(f *models.StackResourceFailure) string {
	if f == nil {
		return ""
	}
	for _, d := range []*models.ContainerFailureDetail{f.Container, f.InitContainer} {
		if d == nil {
			continue
		}
		if d.Message != "" {
			return d.Message
		}
		if d.Reason != "" {
			return d.Reason
		}
	}
	if f.Build != nil {
		if f.Build.Message != "" {
			return f.Build.Message
		}
		if f.Build.Reason != "" {
			return f.Build.Reason
		}
	}
	return ""
}

// resourceEventForState maps an observed resource state and the cluster-agent
// conditions on it to a release timeline event. emit is false when no resource
// event should be recorded: a build in progress (the imagebuild controller's
// build events cover it) or an unmapped state. failureReason is the reason
// derived from the resource's last failure, used on the failed path unless a
// Stalled=True condition supplies a more specific message.
func resourceEventForState(state models.StackResourceState, conditions []models.Condition, failureReason string) (eventType models.ReleaseEventType, reason string, emit bool) {
	switch state {
	case models.StackResourcePhasePending:
		if cond := models.FindCondition(conditions, string(corev1alpha1.StackResourceDependenciesReady)); cond != nil && cond.Status == string(models.ConditionFalse) {
			return models.ReleaseEventTypeResourceWaiting, cond.Message, true
		}
		if cond := models.FindCondition(conditions, string(corev1alpha1.StackResourceBuildReady)); cond != nil && cond.Status == string(models.ConditionFalse) {
			// Building — the imagebuild controller's build events cover this.
			return "", "", false
		}
		return models.ReleaseEventTypeResourceDeploying, "", true
	case models.StackResourcePhaseReady:
		return models.ReleaseEventTypeResourceReady, "", true
	case models.StackResourcePhaseFailed, models.StackResourcePhaseUnknown:
		reason = failureReason
		if cond := models.FindCondition(conditions, string(corev1alpha1.StackResourceStalled)); cond != nil && cond.Status == string(models.ConditionTrue) && cond.Message != "" {
			reason = cond.Message
		}
		return models.ReleaseEventTypeResourceFailed, reason, true
	default:
		return "", "", false
	}
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
