package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addVolumeSourceToVolumesTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202408132333",
		Migrate: func(tx *gorm.DB) error {
			// Add version and workspace column to workspace_storage table
			if err := tx.Exec(`
				ALTER TABLE volumes
				ADD COLUMN IF NOT EXISTS volume_source jsonb,
				DROP COLUMN IF EXISTS local_source,
				DROP COLUMN IF EXISTS build_source;
			`).Error; err != nil {
				return fmt.Errorf("error running volumes table migration 202408132333: %w", err)
			}
			return nil
		},
	}
}
