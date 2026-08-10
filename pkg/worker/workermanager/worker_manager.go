package workermanager

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/observability"
	workerlib "github.com/Stackdome/stackdome/pkg/worker"
)

// Goroutines draining each worker's queue. The workqueue never hands the
// same key to two goroutines (in-flight keys re-queue as dirty), so this
// only requires operands to be comparable values — see models.StackOperand.
const maxConcurrentReconciles = 6

type WorkerManager interface {
	Start(ctx context.Context) error
	Find(operand interface{}) workerlib.Worker
	RegisterWorker(worker workerlib.Worker, operand interface{})
	Stop(drain bool)
	BackgroundJobEnqueuer
}

//go:generate mockgen -source=worker_manager.go -destination=../../mocks/mock_background_job_enqueuer.go -package=mocks

type BackgroundJobEnqueuer interface {
	Enqueue(operand interface{}) error
	EnqueueAfter(operand interface{}, after time.Duration) error
	EnqueueAfterCommit(ctx context.Context, operand interface{}) error
}

type WorkerManagerSpec struct {
	Environment string
	Metrics     *observability.Metrics
}

var _ WorkerManager = (*serviceWorkerManager)(nil)

type serviceWorkerManager struct {
	environment       string
	registeredWorkers map[reflect.Type]workerlib.Worker
	stopChan          chan struct{}
	log               logger.Logger
	metrics           *observability.Metrics
}

func NewWorkerManager(spec WorkerManagerSpec) *serviceWorkerManager {
	workerMgr := &serviceWorkerManager{
		environment:       spec.Environment,
		registeredWorkers: make(map[reflect.Type]workerlib.Worker),
		stopChan:          make(chan struct{}),
		log:               logger.NewLoggerWithPrefix(context.Background(), "worker-manager"),
		metrics:           spec.Metrics,
	}
	return workerMgr
}

// The operand prototype must be a comparable value type: the workqueue
// dedups and rate-limits by map-key equality, and a pointer's key is its
// address, so pointer operands silently break dedup (and non-comparable
// values panic inside the queue). Enqueue routes by the same type, so
// rejecting the prototype here covers every enqueue path.
func (s *serviceWorkerManager) RegisterWorker(worker workerlib.Worker, operand interface{}) {
	t := reflect.TypeOf(operand)
	if t.Kind() == reflect.Pointer || !t.Comparable() {
		panic(fmt.Sprintf("worker %q operand prototype must be a comparable value type, got %s", worker.Name(), t))
	}
	s.registeredWorkers[t] = worker
}
func (s *serviceWorkerManager) Start(ctx context.Context) error {
	if len(s.registeredWorkers) == 0 {
		return errors.New("no workers registered under the worker manager")
	}

	for _, worker := range s.registeredWorkers {
		if worker != nil {
			go func(w workerlib.Worker) {
				s.startWorker(ctx, w)
			}(worker)
			s.log.Info(ctx, "%s worker started", worker.Name())
		}
	}
	return nil
}

func (s *serviceWorkerManager) Stop(drain bool) {
	s.log.Infof("stopping worker manager (drain=%t)", drain)
	close(s.stopChan)
	var wg sync.WaitGroup
	wg.Add(len(s.registeredWorkers))

	// When the workqueue is shutdown, the workers exit their
	// respective go routines.
	for _, worker := range s.registeredWorkers {
		go func(w workerlib.Worker) {
			defer wg.Done()
			if drain {
				w.WorkQueue().ShutDownWithDrain()
			} else {
				w.WorkQueue().ShutDown()
			}
		}(worker)
	}
	wg.Wait()
	s.log.Infof("worker manager stopped")
}

func (s *serviceWorkerManager) Enqueue(operand interface{}) error {
	worker := s.Find(operand)
	if worker == nil {
		return fmt.Errorf("worker not registered for '%s' type", reflect.TypeOf(operand).String())
	}
	worker.EnqueueNow(operand)
	return nil
}

func (s *serviceWorkerManager) EnqueueAfter(operand interface{}, after time.Duration) error {
	worker := s.Find(operand)
	if worker == nil {
		return fmt.Errorf("worker not registered for '%s' type", reflect.TypeOf(operand).String())
	}
	worker.EnqueueAfter(operand, after)
	return nil
}

func (s *serviceWorkerManager) EnqueueAfterCommit(ctx context.Context, operand interface{}) error {
	worker := s.Find(operand)
	if worker == nil {
		return fmt.Errorf("worker not registered for '%s' type", reflect.TypeOf(operand).String())
	}
	db.OnPostCommit(ctx, func() {
		worker.EnqueueNow(operand)
	})
	return nil
}

func (s *serviceWorkerManager) Find(operand interface{}) workerlib.Worker {
	if worker, found := s.registeredWorkers[reflect.TypeOf(operand)]; found {
		return worker
	}
	return nil
}

