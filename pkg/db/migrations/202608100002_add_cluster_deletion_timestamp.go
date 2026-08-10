package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addClusterDeletionTimestamp() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202608100002",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE clusters ADD COLUMN deletion_timestamp timestamptz`).Error; err != nil {
				return fmt.Errorf("add clusters.deletion_timestamp: %w", err)
			}
			return nil
		},
	}
}
