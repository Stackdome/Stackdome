package postgresaddon

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/Stackdome/stackdome/pkg/builders"
	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/worker"
)

const WorkerName = "postgres-addon-worker"

type postgresAddonWorker struct {
	postgresAddonService postgresAddonService
	subReconcilers       []subReconciler
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
	ClusterWrites        *worker.ClusterMutationCoordinator
}

func NewPostgresAddonWorker(spec PostgresAddonWorkerSpec) worker.Worker {
	return &postgresAddonWorker{
		postgresAddonService: spec.PostgresAddonService,
		BaseWorker: worker.NewBaseWorkerWithClusterMutationCoordinator(
			WorkerName, spec.Env, spec.ClusterWrites,
		),
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
	unlock := w.LockResource(addonRef.ID)
	defer unlock()

	addon, err := w.postgresAddonService.InternalGetPostgresAddon(ctx, addonRef.ID)
	if err != nil {
		if err.Is404() {
			w.Logger().Info(ctx, "PostgresAddon %s not found, skipping", addonRef.ID)
			return worker.Result{}, nil
		}
		return worker.Result{}, err
	}
	unlockCluster := w.LockClusterNamespace(addon.ClusterID, addon.Namespace)
	defer unlockCluster()

	w.Logger().Info(ctx, "Processing postgres addon: %s (%s)", addon.Name, addon.ID)
	authorizeMutation := worker.MutationAuthorizer(worker.AllowMutation)
	if addon.DeletionTimestamp == nil {
		authorizeMutation = w.authorizeMutation(addon.ID)
	}
	res, reconcileErr := w.reconcile(ctx, addon, authorizeMutation)
	if reconcileErr != nil {
		if stderrors.Is(reconcileErr, worker.ErrMutationNotAuthorized) {
			w.Logger().Debug(ctx, "Postgres addon desired state changed before cluster mutation")
			return worker.Result{}, nil
		}
		w.Logger().Error(ctx, "Failed to reconcile postgres addon %s: %v", addon.ID, reconcileErr)
		return worker.Result{}, w.WorkerError.NewError("failed to reconcile postgres addon %s: %v", addon.ID, reconcileErr)
	}
	return res, nil
}

func (w *postgresAddonWorker) authorizeMutation(addonID string) worker.MutationAuthorizer {
	return func(ctx context.Context) error {
		addon, serr := w.postgresAddonService.InternalGetPostgresAddon(ctx, addonID)
		if serr != nil {
			return serr
		}
		if addon.DeletionTimestamp != nil {
			return worker.ErrMutationNotAuthorized
		}
		return nil
	}
}

func (w *postgresAddonWorker) reconcile(ctx context.Context, addon *models.PostgresAddon, authorizeMutation worker.MutationAuthorizer) (worker.Result, error) {
	for _, sr := range w.subReconcilers {
		w.Logger().Info(ctx, "Running sub-reconciler: %s for addon: %s", sr.Name(), addon.ID)
		result, err := sr.Reconcile(ctx, addon, authorizeMutation)
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
