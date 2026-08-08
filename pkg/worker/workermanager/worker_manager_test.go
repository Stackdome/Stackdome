package workermanager

import (
	stderrors "errors"
	"time"

	"github.com/Stackdome/stackdome/pkg/observability"
	workerlib "github.com/Stackdome/stackdome/pkg/worker"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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
