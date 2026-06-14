package release

import (
	"context"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/builders"
	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stackdeploy"
	"github.com/ashishmax31/stackdome-api-server/pkg/worker"
)

const (
	ReleaseWorkerName       = "release-worker"
	convergenceTimeout      = 15 * time.Minute
	convergencePollInterval = 15 * time.Second
)

type ReleaseWorkerSpec struct {
	ReleaseService       releaseService
	StackService         stackService
	ClusterManager       clustermanager.ClusterManager
	CRBuilder            builders.ClusterResourceBuilder
	SecretBuilder        builders.SecretBuilder
	SecretService        secretService
	PostgresAddonService postgresAddonService
	VolumeService        volumeService
	Resolver             *stackdeploy.Resolver
	Env                  string
}

type releaseWorker struct {
	releaseService       releaseService
	stackService         stackService
	clusterManager       clustermanager.ClusterManager
	crBuilder            builders.ClusterResourceBuilder
	secretBuilder        builders.SecretBuilder
	secretService        secretService
	postgresAddonService postgresAddonService
	volumeService        volumeService
	resolver             *stackdeploy.Resolver
	worker.BaseWorker
}

var _ worker.Worker = (*releaseWorker)(nil)

func NewReleaseWorker(spec ReleaseWorkerSpec) worker.Worker {
	return &releaseWorker{
		releaseService:       spec.ReleaseService,
		stackService:         spec.StackService,
		clusterManager:       spec.ClusterManager,
		crBuilder:            spec.CRBuilder,
		secretBuilder:        spec.SecretBuilder,
		secretService:        spec.SecretService,
		postgresAddonService: spec.PostgresAddonService,
		volumeService:        spec.VolumeService,
		resolver:             spec.Resolver,
		BaseWorker:           worker.NewBaseWorker(ReleaseWorkerName, spec.Env),
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

	switch release.State {
	case models.ReleaseStatePending, models.ReleaseStateRendering:
		return w.render(ctx, release)
	case models.ReleaseStateApplying:
		return w.applyAndConverge(ctx, release)
	default:
		return worker.Result{}, nil
	}
}

func (w *releaseWorker) GetInput(ctx context.Context) ([]worker.Operand, *errors.ServiceError) {
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

func (w *releaseWorker) fail(ctx context.Context, release *models.StackRelease, msg string) {
	w.Logger().Errorf("release %s failed: %s", release.ID, msg)
	if _, err := w.releaseService.MarkFailed(ctx, release.ID, msg, nil); err != nil {
		w.Logger().Errorf("failed to mark release %s as failed: %v", release.ID, err)
	}
}