func (s *serviceWorkerManager) startWorker(ctx context.Context, worker workerlib.Worker) {
	// If worker implements PeriodicReconcilable,
	// we run startPeriodicWorkEnqueue in a separate goroutine
	// to periodically enqueue work to the worker's workqueue.
	switch val := worker.(type) {
	case workerlib.PeriodicReconcilable:
		go s.startPeriodicWorkEnqueue(ctx, worker, val)
	}

	// Populate worker queue once before starting the worker loop.
	s.populateWorkerQueue(ctx, worker)

	var wg sync.WaitGroup
	for range maxConcurrentReconciles {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for s.processNextWorkItemForWorker(ctx, worker) {
			}
		}()
	}
	wg.Wait()
}

func (s *serviceWorkerManager) populateWorkerQueue(ctx context.Context, worker workerlib.Worker) {
	ctx = withSystemIdentity(ctx)
	worker.Logger().Info(ctx, "populating queue once for worker: %s", worker.Name())
	operandList, err := worker.GetInput(ctx)
	if err != nil {
		worker.Logger().Error(ctx, "failed to populate queue for worker: %v", err)
		return
	}
	for _, operand := range operandList {
		worker.EnqueueNow(operand)
	}
}

func (s *serviceWorkerManager) processNextWorkItemForWorker(ctx context.Context, worker workerlib.Worker) bool {
	operand, shutdown := worker.WorkQueue().Get()
	if shutdown {
		return false
	}
	result, err := s.workerExecute(ctx, worker, operand)
	s.metrics.ObserveWorkerExecution(worker.Name(), workerResultLabel(result, err))
	defer worker.WorkQueue().Done(operand)
	switch {
	case err != nil:
		requeues := worker.WorkQueue().NumRequeues(operand)
		worker.Logger().WithFields(map[string]interface{}{
			"operand":  fmt.Sprintf("%v", operand),
			"requeues": requeues,
		}).Error(ctx, "reconcile error: %v", err)
		worker.WorkQueue().AddRateLimited(operand)
	case result.RequeueAfter > 0:
		worker.WorkQueue().Forget(operand)
		worker.EnqueueAfter(operand, result.RequeueAfter)
	case result.Requeue:
		worker.WorkQueue().AddRateLimited(operand)
	default:
		worker.WorkQueue().Forget(operand)
	}
	return true
}

type workerPanicError struct {
	value interface{}
}

func (e workerPanicError) Error() string {
	return fmt.Sprintf("worker panicked: %v", e.value)
}

func workerResultLabel(result workerlib.Result, err error) string {
	var panicErr workerPanicError
	switch {
	case errors.As(err, &panicErr):
		return observability.WorkerResultPanic
	case err != nil:
		return observability.WorkerResultError
	case result.Requeue || result.RequeueAfter > 0:
		return observability.WorkerResultRequeue
	default:
		return observability.WorkerResultSuccess
	}
}

// startPeriodicReconciler periodically runs the worker's GetInput method
// and enqueues its result to the workers workqueue via the EnqueueNow
// method.
func (s *serviceWorkerManager) startPeriodicWorkEnqueue(ctx context.Context,
	worker workerlib.Worker, periodicWorkerConfig workerlib.PeriodicReconcilable) {
	worker.Logger().Info(ctx, "starting PeriodicReconciler for worker: %s", worker.Name())
	ticker := time.NewTicker(periodicWorkerConfig.Interval())
	defer ticker.Stop()
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			operandList, err := s.getWorkerInput(ctx, worker)
			if err != nil {
				worker.Logger().Error(ctx, "failed to perform periodic reconcile: %v", err)
				continue
			}
			for _, operand := range operandList {
				worker.EnqueueNow(operand)
			}
		}
	}
}
func withSystemIdentity(ctx context.Context) context.Context {
	return auth.SetIdentityInContext(ctx, &auth.Identity{IsSystem: true})
}

func (s *serviceWorkerManager) getWorkerInput(
	ctx context.Context, worker workerlib.Worker) (res []workerlib.Operand, err error) {
	ctx = withSystemIdentity(ctx)
	defer func() {
		if panicErr := recover(); panicErr != nil {
			{
				worker.Logger().Error(ctx, "worker %s panicked: %v", worker.Name(), panicErr)
				err = fmt.Errorf("worker %s panicked: %v", worker.Name(), panicErr)
			}
		}
	}()
	res, serr := worker.GetInput(ctx)
	if serr != nil {
		err = serr
		return
	}
	return
}

func (s *serviceWorkerManager) workerExecute(
	ctx context.Context, worker workerlib.Worker, operand workerlib.Operand) (res workerlib.Result, err error) {
	ctx = withSystemIdentity(ctx)
	defer func() {
		if panicErr := recover(); panicErr != nil {
			{
				worker.Logger().Error(ctx, "worker %s panicked: %v", worker.Name(), panicErr)
				err = workerPanicError{value: panicErr}
			}
		}
	}()
	res, serr := worker.Execute(ctx, operand)
	if serr != nil {
		err = serr
		return
	}
	return
}
