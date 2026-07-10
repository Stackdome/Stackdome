package models

// ReleaseHealth is the computed runtime rollup for a live/active release.
type ReleaseHealth string

const (
	ReleaseHealthOK          ReleaseHealth = "ok"
	ReleaseHealthProgressing ReleaseHealth = "progressing"
	ReleaseHealthDegraded    ReleaseHealth = "degraded"
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
	Current *StackRelease
	Latest  *StackRelease
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

	resources := make(map[string]*StackResourceStatus, len(stack.StackResources))
	for _, r := range stack.StackResources {
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

func rollupHealth(release *StackRelease, stack *Stack) ReleaseHealth {
	health := ReleaseHealthOK
	for _, r := range stack.StackResources {
		switch {
		case r.Status == nil || r.Status.State == StackResourcePhasePending:
			if health == ReleaseHealthOK {
				health = ReleaseHealthProgressing
			}
		case r.Status.State == StackResourcePhaseFailed, r.Status.State == StackResourcePhaseUnknown:
			return ReleaseHealthFailed
		}
	}
	if health == ReleaseHealthOK {
		if IsConditionTrue(stack.Status.Conditions, string(StackConditionDegraded)) {
			return ReleaseHealthDegraded
		}
		if release.State.Active() {
			return ReleaseHealthProgressing
		}
	}
	return health
}
