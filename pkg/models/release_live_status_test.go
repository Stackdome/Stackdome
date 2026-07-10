package models

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("BuildReleaseLiveStatus", func() {
	var stack *Stack
	var release *StackRelease

	ginkgo.BeforeEach(func() {
		release = &StackRelease{ID: "rel-1", State: ReleaseStateReleased}
		stack = &Stack{
			Status: &StackStatus{
				TargetRevision:     "rev-2",
				ObservedCrRevision: "rev-2",
				LastConverged:      &StackConvergenceRecord{ReleaseID: "rel-1", Revision: "rev-2"},
				Conditions:         []Condition{},
			},
			StackResources: []*StackResource{
				{Name: "web", Status: &StackResourceStatus{State: StackResourcePhaseReady, Replicas: 2, AvailableReplicas: 2}},
				{Name: "worker", Status: &StackResourceStatus{State: StackResourcePhaseReady, Replicas: 1, AvailableReplicas: 1}},
			},
		}
	})

	ginkgo.It("overlays the converged (live) release with per-resource status", func() {
		ls := BuildReleaseLiveStatus(release, stack)
		gomega.Expect(ls).NotTo(gomega.BeNil())
		gomega.Expect(ls.Health).To(gomega.Equal(ReleaseHealthOK))
		gomega.Expect(ls.Resources).To(gomega.HaveKey("web"))
		gomega.Expect(ls.Resources["web"].AvailableReplicas).To(gomega.Equal(int32(2)))
		gomega.Expect(ls.TargetRevision).To(gomega.Equal("rev-2"))
	})

	ginkgo.It("overlays an active release even when another release is converged", func() {
		release = &StackRelease{ID: "rel-2", State: ReleaseStateInProgress}
		ls := BuildReleaseLiveStatus(release, stack)
		gomega.Expect(ls).NotTo(gomega.BeNil())
		gomega.Expect(ls.Health).To(gomega.Equal(ReleaseHealthProgressing))
	})

	ginkgo.It("returns nil for a terminal, non-live release", func() {
		release = &StackRelease{ID: "rel-0", State: ReleaseStateSuperseded}
		gomega.Expect(BuildReleaseLiveStatus(release, stack)).To(gomega.BeNil())
	})

	ginkgo.It("returns nil when the stack has no status yet", func() {
		stack.Status = nil
		release.State = ReleaseStateReleased
		gomega.Expect(BuildReleaseLiveStatus(release, stack)).To(gomega.BeNil())
	})

	ginkgo.It("rolls up failed when any resource failed", func() {
		stack.StackResources[0].Status.State = StackResourcePhaseFailed
		gomega.Expect(BuildReleaseLiveStatus(release, stack).Health).To(gomega.Equal(ReleaseHealthFailed))
	})

	ginkgo.It("rolls up progressing when any resource is pending", func() {
		stack.StackResources[1].Status.State = StackResourcePhasePending
		gomega.Expect(BuildReleaseLiveStatus(release, stack).Health).To(gomega.Equal(ReleaseHealthProgressing))
	})

	ginkgo.It("rolls up progressing when a resource has no status yet", func() {
		stack.StackResources[1].Status = nil
		gomega.Expect(BuildReleaseLiveStatus(release, stack).Health).To(gomega.Equal(ReleaseHealthProgressing))
	})

	ginkgo.It("rolls up degraded from the stack Degraded condition when no resource failed", func() {
		stack.Status.Conditions = []Condition{{Type: string(StackConditionDegraded), Status: "True"}}
		gomega.Expect(BuildReleaseLiveStatus(release, stack).Health).To(gomega.Equal(ReleaseHealthDegraded))
	})
})
