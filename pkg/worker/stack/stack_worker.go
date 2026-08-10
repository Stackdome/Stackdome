package stack

import (
	"context"
	stderrors "errors"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/Stackdome/stackdome/pkg/worker"
)

const (
	StackWorkerName = "stack-worker"
)

type stackWorker struct {
	stackService   stackService
	releaseService releaseService
	clusterManager clustermanager.ClusterManager
	subReconcilers []subReconciler
	runtimePolicy  services.RuntimePolicy
	worker.BaseWorker
}

type StackWorkerSpec struct {
	StackService     stackService
	SecretService    secretService
	VolumeService    volumeService
	NamespaceService namespaceService
	Env              string
	ClusterManager   clustermanager.ClusterManager
	RuntimePolicy    services.RuntimePolicy
	ReleaseService   releaseService
	ClusterWrites    *worker.ClusterMutationCoordinator
}

func NewStackWorker(spec StackWorkerSpec) worker.Worker {
	return &stackWorker{
		stackService:   spec.StackService,
		releaseService: spec.ReleaseService,
		clusterManager: spec.ClusterManager,
		runtimePolicy:  spec.RuntimePolicy,
		BaseWorker: worker.NewBaseWorkerWithClusterMutationCoordinator(
			StackWorkerName, spec.Env, spec.ClusterWrites,
		),
		subReconcilers: []subReconciler{
			NewDeprovisionReconciler(DeprovisionReconcilerSpec{
				StackService:     spec.StackService,
				SecretService:    spec.SecretService,
				NamespaceService: spec.NamespaceService,
				Logger:           logger.NewLoggerWithPrefix(context.Background(), "stack-deprovision-reconciler"),
				VolumeService:    spec.VolumeService,
				ClusterManager:   spec.ClusterManager,
			}),
			NewNamespaceReconciler(NamespaceReconcilerSpec{
				ClusterManager:   spec.ClusterManager,
				NamespaceService: spec.NamespaceService,
				Logger:           logger.NewLoggerWithPrefix(context.Background(), "stack-namespace-reconciler"),
			}),
		},
	}
}

func (w *stackWorker) Execute(ctx context.Context, operand worker.Operand) (worker.Result, *errors.ServiceError) {
	stackRef, ok := operand.(models.StackOperand)
	if !ok {
		return worker.Result{}, w.WorkerError.NewError("invalid operand type, expected models.StackOperand")
	}
	unlock := w.LockResource(stackRef.ID)
	defer unlock()

	log := w.Logger().WithField(logger.FieldStackID, stackRef.ID)

	stack, err := w.stackService.InternalGetStack(ctx, stackRef.ID)
	if err != nil {
		if err.Is404() {
			log.Info(ctx, "stack not found, skipping reconciliation")
			return worker.Result{}, nil
		}
		return worker.Result{}, err
	}
	unlockCluster := w.LockClusterNamespace(stack.ClusterID, stack.Namespace)
	defer unlockCluster()
	log.Info(ctx, "processing stack")
	releaseID := ""
	if stack.DeletionTimestamp == nil && w.runtimePolicy.DraftProvisioningMode() == services.ProvisioningModeDatabaseOnly {
		releaseID = stackRef.ReleaseID
		authoritativeRelease, authorityErr := w.releaseService.InternalResolveAuthoritativeWorkloadRelease(ctx, stack)
		if authorityErr != nil {
			return worker.Result{}, authorityErr
		}
		if authoritativeRelease == nil || (releaseID != "" && authoritativeRelease.ID != releaseID) {
			log.Debug(ctx, "skipping stack without a released workload identity")
			return worker.Result{}, nil
		}
		releaseID = authoritativeRelease.ID
		if admissionErr := w.runtimePolicy.RequireActiveAllocation(ctx, stack.OrganisationID); admissionErr != nil {
			if admissionErr.Reason == errors.ErrorCodeTrialInactive {
				log.Debug(ctx, "skipping draft provisioning without an active trial allocation")
				return worker.Result{}, nil
			}
			return worker.Result{}, admissionErr
		}
	}

	if stack.Annotations.ToMap()[models.SkipClusterProvisioningAnnotation] == "true" && stack.DeletionTimestamp == nil && w.Env == config.EnvironmentTest {
		log.Info(ctx, "skipping cluster provisioning due to annotation")
		return worker.Result{}, w.markAsReadyForSkippedClusterProvisioning(ctx, stack)
	}

	if releaseID != "" {
		// Queue retries can outlive a release. Re-read both records at the last
		// boundary before reconciliation so cancellation or supersession wins.
		currentStack, stackErr := w.stackService.InternalGetStack(ctx, stack.ID)
		if stackErr != nil {
			return worker.Result{}, stackErr
		}
		authoritativeRelease, authorityErr := w.releaseService.InternalResolveAuthoritativeWorkloadRelease(ctx, currentStack)
		if authorityErr != nil {
			return worker.Result{}, authorityErr
		}
		if authoritativeRelease == nil || authoritativeRelease.ID != releaseID {
			log.Debug(ctx, "release authority changed before stack reconciliation")
			return worker.Result{}, nil
		}
		stack = currentStack
	}

	authorizeMutation := worker.MutationAuthorizer(worker.AllowMutation)
	if releaseID != "" {
		authorizeMutation = w.authorizeMutation(stack.ID, releaseID)
	}
	res, err := w.reconcile(ctx, stack, authorizeMutation)
	if err != nil {
		if stderrors.Is(err, worker.ErrMutationNotAuthorized) {
			log.Debug(ctx, "release authority changed before cluster mutation")
			return worker.Result{}, nil
		}
		log.Error(ctx, "failed to reconcile stack: %v", err)
		return worker.Result{}, err
	}
	return res, nil
}

