package releasegc

import (
	"context"
	"sort"
	"time"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/worker"
)

const (
	WorkerName     = "release-gc-worker"
	gcPollInterval = 1 * time.Hour
)

type ReleaseGCRequest struct {
	StackID string
}

type releaseStore interface {
	ListTerminalSummariesByStackID(ctx context.Context, stackID string) ([]models.ReleaseSummary, *errors.ServiceError)
	DeleteByIDs(ctx context.Context, ids []string) *errors.ServiceError
	ListStackIDsExceedingRetention(ctx context.Context, cap int) ([]string, *errors.ServiceError)
}

type stackStore interface {
	GetByID(ctx context.Context, id string) (*models.Stack, *errors.ServiceError)
}

type ReleaseGCWorkerSpec struct {
	ReleaseStore releaseStore
	StackStore   stackStore
	Env          string
}

type releaseGCWorker struct {
	releaseStore releaseStore
	stackStore   stackStore
	worker.BaseWorker
}

var _ worker.Worker = (*releaseGCWorker)(nil)

func NewReleaseGCWorker(spec ReleaseGCWorkerSpec) worker.Worker {
	return &releaseGCWorker{
		releaseStore: spec.ReleaseStore,
		stackStore:   spec.StackStore,
		BaseWorker:   worker.NewBaseWorker(WorkerName, spec.Env),
	}
}

func (w *releaseGCWorker) Interval() time.Duration {
	return gcPollInterval
}

func (w *releaseGCWorker) Execute(ctx context.Context, operand worker.Operand) (worker.Result, *errors.ServiceError) {
	req, ok := operand.(*ReleaseGCRequest)
	if !ok {
		return worker.Result{}, w.WorkerError.NewError("invalid operand type, expected *ReleaseGCRequest")
	}

	stack, stackErr := w.stackStore.GetByID(ctx, req.StackID)
	if stackErr != nil {
		if stackErr.Is404() {
			return worker.Result{}, nil
		}
		return worker.Result{}, w.WorkerError.NewError("failed to get stack %s: %v", req.StackID, stackErr)
	}
	settings := stack.EffectiveSettings()

	candidates, err := w.releaseStore.ListTerminalSummariesByStackID(ctx, req.StackID)
	if err != nil {
		return worker.Result{}, w.WorkerError.NewError("failed to list releases for stack %s: %v", req.StackID, err)
	}

	toDelete := releaseIDsToGC(candidates, settings.ReleaseRetentionLimit, settings.MinSuccessfulReleases)
	if len(toDelete) == 0 {
		return worker.Result{}, nil
	}

	if err := w.releaseStore.DeleteByIDs(ctx, toDelete); err != nil {
		return worker.Result{}, w.WorkerError.NewError("failed to delete releases for stack %s: %v", req.StackID, err)
	}

	w.Logger().Infof("GC completed for stack %s: deleted %d of %d terminal releases", req.StackID, len(toDelete), len(candidates))
	return worker.Result{}, nil
}

func (w *releaseGCWorker) GetInput(ctx context.Context) ([]worker.Operand, *errors.ServiceError) {
	stackIDs, err := w.releaseStore.ListStackIDsExceedingRetention(ctx, models.DefaultReleaseRetentionLimit)
	if err != nil {
		return nil, w.WorkerError.NewError("failed to poll for GC candidates: %v", err)
	}
	operands := make([]worker.Operand, 0, len(stackIDs))
	for _, id := range stackIDs {
		operands = append(operands, &ReleaseGCRequest{StackID: id})
	}
	return operands, nil
}

// releaseIDsToGC returns the release IDs to delete, retaining the
// minSuccessful most-recent Released releases unconditionally,
// then filling up to retentionLimit total with the most-recent releases of
// any state.
func releaseIDsToGC(candidates []models.ReleaseSummary, retentionLimit, minSuccessful int) []string {
	bySeqDesc := make([]models.ReleaseSummary, len(candidates))
	copy(bySeqDesc, candidates)
	sort.Slice(bySeqDesc, func(i, j int) bool { return bySeqDesc[i].Sequence > bySeqDesc[j].Sequence })

	keep := make(map[string]struct{})
	releasedSeen := 0
	for _, s := range bySeqDesc {
		if s.State == models.ReleaseStateReleased && releasedSeen < minSuccessful {
			keep[s.ID] = struct{}{}
			releasedSeen++
		}
	}

	for _, s := range bySeqDesc {
		if len(keep) >= retentionLimit {
			break
		}
		keep[s.ID] = struct{}{}
	}

	var del []string
	for _, s := range candidates {
		if _, ok := keep[s.ID]; !ok {
			del = append(del, s.ID)
		}
	}
	return del
}
