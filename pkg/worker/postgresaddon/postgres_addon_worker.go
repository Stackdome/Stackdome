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
	referenceService     referenceService
	releaseService       releaseService
	stackService         stackService
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
	ReleaseService       releaseService
	StackService         stackService
}

func NewPostgresAddonWorker(spec PostgresAddonWorkerSpec) worker.Worker {
	return &postgresAddonWorker{
		postgresAddonService: spec.PostgresAddonService,
		clusterManager:       spec.ClusterManager,
		runtimePolicy:        spec.RuntimePolicy,
		referenceService:     spec.ReferenceService,
		releaseService:       spec.ReleaseService,
		stackService:         spec.StackService,
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
	authorizingReleaseID := ""
	if addon.DeletionTimestamp == nil && w.runtimePolicy.DraftProvisioningMode() == services.ProvisioningModeDatabaseOnly {
		var authorized bool
		var authorizationErr *errors.ServiceError
		authorizingReleaseID, authorized, authorizationErr = w.authorizedByRelease(ctx, addon, addonRef.ReleaseID)
		if authorizationErr != nil {
			return worker.Result{}, authorizationErr
		}
		if !authorized {
			return worker.Result{}, nil
		}
		if admissionErr := w.runtimePolicy.RequireActiveAllocation(ctx, addon.OrganisationID); admissionErr != nil {
			if admissionErr.Reason == errors.ErrorCodeTrialInactive {
				return worker.Result{}, nil
			}
			return worker.Result{}, admissionErr
		}
	}
	if authorizingReleaseID != "" {
		authorized, authorizationErr := w.releaseRemainsAuthoritative(ctx, addon, authorizingReleaseID)
		if authorizationErr != nil {
			return worker.Result{}, authorizationErr
		}
		if !authorized {
			return worker.Result{}, nil
		}
	}

	res, reconcileErr := w.reconcile(ctx, addon)
	if reconcileErr != nil {
		w.Logger().Error(ctx, "Failed to reconcile postgres addon %s: %v", addon.ID, reconcileErr)
		return worker.Result{}, w.WorkerError.NewError("failed to reconcile postgres addon %s: %v", addon.ID, reconcileErr)
	}
	return res, nil
}

func (w *postgresAddonWorker) authorizedByRelease(ctx context.Context, addon *models.PostgresAddon, requestedReleaseID string) (string, bool, *errors.ServiceError) {
	if requestedReleaseID == "" {
		inUse, refs, serr := w.referenceService.IsReferentInUse(ctx, models.ReferentPostgresAddon, addon.ID)
		if serr != nil {
			return "", false, serr
		}
		if !inUse {
			return "", false, nil
		}
		for _, ref := range refs {
			if ref.ReleaseID == nil || *ref.ReleaseID == "" {
				continue
			}
			authorized, authorizationErr := w.releaseRemainsAuthoritative(ctx, addon, *ref.ReleaseID)
			if authorizationErr != nil {
				return "", false, authorizationErr
			}
			if authorized {
				return *ref.ReleaseID, true, nil
			}
		}
		return "", false, nil
	}

	authorized, serr := w.releaseRemainsAuthoritative(ctx, addon, requestedReleaseID)
	return requestedReleaseID, authorized, serr
}

func (w *postgresAddonWorker) releaseRemainsAuthoritative(ctx context.Context, addon *models.PostgresAddon, releaseID string) (bool, *errors.ServiceError) {
	release, serr := w.releaseService.InternalGet(ctx, releaseID)
	if serr != nil {
		return false, serr
	}
	stack, serr := w.stackService.InternalGetStack(ctx, release.StackID)
	if serr != nil {
		return false, serr
	}
	var active *models.StackRelease
	if release.State == models.ReleaseStatePending || release.State == models.ReleaseStateInProgress {
		active, serr = w.releaseService.InternalGetActiveByStackID(ctx, release.StackID)
		if serr != nil {
			return false, serr
		}
	}
	if !release.IsAuthoritativeWorkloadRelease(stack, active) {
		return false, nil
	}
	if release.Snapshot.Stack.OrganisationID != addon.OrganisationID {
		return false, w.WorkerError.NewError("release %s does not authorize postgres addon %s", release.ID, addon.ID)
	}
	for _, connection := range release.Snapshot.Connections {
		if (connection.From.Type == models.TopologyNodeTypePostgresAddon && connection.From.Id == addon.ID) ||
			(connection.To.Type == models.TopologyNodeTypePostgresAddon && connection.To.Id == addon.ID) {
			return true, nil
		}
	}
	return false, nil
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
