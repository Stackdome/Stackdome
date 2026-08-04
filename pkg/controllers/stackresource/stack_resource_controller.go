package stackresource

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Stackdome/stackdome/pkg/controllers"
	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

//go:generate mockgen -source=stack_resource_controller.go -destination=stack_resource_controller_mock.go -package=stackresource

type releaseResolver interface {
	InternalGet(ctx context.Context, releaseID string) (*models.StackRelease, *apperrors.ServiceError)
	InternalGetActiveByStackID(ctx context.Context, stackID string) (*models.StackRelease, *apperrors.ServiceError)
	InternalGetLatestByStackID(ctx context.Context, stackID string) (*models.StackRelease, *apperrors.ServiceError)
}

type resourceEventRecorder interface {
	RecordResourceEvent(ctx context.Context, release *models.StackRelease, resourceName string, eventType models.ReleaseEventType, reason string, message string) *apperrors.ServiceError
}

type stackResourceReconciler struct {
	client               client.Client
	stackResourceService services.StackResourceService
	stackService         services.StackService
	releaseResolver      releaseResolver
	eventRecorder        resourceEventRecorder
	logger               logger.Logger
	env                  string
}

type StackResourceReconcilerSpec struct {
	StackResourceService services.StackResourceService
	StackService         services.StackService
	ReleaseResolver      releaseResolver
	EventRecorder        resourceEventRecorder
	Log                  logger.Logger
	Env                  string
}

