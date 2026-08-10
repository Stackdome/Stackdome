package worker

import (
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ClusterMutationCoordinator", func() {
	DescribeTable("serializes mutation families that share a Kubernetes namespace", func(firstName, secondName string) {
		coordinator := NewClusterMutationCoordinator()
		firstWorker := NewBaseWorkerWithClusterMutationCoordinator(firstName, "test", coordinator)
		secondWorker := NewBaseWorkerWithClusterMutationCoordinator(secondName, "test", coordinator)

		firstUnlock := firstWorker.LockClusterNamespace("cluster-1", "namespace-1")

		var entered atomic.Bool
		done := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			unlock := secondWorker.LockClusterNamespace("cluster-1", "namespace-1")
			defer unlock()
			entered.Store(true)
			close(done)
		}()

		Expect(firstWorker.Name()).NotTo(Equal(secondWorker.Name()))
		Consistently(entered.Load, 100*time.Millisecond).Should(BeFalse())
		firstUnlock()
		Eventually(done).Should(BeClosed())
	},
		Entry("release apply and stack deletion", "release-worker", "stack-worker"),
		Entry("volume update and stack deletion", "volume-worker", "stack-worker"),
		Entry("Postgres deletion and active reconcile", "postgres-delete-worker", "postgres-active-worker"),
		Entry("Postgres reconcile and stack deletion", "postgres-addon-worker", "stack-worker"),
	)

	It("does not serialize different Kubernetes namespaces", func() {
		coordinator := NewClusterMutationCoordinator()
		unlock := coordinator.LockClusterNamespace("cluster-1", "namespace-1")
		defer unlock()

		done := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			otherUnlock := coordinator.LockClusterNamespace("cluster-1", "namespace-2")
			defer otherUnlock()
			close(done)
		}()

		Eventually(done).Should(BeClosed())
	})
})
