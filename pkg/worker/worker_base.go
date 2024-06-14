package worker

import (
	"context"
	"time"

	"github.com/ashishmax31/soradev-api-server/pkg/logger"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BaseWorker implements generic functionalities for a worker.
// Actual workers types are supposed to embed this type in them.
type BaseWorker struct {
	Queue         workqueue.RateLimitingInterface
	WorkerName    string
	WorkerError   WorkerError
	logger        logger.Logger
	Env           string
	ClusterClient client.Client
}

func NewBaseWorker(workerName string, env string, clusterClient client.Client) BaseWorker {
	return BaseWorker{
		WorkerName:  workerName,
		logger:      logger.NewLoggerWithPrefix(context.Background(), workerName),
		WorkerError: NewWorkerError(workerName),
		Queue: workqueue.NewNamedRateLimitingQueue(
			workqueue.DefaultControllerRateLimiter(), workerName),
		Env:           env,
		ClusterClient: clusterClient,
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

func (w *BaseWorker) WorkQueue() workqueue.RateLimitingInterface {
	return w.Queue
}
