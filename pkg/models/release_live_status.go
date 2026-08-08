package models

// ReleaseHealth is the computed runtime rollup for a live/active release.
type ReleaseHealth string

const (
	ReleaseHealthOK          ReleaseHealth = "ok"
	ReleaseHealthProgressing ReleaseHealth = "progressing"
	ReleaseHealthDegraded    ReleaseHealth = "degraded"
	ReleaseHealthUnavailable ReleaseHealth = "unavailable"
	ReleaseHealthFailed      ReleaseHealth = "failed"
)

// ReleaseLiveStatus is a read-time overlay: current stack/resource status
// attributed to a release. Never persisted.
type ReleaseLiveStatus struct {
	Health           ReleaseHealth
	Resources        map[string]*StackResourceStatus
	Conditions       []Condition
	TargetRevision   string
	ObservedRevision string
}

// StackReleaseRefs pairs the currently converged (live) release with the
// highest-sequence one for stack embedding.
type StackReleaseRefs struct {
	Converged *StackRelease
	Latest    *StackRelease
	// DeployHistory is DeployHistoryDays of deploy counts, oldest first, with
	// zero-filled gaps. Nil for a stack that has never deployed — which is the
	// distinction the card needs: no bars at all reads "No deploys yet", where
	// fourteen zeroes would draw a flat line and claim the stack went quiet.
	DeployHistory []int
}

// BuildReleaseLiveStatus overlays current stack/resource status onto a release.
// Returns nil unless the release is active (Pending/InProgress) or live
// (the stack's last converged release).
func BuildReleaseLiveStatus(release *StackRelease, stack *Stack) *ReleaseLiveStatus {
	if stack.Status == nil {
		return nil
	}
	live := stack.Status.LastConverged != nil && stack.Status.LastConverged.ReleaseID == release.ID
	if !release.State.Active() && !live {
		return nil
	}

	members := make(map[string]struct{}, len(release.Snapshot.Resources))
	for _, r := range release.Snapshot.Resources {
		members[r.Name] = struct{}{}
	}
	liveResources := make([]*StackResource, 0, len(stack.StackResources))
	for _, r := range stack.StackResources {
		if _, ok := members[r.Name]; ok {
			liveResources = append(liveResources, r)
		}
	}

	resources := make(map[string]*StackResourceStatus, len(liveResources))
	for _, r := range liveResources {
		resources[r.Name] = r.Status
	}

	return &ReleaseLiveStatus{
		Health:           rollupHealth(release, stack),
		Resources:        resources,
		Conditions:       stack.Status.Conditions,
		TargetRevision:   stack.Status.TargetRevision,
		ObservedRevision: stack.Status.ObservedCrRevision,
	}
}

// rollupHealth maps the stack's aggregate conditions (cluster-agent v0.6.6+) to a
// single release health, mirroring the agent's phase priority:
// Failed(Stalled) > Progressing > Degraded > Converged. Available is checked below
// Converged to keep a serving-but-not-yet-converged stack (e.g. still on the prior
// revision) out of the unavailable arm. The fall-through splits on whether the
// release is still deploying (progressing) or a live release that fell out of
// serving (unavailable).
func rollupHealth(release *StackRelease, stack *Stack) ReleaseHealth {
	conds := stack.Status.Conditions
	switch {
	case IsConditionTrue(conds, string(StackConditionStalled)):
		return ReleaseHealthFailed
	case IsConditionTrue(conds, string(StackConditionProgressing)):
		return ReleaseHealthProgressing
	case IsConditionTrue(conds, string(StackConditionDegraded)):
		return ReleaseHealthDegraded
	case IsConditionTrue(conds, string(StackConditionConverged)):
		return ReleaseHealthOK
	case IsConditionTrue(conds, string(StackConditionAvailable)):
		return ReleaseHealthOK // serving traffic (e.g. prior revision) though not yet converged
	case release.State.Active():
		return ReleaseHealthProgressing // active deploy, no rollout condition yet (pods not created)
	default:
		return ReleaseHealthUnavailable // live release, no positive condition, not serving
	}
}
