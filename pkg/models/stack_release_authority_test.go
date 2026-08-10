package models

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("StackRelease workload authority", func() {
	var stack *Stack
	var release *StackRelease

	ginkgo.BeforeEach(func() {
		stack = &Stack{
			ID: "stack-1", OrganisationID: "org-1", ProjectID: "project-1", ClusterID: "cluster-1", NamespaceID: "namespace-1", Namespace: "demo",
			Status: &StackStatus{LastConverged: &StackConvergenceRecord{ReleaseID: "release-a"}},
		}
		release = &StackRelease{
			ID: "release-a", StackID: "stack-1", State: ReleaseStateReleased,
			Snapshot: StackSnapshot{Stack: StackShellSnapshot{ID: "stack-1", OrganisationID: "org-1", ProjectID: "project-1", ClusterID: "cluster-1", NamespaceID: "namespace-1", Namespace: "demo"}},
		}
	})

	ginkgo.It("accepts only the exact current converged release after completion", func() {
		gomega.Expect(release.IsAuthoritativeWorkloadRelease(stack, nil)).To(gomega.BeTrue())
		stack.Status.LastConverged.ReleaseID = "release-b"
		gomega.Expect(release.IsAuthoritativeWorkloadRelease(stack, nil)).To(gomega.BeFalse())
	})

	ginkgo.It("keeps the exact persisted converged release live after supersession", func() {
		release.State = ReleaseStateSuperseded

		gomega.Expect(release.IsAuthoritativeWorkloadRelease(stack, nil)).To(gomega.BeTrue())
	})

	ginkgo.It("gives a newer active release authority over the persisted release", func() {
		release.State = ReleaseStateSuperseded
		active := &StackRelease{ID: "release-b", StackID: stack.ID, State: ReleaseStateInProgress}

		gomega.Expect(release.IsAuthoritativeWorkloadRelease(stack, active)).To(gomega.BeFalse())
	})

	ginkgo.It("accepts a pending rollback only while it is the active release", func() {
		release.ID = "rollback-1"
		release.State = ReleaseStatePending
		active := *release
		gomega.Expect(release.IsAuthoritativeWorkloadRelease(stack, &active)).To(gomega.BeTrue())
		active.ID = "newer-release"
		gomega.Expect(release.IsAuthoritativeWorkloadRelease(stack, &active)).To(gomega.BeFalse())
	})

	ginkgo.It("fails closed for cancellation, failure, and identity drift", func() {
		for _, state := range []StackReleaseState{ReleaseStateCancelled, ReleaseStateFailed} {
			release.State = state
			gomega.Expect(release.IsAuthoritativeWorkloadRelease(stack, release)).To(gomega.BeFalse())
		}
		release.State = ReleaseStateReleased
		release.Snapshot.Stack.Namespace = "other"
		gomega.Expect(release.IsAuthoritativeWorkloadRelease(stack, nil)).To(gomega.BeFalse())
	})
})
