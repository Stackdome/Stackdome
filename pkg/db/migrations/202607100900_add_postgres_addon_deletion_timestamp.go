package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addPostgresAddonDeletionTimestamp() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607100900",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec("ALTER TABLE postgres_addons ADD COLUMN deletion_timestamp timestamptz DEFAULT NULL").Error; err != nil {
				return fmt.Errorf("add deletion_timestamp column: %w", err)
			}
			// Backfill: addons already marked Deleting via the status state
			// must keep their deletion intent under the new column.
			if err := tx.Exec("UPDATE postgres_addons SET deletion_timestamp = NOW() WHERE status->>'state' = 'Deleting'").Error; err != nil {
				return fmt.Errorf("backfill deletion_timestamp for deleting addons: %w", err)
			}
			return nil
		},
	}
}
