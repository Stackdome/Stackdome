package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/Stackdome/stackdome/pkg/logger"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/utils/keymutex"
)

const resourceLockShards = 64

// BaseWorker implements generic functionalities for a worker.
// Actual workers types are supposed to embed this type in them.
type BaseWorker struct {
	Queue         workqueue.TypedRateLimitingInterface[Operand]
	WorkerName    string
	WorkerError   WorkerError
	logger        logger.Logger
	Env           string
	resourceLocks keymutex.KeyMutex
}

func NewBaseWorker(workerName string, env string) BaseWorker {
	return BaseWorker{
		WorkerName:  workerName,
		logger:      logger.NewLoggerWithPrefix(context.Background(), workerName),
		WorkerError: NewWorkerError(workerName),
		Queue: workqueue.NewTypedRateLimitingQueueWithConfig[Operand](
			workqueue.DefaultTypedControllerRateLimiter[Operand](),
			workqueue.TypedRateLimitingQueueConfig[Operand]{
				Name: workerName,
			},
		),
		Env:           env,
		resourceLocks: keymutex.NewHashed(resourceLockShards),
	}
}

// LockResource serializes operands that target the same persisted resource.
func (w *BaseWorker) LockResource(id string) func() {
	w.resourceLocks.LockKey(id)
	return func() {
		if err := w.resourceLocks.UnlockKey(id); err != nil {
			panic(fmt.Sprintf("unlock worker resource %q: %v", id, err))
		}
	}
}

// Clients call this method to enqueue work to a worker immeadiately.
func (w *BaseWorker) EnqueueNow(operand Operand) {
	w.Queue.Add(operand)
}

// Clients call this method to enqueue a worker after the passed duration.
func (w *BaseWorker) EnqueueAfter(operand Operand, after time.Duration) {
	w.Queue.AddAfter(operand, after)
}

func (w *BaseWorker) Name() string {
	return w.WorkerName
}

func (w *BaseWorker) Logger() logger.Logger {
	return w.logger
}

func (w *BaseWorker) WorkQueue() workqueue.TypedRateLimitingInterface[Operand] {
	return w.Queue
}
