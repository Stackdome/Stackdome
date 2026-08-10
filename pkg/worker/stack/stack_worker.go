package stack

import (
	"context"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
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
	worker.BaseWorker
}

type StackWorkerSpec struct {
	StackService     stackService
	SecretService    secretService
	VolumeService    volumeService
	NamespaceService namespaceService
	Env              string
	ClusterManager   clustermanager.ClusterManager
	ReleaseService   releaseService
	ClusterWrites    *worker.ClusterMutationCoordinator
}

func NewStackWorker(spec StackWorkerSpec) worker.Worker {
	return &stackWorker{
		stackService:   spec.StackService,
		releaseService: spec.ReleaseService,
		clusterManager: spec.ClusterManager,
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
			NewResourceRestartReconciler(ResourceRestartReconcilerSpec{
				ClusterManager: spec.ClusterManager,
				Logger:         logger.NewLoggerWithPrefix(context.Background(), "stack-resource-restart-reconciler"),
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
	if stack.DeletionTimestamp == nil {
		authoritativeRelease, authorityErr := w.releaseService.InternalResolveAuthoritativeWorkloadRelease(ctx, stack)
		if authorityErr != nil {
			return worker.Result{}, authorityErr
		}
		if authoritativeRelease == nil || (stackRef.ReleaseID != "" && authoritativeRelease.ID != stackRef.ReleaseID) {
			log.Debug(ctx, "skipping stack without a released workload identity")
			return worker.Result{}, nil
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
