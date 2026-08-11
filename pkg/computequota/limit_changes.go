package computequota

import "github.com/Stackdome/stackdome/pkg/models"

// VolumeLimitChange describes a proposed volume creation.
type VolumeLimitChange struct {
	OrganisationID string
	Size           string
}

// PostgresAddonLimitChange describes a proposed PostgreSQL addon create or update.
type PostgresAddonLimitChange struct {
	OrganisationID string
	CreatesAddon   bool
	Addon          *models.PostgresAddon
}
