package release

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/builders"
	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stackdeploy"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/worker"
	"github.com/Stackdome/stackdome/pkg/worker/releasegc"
	"github.com/Stackdome/stackdome/pkg/worker/workermanager"
)

const (
	ReleaseWorkerName       = "release-worker"
	convergencePollInterval = 15 * time.Second
)

type ReleaseWorkerSpec struct {
	ReleaseService        releaseService
	EventRecorder         eventRecorder
	StackService          stackService
	ClusterManager        clustermanager.ClusterManager
	CRBuilder             builders.ClusterResourceBuilder
	SecretBuilder         builders.SecretBuilder
	SecretService         secretService
	CredentialResolver    credentials.Resolver
	PostgresAddonService  postgresAddonService
	VolumeService         volumeService
	Resolver              *stackdeploy.Resolver
	ReleaseWorkerEnqueuer workermanager.BackgroundJobEnqueuer
	ValidationRecords     stores.ResourceValidationRecordStore
	RegistryClients       registryClientProvider
	Env                   string
}

type releaseWorker struct {
	releaseService        releaseService
	releaseWorkerEnqueuer workermanager.BackgroundJobEnqueuer
	subReconcilers        []subReconciler
	worker.BaseWorker
}

var _ worker.Worker = (*releaseWorker)(nil)

func NewReleaseWorker(spec ReleaseWorkerSpec) worker.Worker {
	return &releaseWorker{
		releaseService:        spec.ReleaseService,
		releaseWorkerEnqueuer: spec.ReleaseWorkerEnqueuer,
		subReconcilers: []subReconciler{
			newGatekeeperReconciler(spec),
			newSimulatorReconciler(spec),
			newValidationReconciler(spec),
			newRenderReconciler(spec),
			newApplyReconciler(spec),
			newConvergeReconciler(spec),
		},
		BaseWorker: worker.NewBaseWorker(ReleaseWorkerName, spec.Env),
	}
}

func (w *releaseWorker) Interval() time.Duration {
	return 30 * time.Second
}

func (w *releaseWorker) Execute(ctx context.Context, operand worker.Operand) (worker.Result, *errors.ServiceError) {
	releaseRef, ok := operand.(*models.StackRelease)
	if !ok {
		return worker.Result{}, w.WorkerError.NewError("invalid operand type, expected *models.StackRelease")
	}

	release, serr := w.releaseService.InternalGet(ctx, releaseRef.ID)
	if serr != nil {
		if serr.Is404() {
			w.Logger().Infof("release %s not found, skipping", releaseRef.ID)
			return worker.Result{}, nil
		}
		return worker.Result{}, serr
	}

	if release.State.Terminal() {
		return worker.Result{}, nil
	}

	w.Logger().Infof("processing release %s (stack=%s, state=%s)", release.ID, release.StackID, release.State)

	res, reconcileErr := w.reconcile(ctx, release)
	if reconcileErr != nil {
		w.Logger().Errorf("failed to reconcile release %s: %v", release.ID, reconcileErr)
		return worker.Result{}, w.WorkerError.NewError("failed to reconcile release %s: %v", release.ID, reconcileErr)
	}

	if w.releaseWorkerEnqueuer != nil {
		updated, _ := w.releaseService.InternalGet(ctx, releaseRef.ID)
		if updated != nil && updated.State.Terminal() {
			if err := w.releaseWorkerEnqueuer.Enqueue(&releasegc.ReleaseGCRequest{StackID: updated.StackID}); err != nil {
				w.Logger().Errorf("failed to enqueue release GC for stack %s: %v", updated.StackID, err)
			}
		}
	}

	return res, nil
}

func (w *releaseWorker) reconcile(ctx context.Context, release *models.StackRelease) (worker.Result, error) {
	for _, sr := range w.subReconcilers {
		w.Logger().Infof("running sub-reconciler: %s for release: %s", sr.Name(), release.ID)
		result, err := sr.Reconcile(ctx, release)
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

func (w *releaseWorker) GetInput(ctx context.Context) ([]worker.Operand, *errors.ServiceError) {
	// TODO: Add pagination..
	releases, serr := w.releaseService.InternalListActive(ctx)
	if serr != nil {
		return nil, w.WorkerError.NewError("failed to list active releases: %v", serr)
	}
	operands := make([]worker.Operand, 0, len(releases))
	for _, r := range releases {
		operands = append(operands, &models.StackRelease{ID: r.ID})
	}
	return operands, nil
}
