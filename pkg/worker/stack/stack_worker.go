package stack

import (
	"context"

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
}

func NewStackWorker(spec StackWorkerSpec) worker.Worker {
	if spec.RuntimePolicy == nil {
		panic("stack.NewStackWorker: RuntimePolicy is required")
	}
	return &stackWorker{
		stackService:   spec.StackService,
		clusterManager: spec.ClusterManager,
		runtimePolicy:  spec.RuntimePolicy,
		BaseWorker:     worker.NewBaseWorker(StackWorkerName, spec.Env),
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
	stackID, ok := operand.(models.StackOperand)
	if !ok {
		return worker.Result{}, w.WorkerError.NewError("invalid operand type, expected models.StackOperand")
	}

	log := w.Logger().WithField(logger.FieldStackID, stackID.ID)

	stack, err := w.stackService.InternalGetStack(ctx, stackID.ID)
	if err != nil {
		if err.Is404() {
			log.Info(ctx, "stack not found, skipping reconciliation")
			return worker.Result{}, nil
		}
		return worker.Result{}, err
	}
	log.Info(ctx, "processing stack")
	if stack.DeletionTimestamp == nil && w.runtimePolicy.DraftProvisioningMode() == services.ProvisioningModeDatabaseOnly {
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

	res, err := w.reconcile(ctx, stack)
	if err != nil {
		log.Error(ctx, "failed to reconcile stack: %v", err)
		return worker.Result{}, err
	}
	return res, nil
}

func (w *stackWorker) reconcile(ctx context.Context, stack *models.Stack) (worker.Result, *errors.ServiceError) {
	log := w.Logger().WithField(logger.FieldStackID, stack.ID)
	log.Info(ctx, "reconciling stack")

	for _, subReconciler := range w.subReconcilers {
		log.Debug(ctx, "reconciling with sub-reconciler: %s", subReconciler.Name())
		result, err := subReconciler.Reconcile(ctx, stack)
		switch {
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
