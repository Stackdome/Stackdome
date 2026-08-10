package workermanager

import (
	"context"
	stderrors "errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/observability"
	workerlib "github.com/Stackdome/stackdome/pkg/worker"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/util/workqueue"
)

type testOperand struct{ ID string }

// testWorker wraps a real BaseWorker (real workqueue) around a configurable
// Execute so queue semantics — dedup, per-key serialization, concurrency —
// are exercised for real instead of mocked away.
type testWorker struct {
	execute func(ctx context.Context, operand workerlib.Operand) (workerlib.Result, *errors.ServiceError)
	workerlib.BaseWorker
}

func (w *testWorker) Execute(ctx context.Context, operand workerlib.Operand) (workerlib.Result, *errors.ServiceError) {
	return w.execute(ctx, operand)
}

func (w *testWorker) GetInput(ctx context.Context) ([]workerlib.Operand, *errors.ServiceError) {
	return nil, nil
}

func newTestWorker(execute func(ctx context.Context, operand workerlib.Operand) (workerlib.Result, *errors.ServiceError)) *testWorker {
	return &testWorker{
		execute:    execute,
		BaseWorker: workerlib.NewBaseWorker("test-worker", "test"),
	}
}

var _ = Describe("serviceWorkerManager", func() {
	var mgr *serviceWorkerManager

	startManager := func(w workerlib.Worker) {
		mgr = NewWorkerManager(WorkerManagerSpec{Environment: "test", Metrics: observability.NewMetrics()})
		mgr.RegisterWorker(w, testOperand{})
		Expect(mgr.Start(context.Background())).To(Succeed())
	}

	It("rejects pointer and non-comparable operand prototypes at registration", func() {
		w := newTestWorker(nil)
		mgr := NewWorkerManager(WorkerManagerSpec{Environment: "test", Metrics: observability.NewMetrics()})

		Expect(func() { mgr.RegisterWorker(w, &testOperand{}) }).To(PanicWith(ContainSubstring("comparable value type")))
		Expect(func() { mgr.RegisterWorker(w, struct{ IDs []string }{}) }).To(PanicWith(ContainSubstring("comparable value type")))
		Expect(func() { mgr.RegisterWorker(w, testOperand{}) }).NotTo(Panic())
	})

	It("dedups value operands enqueued while one is already queued or in flight", func() {
		gate := make(chan struct{})
		var calls atomic.Int32
		w := newTestWorker(func(ctx context.Context, operand workerlib.Operand) (workerlib.Result, *errors.ServiceError) {
			calls.Add(1)
			<-gate
			return workerlib.Result{}, nil
		})
		startManager(w)
		defer mgr.Stop(false)

		Expect(mgr.Enqueue(testOperand{ID: "a"})).To(Succeed())
		Eventually(calls.Load).Should(BeEquivalentTo(1))

		// Re-enqueued 3x while in flight: collapses to one dirty entry,
		// so exactly one more Execute after the gate opens.
		for range 3 {
			Expect(mgr.Enqueue(testOperand{ID: "a"})).To(Succeed())
		}
		close(gate)

		Eventually(calls.Load).Should(BeEquivalentTo(2))
		Consistently(calls.Load, 200*time.Millisecond).Should(BeEquivalentTo(2))
	})

	It("processes distinct operands concurrently but never the same operand twice at once", func() {
		var (
			mu       sync.Mutex
			inFlight = map[workerlib.Operand]int{}
			peak     int
			overlap  bool
			done     atomic.Int32
		)
		w := newTestWorker(func(ctx context.Context, operand workerlib.Operand) (workerlib.Result, *errors.ServiceError) {
			mu.Lock()
			inFlight[operand]++
			if inFlight[operand] > 1 {
				overlap = true
			}
			total := 0
			for _, n := range inFlight {
				total += n
			}
			if total > peak {
				peak = total
			}
			mu.Unlock()

			time.Sleep(50 * time.Millisecond)

			mu.Lock()
			inFlight[operand]--
			mu.Unlock()
			done.Add(1)
			return workerlib.Result{}, nil
		})
		startManager(w)
		defer mgr.Stop(false)

		for _, id := range []string{"a", "b", "c", "d", "e", "f"} {
			Expect(mgr.Enqueue(testOperand{ID: id})).To(Succeed())
			// Same key again while (likely) in flight: must not run in parallel
			// with itself.
			Expect(mgr.Enqueue(testOperand{ID: id})).To(Succeed())
		}

		Eventually(done.Load, 5*time.Second).Should(BeNumerically(">=", 6))

		mu.Lock()
		defer mu.Unlock()
		Expect(overlap).To(BeFalse(), "same operand ran concurrently with itself")
		Expect(peak).To(BeNumerically(">", 1), "expected concurrent processing of distinct operands")
	})

	It("keeps transient non-periodic work retryable after twenty errors", func() {
		w := newTestWorker(func(context.Context, workerlib.Operand) (workerlib.Result, *errors.ServiceError) {
			return workerlib.Result{}, errors.GeneralError("transient failure")
		})
		w.Queue = workqueue.NewTypedRateLimitingQueue[workerlib.Operand](
			workqueue.NewTypedItemFastSlowRateLimiter[workerlib.Operand](0, 0, 1),
		)
		mgr = NewWorkerManager(WorkerManagerSpec{Environment: "test", Metrics: observability.NewMetrics()})
		operand := testOperand{ID: "eventual-work"}
		w.Queue.Add(operand)

		for range 22 {
			Expect(mgr.processNextWorkItemForWorker(context.Background(), w)).To(BeTrue())
			Eventually(w.Queue.Len).Should(Equal(1))
		}
	})
})

var _ = Describe("worker execution result metrics", func() {
	DescribeTable("maps worker results to bounded labels",
		func(result workerlib.Result, err error, expected string) {
			Expect(workerResultLabel(result, err)).To(Equal(expected))
		},
		Entry("success", workerlib.Result{}, nil, observability.WorkerResultSuccess),
		Entry("error", workerlib.Result{}, stderrors.New("failed"), observability.WorkerResultError),
		Entry("immediate requeue", workerlib.Result{Requeue: true}, nil, observability.WorkerResultRequeue),
		Entry("delayed requeue", workerlib.Result{RequeueAfter: time.Second}, nil, observability.WorkerResultRequeue),
		Entry("panic", workerlib.Result{}, workerPanicError{value: "boom"}, observability.WorkerResultPanic),
	)
})
