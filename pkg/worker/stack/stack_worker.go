package stack

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/builders"
	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/worker"
)

const (
	StackWorkerName = "stack-worker"
)

type stackWorker struct {
	stackService   stackService
	clusterManager clustermanager.ClusterManager
	secretService  secretService
	env            string
	subReconcilers []subReconciler
	worker.BaseWorker
}

type StackWorkerSpec struct {
	StackService         stackService
	SecretService        secretService
	VolumeService        volumeService
	CRBuilder            builders.ClusterResourceBuilder
	SecretBuilder        builders.SecretBuilder
	NamespaceService     namespaceService
	PostgresAddonService postgresAddonService
	ResourceUsageService resourceUsageService
	Env                  string
	ClusterManager       clustermanager.ClusterManager
}

func NewStackWorker(spec StackWorkerSpec) worker.Worker {
	return &stackWorker{
		stackService:   spec.StackService,
		clusterManager: spec.ClusterManager,
		secretService:  spec.SecretService,
		BaseWorker:     worker.NewBaseWorker(StackWorkerName, spec.Env),
		subReconcilers: []subReconciler{
			NewDeprovisionReconciler(DeprovisionReconcilerSpec{
				StackService:         spec.StackService,
				SecretService:        spec.SecretService,
				NamespaceService:     spec.NamespaceService,
				Logger:               logger.NewLoggerWithPrefix(context.Background(), "stack-deprovision-reconciler"),
				VolumeService:        spec.VolumeService,
				ClusterManager:       spec.ClusterManager,
				ResourceUsageService: spec.ResourceUsageService,
			}),
			NewValidationReconciler(ValidationReconcilerSpec{
				Logger:         logger.NewLoggerWithPrefix(context.Background(), "stack-validation-reconciler"),
				StackCRBuilder: spec.CRBuilder,
				SecretService:  spec.SecretService,
				StackService:   spec.StackService,
			}),
			NewNamespaceReconciler(NamespaceReconcilerSpec{
				ClusterManager:   spec.ClusterManager,
				NamespaceService: spec.NamespaceService,
			}),
			NewSecretReconciler(SecretReconcilerSpec{
				ClusterManager:       spec.ClusterManager,
				SecretService:        spec.SecretService,
				ResourceUsageService: spec.ResourceUsageService,
				Logger:               logger.NewLoggerWithPrefix(context.Background(), "stack-secret-reconciler"),
			}),
			NewConnectionReconciler(ConnectionReconcilerSpec{
				VolumeService: spec.VolumeService,
			}),
			NewVolumeReconciler(VolumeReconcilerSpec{
				ClusterManager:  spec.ClusterManager,
				VolumeService:   spec.VolumeService,
				VolumeCrBuilder: spec.CRBuilder,
			}),
			NewAddonEnvReconciler(AddonEnvReconcilerSpec{
				PostgresAddonService: spec.PostgresAddonService,
				SecretService:        spec.SecretService,
			}),
			NewStackReconciler(StackReconcilerSpec{
				ClusterManager: spec.ClusterManager,
				StackService:   spec.StackService,
				StackCRBuilder: spec.CRBuilder,
			}),
		},
	}
}

func (w *stackWorker) Execute(ctx context.Context, operand worker.Operand) (worker.Result, *errors.ServiceError) {
	stackID, ok := operand.(*models.Stack)
	if !ok {
		return worker.Result{}, w.WorkerError.NewError("invalid operand type, expected *models.Stack")
	}

	stack, err := w.stackService.InternalGetStack(ctx, stackID.ID)
	if err != nil {
		if err.Is404() {
			w.Logger().Infof("Stack %s not found, skipping reconciliation", stackID.ID)
			return worker.Result{}, nil
		}
		return worker.Result{}, err
	}
	w.Logger().Infof("Processing stack: %s", stack.ID)

	res, err := w.reconcile(ctx, stack)
	if err != nil {
		w.Logger().Errorf("Failed to reconcile stack %s: %v", stack.ID, err)
		return worker.Result{}, err
	}
	return res, nil
}

func (w *stackWorker) reconcile(ctx context.Context, stack *models.Stack) (worker.Result, *errors.ServiceError) {
	w.Logger().Infof("Reconciling stack: %s", stack.ID)

	for _, subReconciler := range w.subReconcilers {
		w.Logger().Info(ctx, "reconciling with sub-reconciler: %s", subReconciler.Name())
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
	res, err := w.stackService.InternalList(ctx, "status->>'state' IN ?", []models.StackState{
		models.StackPending,
		models.StackError,
		models.StackDeleting,
	})
	if err != nil {
		return nil, w.WorkerError.NewError("failed to list pending stacks: %v", err)
	}

	operands := make([]worker.Operand, 0)
	for _, stack := range res {
		operands = append(operands, &models.Stack{
			ID: stack.ID,
		})
	}
	return operands, nil
}