func NewStackResourceReconciler(spec StackResourceReconcilerSpec) *stackResourceReconciler {
	return &stackResourceReconciler{
		client:               nil,
		stackResourceService: spec.StackResourceService,
		stackService:         spec.StackService,
		releaseResolver:      spec.ReleaseResolver,
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
	w.logger.Info(ctx, "Reconciling stack resource: %v", req.NamespacedName)

	stackResourceCr := &corev1alpha1.StackResource{}
	if err := w.client.Get(ctx, req.NamespacedName, stackResourceCr); err != nil {
		if errors.IsNotFound(err) {
			w.logger.Info(ctx, "StackResource %v not found", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	stackID, ok := stackResourceCr.Labels[corev1alpha1.LabelStackID]
	if !ok {
		// How are we here? The predicate should have prevented this.
		w.logger.Error(ctx, "StackResource %s does not have a stack id label", stackResourceCr.Name)
		return ctrl.Result{}, nil
	}

	dbStackResource, serr := w.stackResourceService.InternalGetByStackIDAndResourceName(ctx, stackID, stackResourceCr.Name)
	if serr != nil {
		if serr.Code == apperrors.ErrorNotFound {
			w.logger.Info(ctx, "StackResource %s not found in DB", stackResourceCr.Name)
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get stack resource from db: %w", serr)
	}

	if dbStackResource.Status == nil || dbStackResource.Status.LastObservedStatusHash != stackResourceCr.Status.StatusHash {
		dbStackResource.Status = computeStatusRewrite(dbStackResource.Status, stackResourceCr)
		if serr := w.stackResourceService.UpdateStatus(ctx, dbStackResource.ID, dbStackResource.Status); serr != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update stack resource status: %w", serr)
		}
		w.logger.WithFields(map[string]interface{}{
			logger.FieldStackID:  stackID,
			logger.FieldResource: stackResourceCr.Name,
		}).Debug(ctx, "synced stack resource status from cluster")
		w.recordResourceEvent(ctx, stackID, dbStackResource, stackResourceCr)
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

// releaseEvent is one timeline event mapped from the CR, ready to record.
// The mapping functions (resourceEvent, portsClosedEvent) return nil when
// the CR's state maps to nothing.
type releaseEvent struct {
	Type    models.ReleaseEventType
	Reason  string
	Message string
}

// recordResourceEvent puts the resource's state on a release timeline.
//
//   - Runs only when StatusHash changed, so each state is recorded once.
//   - The CR's release annotation says which release the state belongs to.
//   - A failed record: log and move on. The status is already saved.
func (w *stackResourceReconciler) recordResourceEvent(ctx context.Context, stackID string, resource *models.StackResource, cr *corev1alpha1.StackResource) {
	workload := resourceEvent(cr)
	ports := portsClosedEvent(cr)
	tls := tlsEvent(cr)
	if workload == nil && ports == nil && tls == nil {
		return
	}

	release := w.resolveRelease(ctx, stackID, cr)
	if release == nil {
		return
	}

	w.recordOnRelease(ctx, release, resource.Name, workload)
	w.recordOnRelease(ctx, release, resource.Name, ports)
	w.recordOnRelease(ctx, release, resource.Name, tls)
}

func (w *stackResourceReconciler) resolveRelease(ctx context.Context, stackID string, cr *corev1alpha1.StackResource) *models.StackRelease {
	return controllers.ResolveEventRelease(ctx, w.releaseResolver, w.logger, stackID, cr.Annotations[corev1alpha1.ReleaseIDAnnotation])
}

func (w *stackResourceReconciler) recordOnRelease(ctx context.Context, release *models.StackRelease, resourceName string, ev *releaseEvent) {
	if ev == nil {
		return
	}
	if recErr := w.eventRecorder.RecordResourceEvent(ctx, release, resourceName, ev.Type, ev.Reason, ev.Message); recErr != nil {
		w.logger.Error(ctx, "failed to record %s event for %s: %v", ev.Type, resourceName, recErr)
	}
}

// currentGenCondition returns the condition only when it was written for the
// CR's current generation; a condition carrying a prior generation describes
// a superseded rollout and must not produce events for this one. Computed
// from the raw CR because the server-side condition model drops
// ObservedGeneration.
func currentGenCondition(cr *corev1alpha1.StackResource, condType string) *metav1.Condition {
	cond := meta.FindStatusCondition(cr.Status.Conditions, condType)
	if cond == nil || cond.ObservedGeneration != cr.Generation {
		return nil
	}
	return cond
}

// condIs returns the current-generation condition only when it has the given
// status; nil otherwise.
func condIs(cr *corev1alpha1.StackResource, condType string, status metav1.ConditionStatus) *metav1.Condition {
	cond := currentGenCondition(cr, condType)
	if cond == nil || cond.Status != status {
		return nil
	}
	return cond
}

// portsClosedEvent flags a converged, Available resource whose last dial found
// a declared port closed. Scoped to that state on purpose: while the rollout is
// still converging, resourceEvent's deploying branch already reports the same
// dial.
func portsClosedEvent(cr *corev1alpha1.StackResource) *releaseEvent {
	if cr.Status.ObservedGeneration != cr.Generation {
		return nil
	}
	if cr.Status.PortCheck == nil || cr.Status.PortCheck.Status != corev1alpha1.PortCheckStatusTypeFailure {
		return nil
	}
	if condIs(cr, string(corev1alpha1.StackResourceConverged), metav1.ConditionTrue) == nil ||
		condIs(cr, string(corev1alpha1.StackResourceStatusAvailable), metav1.ConditionTrue) == nil {
		return nil
	}
	return &releaseEvent{
		models.ReleaseEventTypeResourcePortsClosed,
		controllers.ReasonPortNotListening,
		portDialMessage(cr.Status.PortCheck.FailingPortNumbers),
	}
}

// tlsEvent maps the current-generation TLSConfigured condition to a TLS event.
// Pre-0.6.11 agents set True on issuer discovery alone, so True is keyed on the
// reason. Any other False reason is a terminal issuance failure (the site
// serves plain HTTP).
func tlsEvent(cr *corev1alpha1.StackResource) *releaseEvent {
	cond := currentGenCondition(cr, string(corev1alpha1.StackResourceTLSConfigured))
	if cond == nil {
		return nil
	}
	switch {
	case cond.Status == metav1.ConditionTrue && cond.Reason == controllers.ReasonTLSReady:
		return &releaseEvent{models.ReleaseEventTypeResourceTLSReady, cond.Reason, cond.Message}
	case cond.Status == metav1.ConditionFalse && cond.Reason == controllers.ReasonCertificateIssuing:
		return &releaseEvent{models.ReleaseEventTypeResourceTLSIssuing, cond.Reason, cond.Message}
	case cond.Status == metav1.ConditionFalse:
		return &releaseEvent{models.ReleaseEventTypeResourceTLSFailed, cond.Reason, cond.Message}
	}
	return nil
}

// resourceEvent maps the agent's summary verdict to one workload event.
//   - Stale or absent summary: superseded rollout (or an agent without the
//     field, a wiring bug) — nothing to record.
//   - Building: build events narrate that phase.
//   - Failed with ReasonBuildFailed: build_failed is already on the timeline.
func resourceEvent(cr *corev1alpha1.StackResource) *releaseEvent {
	s := cr.Status.Summary
	if s == nil || s.ObservedGeneration != cr.Generation {
		return nil
	}
	switch s.State {
	case corev1alpha1.SummaryStateWaiting:
		return &releaseEvent{models.ReleaseEventTypeResourceWaiting, s.Reason, s.Message}
	case corev1alpha1.SummaryStateDeploying:
		return &releaseEvent{models.ReleaseEventTypeResourceDeploying, s.Reason, s.Message}
	case corev1alpha1.SummaryStateReady:
		return &releaseEvent{models.ReleaseEventTypeResourceReady, s.Reason, s.Message}
	case corev1alpha1.SummaryStateFailed:
		if s.Reason == corev1alpha1.ReasonBuildFailed {
			return nil
		}
		return &releaseEvent{models.ReleaseEventTypeResourceFailed, s.Reason, s.Message}
	}
	return nil
}

// portDialMessage names the declared ports the last dial proved closed.
func portDialMessage(ports []int32) string {
	if len(ports) == 0 {
		return "declared ports not accepting connections"
	}
	strs := make([]string, len(ports))
	for i, p := range ports {
		strs[i] = strconv.Itoa(int(p))
	}
	noun := "port"
	if len(ports) > 1 {
		noun = "ports"
	}
	return fmt.Sprintf("%s %s not accepting connections", noun, strings.Join(strs, ", "))
}

// computeStatusRewrite rebuilds the server-side status from the cluster CR.
//
// The workload CR never reports build failures — those belong to the
// imagebuild controller. So:
//   - CR has no failure: keep an existing build_failure, don't clobber it.
//   - CR has a runtime failure: it wins.
//   - Clearing a build_failure is the imagebuild controller's job, on success.
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
		LastFailure:            controllers.MapLastFailureDetails(clusterInstance.Name, clusterInstance.Annotations[corev1alpha1.ReleaseIDAnnotation], clusterInstance.Status.LastFailureDetails),
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