func (w *stackWorker) authorizeMutation(stackID, releaseID string) worker.MutationAuthorizer {
	return func(ctx context.Context) error {
		stack, serr := w.stackService.InternalGetStack(ctx, stackID)
		if serr != nil {
			return serr
		}
		if stack.DeletionTimestamp != nil {
			return worker.ErrMutationNotAuthorized
		}
		authoritativeRelease, serr := w.releaseService.InternalResolveAuthoritativeWorkloadRelease(ctx, stack)
		if serr != nil {
			return serr
		}
		if authoritativeRelease == nil || authoritativeRelease.ID != releaseID {
			return worker.ErrMutationNotAuthorized
		}
		if admissionErr := w.runtimePolicy.RequireActiveAllocation(ctx, stack.OrganisationID); admissionErr != nil {
			if admissionErr.Reason == errors.ErrorCodeTrialInactive {
				return worker.ErrMutationNotAuthorized
			}
			return admissionErr
		}
		return nil
	}
}

func (w *stackWorker) reconcile(ctx context.Context, stack *models.Stack, authorizeMutation worker.MutationAuthorizer) (worker.Result, *errors.ServiceError) {
	log := w.Logger().WithField(logger.FieldStackID, stack.ID)
	log.Info(ctx, "reconciling stack")

	for _, subReconciler := range w.subReconcilers {
		log.Debug(ctx, "reconciling with sub-reconciler: %s", subReconciler.Name())
		result, err := subReconciler.Reconcile(ctx, stack, authorizeMutation)
		switch {
		case stderrors.Is(err, worker.ErrMutationNotAuthorized):
			return worker.Result{}, nil
		case err != nil:
			return worker.Result{}, w.WorkerError.NewError("failed to reconcile stack %s with sub-reconciler %s: %v", stack.ID, subReconciler.Name(), err)
		case result.resultStop:
			return worker.Result{}, nil
		case result.resultRequeue:
			return worker.Result{Requeue: true}, nil
		case result.resultRequeueAfter != nil:
			return worker.Result{RequeueAfter: *result.resultRequeueAfter}, nil
		}
	}

	return worker.Result{}, nil
}

func (w *stackWorker) GetInput(ctx context.Context) ([]worker.Operand, *errors.ServiceError) {
	if w.runtimePolicy.DraftProvisioningMode() == services.ProvisioningModeDatabaseOnly {
		return w.getCloudInput(ctx)
	}
	res, err := w.stackService.InternalList(ctx, "status->>'state' IN ? OR deletion_timestamp IS NOT NULL",
		[]models.StackState{
			models.StackPending,
			models.StackDeleting,
		})
	if err != nil {
		return nil, w.WorkerError.NewError("failed to list pending stacks: %v", err)
	}

	operands := make([]worker.Operand, 0)
	for _, stack := range res {
		operands = append(operands, models.StackOperand{ID: stack.ID})
	}
	return operands, nil
}

func (w *stackWorker) getCloudInput(ctx context.Context) ([]worker.Operand, *errors.ServiceError) {
	scan, serr := w.releaseService.InternalListAuthoritativeWorkload(ctx)
	if serr != nil {
		return nil, w.WorkerError.NewError("failed to list authoritative stack releases: %v", serr)
	}
	operands := make([]worker.Operand, 0, len(scan.Releases)+len(scan.DeletingStackIDs))
	for _, release := range scan.Releases {
		operands = append(operands, models.StackOperand{ID: release.StackID, ReleaseID: release.ReleaseID})
	}
	for _, stackID := range scan.DeletingStackIDs {
		operands = append(operands, models.StackOperand{ID: stackID})
	}
	return operands, nil
}

func (w *stackWorker) markAsReadyForSkippedClusterProvisioning(ctx context.Context, stack *models.Stack) *errors.ServiceError {
	if stack.Status == nil {
		stack.Status = &models.StackStatus{}
	}
	stack.Status.State = models.StackReady
	stack.Status.Message = "Stack is ready"
	stack.Status.ObservedCrRevision = stack.CrRevision
	stack.Status.LastObservedStatusHash = "computed-hash"
	return w.stackService.UpdateStatus(ctx, stack.ID, stack.Status)
}
