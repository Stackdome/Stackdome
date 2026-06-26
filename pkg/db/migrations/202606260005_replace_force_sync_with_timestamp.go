package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func replaceForceSyncWithTimestamp() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202606260005",
		Migrate: func(tx *gorm.DB) error {
			sql := `
				ALTER TABLE preview_stacks
					DROP COLUMN IF EXISTS force_sync,
					ADD COLUMN force_sync_requested_at TIMESTAMPTZ;
			`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to replace force_sync with timestamp: %w", err)
			}
			return nil
		},
	}
}
