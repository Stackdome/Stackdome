package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/ashishmax31/soradev-api-server/pkg/errors"
	"github.com/ashishmax31/soradev-api-server/pkg/logger"
	"k8s.io/client-go/util/workqueue"
)

type Operand interface{}

// Result contains the result of a worker's Execute invocation.
type Result struct {
	Requeue bool

	// RequeueAfter if greater than 0, tells the worker manager to requeue the operand after the Duration.
	// Implies that Requeue is true, there is no need to set Requeue to true at the same time as RequeueAfter.
	RequeueAfter time.Duration
}

// IsZero returns true if this result is empty.
func (r *Result) IsZero() bool {
	if r == nil {
		return true
	}
	return *r == Result{}
}

// Workers both run periodically at `Interval()` and
// when clients call `EnqueueNow()`.
type Worker interface {
	Execute(ctx context.Context, operand Operand) (Result, *errors.ServiceError)
	EnqueueNow(operand Operand)
	EnqueueAfter(Operand Operand, after time.Duration)
	WorkQueue() workqueue.RateLimitingInterface
	// Function that returns the list of operands to be passed to the workers execute method.
	GetInput(context.Context) ([]Operand, *errors.ServiceError)
	Logger() logger.Logger
	Name() string
}

// If a worker implements PeriodicReconcilable, the worker manager would run
// enqueue GetInput()'s result to the worker's work queue every Interval().
type PeriodicReconcilable interface {
	Interval() time.Duration
}

type WorkerError interface {
	NewError(format string, args ...any) *errors.ServiceError
}

type workerError struct {
	workerName string
}

func NewWorkerError(workerName string) *workerError {
	return &workerError{
		workerName: workerName,
	}
}

func (e *workerError) NewError(format string, args ...any) *errors.ServiceError {
	errString := fmt.Sprintf(format, args...)
	return errors.GeneralError("worker: '%s' error: %s", e.workerName, errString)
}

func ToOperandList[T any](in ...T) []Operand {
	res := make([]Operand, len(in))
	for i := range in {
		res[i] = Operand(in[i])
	}
	return res
}
