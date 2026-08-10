package worker

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BaseWorker resource serialization", func() {
	It("does not run two release operands for one resource concurrently", func() {
		worker := NewBaseWorker("test-worker", "test")
		firstEntered := make(chan struct{})
		releaseFirst := make(chan struct{})
		secondEntered := make(chan struct{})

		go func() {
			unlock := worker.LockResource("volume-1")
			defer unlock()
			close(firstEntered)
			<-releaseFirst
		}()
		<-firstEntered

		go func() {
			unlock := worker.LockResource("volume-1")
			defer unlock()
			close(secondEntered)
		}()

		Consistently(secondEntered, 50*time.Millisecond).ShouldNot(Receive())
		close(releaseFirst)
		Eventually(secondEntered).Should(BeClosed())
	})
})
