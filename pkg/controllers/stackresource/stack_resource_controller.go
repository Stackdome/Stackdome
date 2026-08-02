package stackresource

import (
	"context"
	"fmt"

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

type releaseActiveChecker interface {
	InternalGetActiveByStackID(ctx context.Context, stackID string) (*models.StackRelease, *apperrors.ServiceError)
}

type resourceEventRecorder interface {
	RecordResourceEvent(ctx context.Context, release *models.StackRelease, resourceName string, eventType models.ReleaseEventType, reason string, message string) *apperrors.ServiceError
}

// stalledReasonBuildFailed mirrors the Stalled condition reason the cluster
// agent sets on a terminal build failure (image_build_reconciler.go); the
// agent does not export its reason strings.
const stalledReasonBuildFailed = "BuildFailed"

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

// recordResourceEvent records the observed resource state onto the active
// release timeline. It runs only inside the StatusHash-change gate, so each
// state transition is recorded at most once. A missing active release or a
// recorder failure is non-fatal: the status update has already been persisted,
// so both paths only log and return.
func (w *stackResourceReconciler) recordResourceEvent(ctx context.Context, stackID string, resource *models.StackResource, cr *corev1alpha1.StackResource) {
	if resource.Status == nil {
		return
	}
	eventType, reason, message, emit := resourceEvent(resource.Status.Conditions, resource.Status.LastFailure, convergedForGeneration(cr), stalledForGeneration(cr), observedForGeneration(cr))
	if !emit {
		return
	}
	active, serr := w.releaseChecker.InternalGetActiveByStackID(ctx, stackID)
	if serr != nil || active == nil {
		w.logger.Debug(ctx, "no active release for stack %s; skipping resource event for %s", stackID, resource.Name)
		return
	}
	if recErr := w.eventRecorder.RecordResourceEvent(ctx, active, resource.Name, eventType, reason, message); recErr != nil {
		w.logger.Error(ctx, "failed to record resource event for %s: %v", resource.Name, recErr)
	}
}

// runtimeFailureDetail returns the reason code and long message of a runtime
// crash, main container first, then init container. Build failures are the
// imagebuild controller's to surface, so they are never read here.
func runtimeFailureDetail(f *models.StackResourceFailure) (reason, message string) {
	if f == nil || f.Type != models.FailureTypeRuntimeCrash {
		return "", ""
	}
	for _, d := range []*models.ContainerFailureDetail{f.Container, f.InitContainer} {
		if d == nil {
			continue
		}
		if d.Reason != "" || d.Message != "" {
			return d.Reason, d.Message
		}
	}
	return "", ""
}

// convergedForGeneration mirrors the cluster agent's isResourceConverged: the
// Converged condition only counts when it was written for the CR's current
// generation, otherwise it is a leftover from a previous rollout. Computed
// from the raw CR because the server-side condition model drops
// ObservedGeneration.
func convergedForGeneration(cr *corev1alpha1.StackResource) bool {
	cond := meta.FindStatusCondition(cr.Status.Conditions, string(corev1alpha1.StackResourceConverged))
	return cond != nil && cond.Status == metav1.ConditionTrue && cond.ObservedGeneration == cr.Generation
}

// stalledForGeneration reports whether the Stalled condition was written for the
// CR's current generation. Like convergedForGeneration it reads the raw CR
// because the server-side condition model drops ObservedGeneration. The agent
// does not rewrite Stalled on a NotReady report, so a stale Stalled=True can
// outlive its generation even after the top-level Status.ObservedGeneration
// advances — the condition's own ObservedGeneration is the authoritative signal.
func stalledForGeneration(cr *corev1alpha1.StackResource) bool {
	cond := meta.FindStatusCondition(cr.Status.Conditions, string(corev1alpha1.StackResourceStalled))
	return cond != nil && cond.ObservedGeneration == cr.Generation
}

// observedForGeneration reports whether the CR's status was written for its
// current generation. The cluster agent stamps Status.ObservedGeneration on
// every terminal report, so a status still carrying a prior generation's value
// is a leftover from a superseded rollout. Used to gate the runtime-failure
// path, whose LastFailureDetails carry no generation of their own.
func observedForGeneration(cr *corev1alpha1.StackResource) bool {
	return cr.Status.ObservedGeneration == cr.Generation
}

// resourceEvent maps the cluster-agent's status conditions (and any captured
// runtime crash detail) to a release timeline event. Conditions are the
// authoritative signal — Phase is a coarse rollup the agent writes alongside
// them. emit is false when the imagebuild controller's build events already
// cover the state (build in progress, terminal build failure) or no condition
// maps to an event.
//
// Available=True does NOT imply the rollout landed: the agent also reports it
// when only the previous revision is still serving (Phase=Degraded). Ready
// therefore additionally requires converged, and the not-converged path keeps
// the Converged condition's detail so a stuck rollout is diagnosable from the
// timeline — or the agent's readiness diagnosis when it has one, since the
// condition only ever says "not ready yet".
//
// A failure carried over from a prior generation must not be attributed to the
// new release, so every failure path is gated the same way the Ready path is
// gated on convergedForGeneration: the Stalled condition only counts for its own
// current generation (stalledCurrentGeneration), and the runtime crash detail
// and the readiness diagnosis only count when the status as a whole is for the
// current generation (runtimeCurrentGeneration). A terminal build failure is
// suppressed regardless of generation because those events are always the
// imagebuild controller's.
func resourceEvent(conditions []models.Condition, failure *models.StackResourceFailure, converged, stalledCurrentGeneration, runtimeCurrentGeneration bool) (eventType models.ReleaseEventType, reason, message string, emit bool) {
	if cond := models.FindCondition(conditions, string(corev1alpha1.StackResourceStalled)); cond != nil && cond.Status == string(models.ConditionTrue) {
		if cond.Reason == stalledReasonBuildFailed {
			return "", "", "", false
		}
		if stalledCurrentGeneration {
			return models.ReleaseEventTypeResourceFailed, cond.Reason, cond.Message, true
		}
	}
	if cond := models.FindCondition(conditions, string(corev1alpha1.StackResourceDependenciesReady)); cond != nil && cond.Status == string(models.ConditionFalse) {
		return models.ReleaseEventTypeResourceWaiting, cond.Reason, cond.Message, true
	}
	if cond := models.FindCondition(conditions, string(corev1alpha1.StackResourceBuildReady)); cond != nil && cond.Status == string(models.ConditionFalse) {
		return "", "", "", false
	}
	if r, m := runtimeFailureDetail(failure); (r != "" || m != "") && runtimeCurrentGeneration {
		return models.ReleaseEventTypeResourceFailed, r, m, true
	}
	available := models.FindCondition(conditions, string(corev1alpha1.StackResourceStatusAvailable))
	availableTrue := available != nil && available.Status == string(models.ConditionTrue)
	convergedCond := models.FindCondition(conditions, string(corev1alpha1.StackResourceConverged))
	if availableTrue && converged {
		return models.ReleaseEventTypeResourceReady, "", "", true
	}
	if convergedCond != nil && !converged {
		if convergedCond.Status == string(models.ConditionFalse) {
			reason, message = convergedCond.Reason, convergedCond.Message
		}
		if runtimeCurrentGeneration && failure != nil && failure.Type == models.FailureTypeReadinessFailure && failure.Container != nil {
			reason, message = failure.Container.Reason, failure.Container.Message
		}
		if availableTrue {
			if message == "" {
				message = "previous revision still serving"
			} else {
				message = "previous revision still serving; " + message
			}
		}
		return models.ReleaseEventTypeResourceDeploying, reason, message, true
	}
	return "", "", "", false
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
