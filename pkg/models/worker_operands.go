package models

// Workqueue operands. Must be comparable values, never pointers:
//   - the workqueue dedups and rate-limits by map-key equality,
//   - a pointer's map key is its address, so every allocation looks like
//     new work and duplicates accumulate for long-running operands,
//   - value keys also guarantee no two goroutines process the same
//     operand concurrently once workers run with multiple goroutines.
type StackOperand struct {
	ID        string
	ReleaseID string
}

type StackReleaseOperand struct{ ID string }

type VolumeOperand struct{ ID string }

type PostgresAddonOperand struct{ ID string }

type OrgInviteOperand struct{ ID string }

type PreviewStackOperand struct{ ID string }

// ClusterImageRegistryOperand serializes all registry reconciliation for one
// cluster, including cluster deletion once its registries have been removed.
type ClusterImageRegistryOperand struct{ ClusterID string }

type WorkloadReleaseRef struct {
	StackID   string
	ReleaseID string
}

type WorkloadAuthorityScan struct {
	Releases         []WorkloadReleaseRef
	DeletingStackIDs []string
}
