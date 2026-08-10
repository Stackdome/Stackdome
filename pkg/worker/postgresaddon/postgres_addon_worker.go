package postgresaddon

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/builders"
	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/Stackdome/stackdome/pkg/worker"
)

const WorkerName = "postgres-addon-worker"

type postgresAddonWorker struct {
	postgresAddonService postgresAddonService
	clusterManager       clustermanager.ClusterManager
	subReconcilers       []subReconciler
	runtimePolicy        services.RuntimePolicy
	worker.BaseWorker
}

type PostgresAddonWorkerSpec struct {
	PostgresAddonService postgresAddonService
	ObjectStoreService   objectStoreService
	NamespaceService     namespaceService
	SecretService        secretService
	ReferenceService     referenceService
	ClusterManager       clustermanager.ClusterManager
	CRBuilder            builders.PostgresClusterBuilder
	Env                  string
	RuntimePolicy        services.RuntimePolicy
}

func NewPostgresAddonWorker(spec PostgresAddonWorkerSpec) worker.Worker {
	if spec.RuntimePolicy == nil {
		panic("postgresaddon.NewPostgresAddonWorker: RuntimePolicy is required")
	}
	return &postgresAddonWorker{
		postgresAddonService: spec.PostgresAddonService,
		clusterManager:       spec.ClusterManager,
		runtimePolicy:        spec.RuntimePolicy,
		BaseWorker:           worker.NewBaseWorker(WorkerName, spec.Env),
		subReconcilers: []subReconciler{
			newDeprovisionReconciler(spec),
			newNamespaceReconciler(spec),
			newImageCatalogReconciler(spec),
			newObjectStoreDependencyReconciler(spec),
			newSecretReconciler(spec),
			newPostgresClusterReconciler(spec),
		},
	}
}

func (w *postgresAddonWorker) Interval() time.Duration {
	return 30 * time.Second
}

func (w *postgresAddonWorker) Execute(ctx context.Context, operand worker.Operand) (worker.Result, *errors.ServiceError) {
	addonRef, ok := operand.(models.PostgresAddonOperand)
	if !ok {
		return worker.Result{}, w.WorkerError.NewError("invalid operand type, expected models.PostgresAddonOperand")
	}

	addon, err := w.postgresAddonService.InternalGetPostgresAddon(ctx, addonRef.ID)
	if err != nil {
		if err.Is404() {
			w.Logger().Info(ctx, "PostgresAddon %s not found, skipping", addonRef.ID)
			return worker.Result{}, nil
		}
		return worker.Result{}, err
	}

	w.Logger().Info(ctx, "Processing postgres addon: %s (%s)", addon.Name, addon.ID)
	if addon.DeletionTimestamp == nil && w.runtimePolicy.DraftProvisioningMode() == services.ProvisioningModeDatabaseOnly {
		if admissionErr := w.runtimePolicy.RequireActiveAllocation(ctx, addon.OrganisationID); admissionErr != nil {
			if admissionErr.Reason == errors.ErrorCodeTrialInactive {
				return worker.Result{}, nil
			}
			return worker.Result{}, admissionErr
		}
	}

	res, reconcileErr := w.reconcile(ctx, addon)
	if reconcileErr != nil {
		w.Logger().Error(ctx, "Failed to reconcile postgres addon %s: %v", addon.ID, reconcileErr)
		return worker.Result{}, w.WorkerError.NewError("failed to reconcile postgres addon %s: %v", addon.ID, reconcileErr)
	}
	return res, nil
}

func (w *postgresAddonWorker) reconcile(ctx context.Context, addon *models.PostgresAddon) (worker.Result, error) {
	for _, sr := range w.subReconcilers {
		w.Logger().Info(ctx, "Running sub-reconciler: %s for addon: %s", sr.Name(), addon.ID)
		result, err := sr.Reconcile(ctx, addon)
		switch {
		case err != nil:
			return worker.Result{}, err
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

func (w *postgresAddonWorker) GetInput(ctx context.Context) ([]worker.Operand, *errors.ServiceError) {
	addons, err := w.postgresAddonService.InternalList(ctx,
		"status->>'state' IN ? OR deletion_timestamp IS NOT NULL",
		[]string{
			string(models.PostgresAddonStatePending),
			string(models.PostgresAddonStateError),
			string(models.PostgresAddonStateDeleting),
		},
	)
	if err != nil {
		return nil, w.WorkerError.NewError("failed to list pending postgres addons: %v", err)
	}

	operands := make([]worker.Operand, len(addons))
	for i, addon := range addons {
		operands[i] = models.PostgresAddonOperand{ID: addon.ID}
	}
	return operands, nil
}
