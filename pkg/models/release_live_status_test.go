package models

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("BuildReleaseLiveStatus", func() {
	var stack *Stack
	var release *StackRelease

	// cond builds a True stack condition of the given type.
	cond := func(t StackConditionType) Condition {
		return Condition{Type: string(t), Status: string(ConditionTrue)}
	}

	ginkgo.BeforeEach(func() {
		release = &StackRelease{
			ID:    "rel-1",
			State: ReleaseStateReleased,
			Snapshot: StackSnapshot{
				Resources: []*StackResource{{Name: "web"}, {Name: "worker"}},
			},
		}
		stack = &Stack{
			Status: &StackStatus{
				TargetRevision:     "rev-2",
				ObservedCrRevision: "rev-2",
				LastConverged:      &StackConvergenceRecord{ReleaseID: "rel-1", Revision: "rev-2"},
				// Healthy converged baseline: settled and serving.
				Conditions: []Condition{cond(StackConditionConverged), cond(StackConditionAvailable)},
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
		release = &StackRelease{ID: "rel-2", State: ReleaseStateInProgress, Snapshot: release.Snapshot}
		stack.Status.Conditions = []Condition{cond(StackConditionProgressing)}
		ls := BuildReleaseLiveStatus(release, stack)
		gomega.Expect(ls).NotTo(gomega.BeNil())
		gomega.Expect(ls.Health).To(gomega.Equal(ReleaseHealthProgressing))
	})

	ginkgo.It("returns nil for a terminal, non-live release", func() {
		release = &StackRelease{ID: "rel-0", State: ReleaseStateSuperseded, Snapshot: release.Snapshot}
		gomega.Expect(BuildReleaseLiveStatus(release, stack)).To(gomega.BeNil())
	})

	ginkgo.It("returns nil when the stack has no status yet", func() {
		stack.Status = nil
		release.State = ReleaseStateReleased
		gomega.Expect(BuildReleaseLiveStatus(release, stack)).To(gomega.BeNil())
	})

	// Health is a direct rollup of the stack's aggregate conditions (cluster-agent
	// v0.6.6+), mirroring the agent's phase priority Failed > Progressing > Degraded > Converged.

	ginkgo.It("rolls up ok from the Converged condition", func() {
		stack.Status.Conditions = []Condition{cond(StackConditionConverged), cond(StackConditionAvailable)}
		gomega.Expect(BuildReleaseLiveStatus(release, stack).Health).To(gomega.Equal(ReleaseHealthOK))
	})

	ginkgo.It("rolls up ok when serving (Available) but not yet converged", func() {
		// Serving traffic on the prior revision while the newest release has not
		// converged — Available without Converged must not read as unavailable.
		stack.Status.Conditions = []Condition{cond(StackConditionAvailable)}
		gomega.Expect(BuildReleaseLiveStatus(release, stack).Health).To(gomega.Equal(ReleaseHealthOK))
	})

	ginkgo.It("rolls up progressing from the Progressing condition", func() {
		stack.Status.Conditions = []Condition{cond(StackConditionProgressing)}
		gomega.Expect(BuildReleaseLiveStatus(release, stack).Health).To(gomega.Equal(ReleaseHealthProgressing))
	})

	ginkgo.It("rolls up degraded from the Degraded condition", func() {
		stack.Status.Conditions = []Condition{cond(StackConditionDegraded)}
		gomega.Expect(BuildReleaseLiveStatus(release, stack).Health).To(gomega.Equal(ReleaseHealthDegraded))
	})

	ginkgo.It("rolls up failed from the Stalled condition", func() {
		stack.Status.Conditions = []Condition{cond(StackConditionStalled)}
		gomega.Expect(BuildReleaseLiveStatus(release, stack).Health).To(gomega.Equal(ReleaseHealthFailed))
	})

	ginkgo.It("prefers failed over progressing when both Stalled and Progressing are set", func() {
		stack.Status.Conditions = []Condition{cond(StackConditionStalled), cond(StackConditionProgressing)}
		gomega.Expect(BuildReleaseLiveStatus(release, stack).Health).To(gomega.Equal(ReleaseHealthFailed))
	})

	ginkgo.It("prefers progressing over degraded when both are set", func() {
		stack.Status.Conditions = []Condition{cond(StackConditionProgressing), cond(StackConditionDegraded)}
		gomega.Expect(BuildReleaseLiveStatus(release, stack).Health).To(gomega.Equal(ReleaseHealthProgressing))
	})

	ginkgo.It("rolls up progressing for an active release with no rollout condition yet", func() {
		release.State = ReleaseStateInProgress
		stack.Status.Conditions = []Condition{}
		gomega.Expect(BuildReleaseLiveStatus(release, stack).Health).To(gomega.Equal(ReleaseHealthProgressing))
	})

	ginkgo.It("rolls up unavailable for a live release that is no longer serving", func() {
		// Live (non-active) release with no positive condition and not serving —
		// rolled out once, now down.
		release.State = ReleaseStateReleased
		stack.Status.Conditions = []Condition{}
		gomega.Expect(BuildReleaseLiveStatus(release, stack).Health).To(gomega.Equal(ReleaseHealthUnavailable))
	})

	ginkgo.It("scopes live_status.resources to only the release's snapshot members", func() {
		release.Snapshot = StackSnapshot{Resources: []*StackResource{{Name: "web"}}}
		stack.StackResources = append(stack.StackResources, &StackResource{
			Name:   "busybox",
			Status: &StackResourceStatus{State: StackResourcePhaseReady},
		})

		ls := BuildReleaseLiveStatus(release, stack)
		gomega.Expect(ls.Resources).To(gomega.HaveLen(1))
		gomega.Expect(ls.Resources).To(gomega.HaveKey("web"))
		gomega.Expect(ls.Resources).NotTo(gomega.HaveKey("busybox"))
	})
})
