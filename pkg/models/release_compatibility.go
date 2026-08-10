package models

import "fmt"

// ValidatePinnedVolumeGitRevisions checks the immutable input required by
// cloud release reconciliation. Older releases created before commit pinning
// was introduced fail this check instead of resolving mutable Git state.
func ValidatePinnedVolumeGitRevisions(snapshot StackSnapshot) error {
	for _, volume := range snapshot.Volumes {
		if volume == nil || volume.VolumeSource == nil || volume.VolumeSource.GitRepoSource == nil {
			continue
		}
		if volume.VolumeSource.GitRepoSource.Revision.Commit == "" {
			return fmt.Errorf("volume %q has no resolved Git commit", volume.Name)
		}
	}
	return nil
}
