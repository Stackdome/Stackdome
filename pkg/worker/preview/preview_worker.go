package preview

import (
	"context"
	"sync"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/worker"
)

const (
	PreviewWorkerName    = "preview-worker"
	convergePollInterval = 15 * time.Second
)

type PreviewWorkerSpec struct {
	PreviewStackService previewStackService
	PreviewStackStore   previewStackStore
	ConfigStore         configStore
	ReleaseService      releaseService
	StackService        stackService
	Env                 string
}

type previewWorker struct {
	previewStackStore previewStackStore
	stackfileCache    sync.Map
	previewCacheKeys  sync.Map
	subReconcilers    []subReconciler
	worker.BaseWorker
}

var _ worker.Worker = (*previewWorker)(nil)

func NewPreviewWorker(spec PreviewWorkerSpec) worker.Worker {
	w := &previewWorker{
		previewStackStore: spec.PreviewStackStore,
		BaseWorker:        worker.NewBaseWorker(PreviewWorkerName, spec.Env),
	}
	w.subReconcilers = []subReconciler{
		newDeprovisionReconciler(spec, &w.stackfileCache, &w.previewCacheKeys),
		newProvisionReconciler(spec, &w.stackfileCache, &w.previewCacheKeys),
		newConvergeReconciler(spec),
	}
	return w
}

func (w *previewWorker) Execute(ctx context.Context, operand worker.Operand) (worker.Result, *errors.ServiceError) {
	previewRef, ok := operand.(*models.PreviewStack)
	if !ok {
		return worker.Result{}, w.WorkerError.NewError("invalid operand type, expected *models.PreviewStack")
	}

	preview, serr := w.previewStackStore.GetByID(ctx, previewRef.ID)
	if serr != nil {
		if serr.Is404() {
			w.Logger().Infof("preview %s not found, skipping", previewRef.ID)
			return worker.Result{}, nil
		}
		return worker.Result{}, serr
	}

	if isInTerminalState(preview.Status.Phase) {
		return worker.Result{}, nil
	}

	w.Logger().Infof("processing preview %s (phase=%s)", preview.ID, preview.Status.Phase)

	res, reconcileErr := w.reconcile(ctx, preview)
	if reconcileErr != nil {
		w.Logger().Errorf("failed to reconcile preview %s: %v", preview.ID, reconcileErr)
		return worker.Result{}, w.WorkerError.NewError("failed to reconcile preview %s: %v", preview.ID, reconcileErr)
	}

	return res, nil
}

func (w *previewWorker) reconcile(ctx context.Context, preview *models.PreviewStack) (worker.Result, error) {
	for _, sr := range w.subReconcilers {
		w.Logger().Infof("running sub-reconciler: %s for preview: %s", sr.Name(), preview.ID)
		result, err := sr.Reconcile(ctx, preview)
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

func (w *previewWorker) Interval() time.Duration {
	return 30 * time.Second
}

const activePreviewPageSize = 50

func (w *previewWorker) GetInput(ctx context.Context) ([]worker.Operand, *errors.ServiceError) {
	var operands []worker.Operand
	for page := 1; ; page++ {
		previews, sErr := w.previewStackStore.ListActive(ctx, page, activePreviewPageSize)
		if sErr != nil {
			return nil, w.WorkerError.NewError("failed to list active previews: %v", sErr)
		}
		for _, p := range previews {
			operands = append(operands, &models.PreviewStack{ID: p.ID})
		}
		if len(previews) < activePreviewPageSize {
			break
		}
	}
	return operands, nil
}

func isInTerminalState(phase models.PreviewStackPhase) bool {
	return phase == models.PreviewStackPhaseReady || phase == models.PreviewStackPhaseFailed
}
