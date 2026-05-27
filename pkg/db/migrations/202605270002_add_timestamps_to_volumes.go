package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addTimestampsToVolumesTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202605270002",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE volumes ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`).Error; err != nil {
				return fmt.Errorf("error adding created_at column to volumes table: %w", err)
			}
			if err := tx.Exec(`ALTER TABLE volumes ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`).Error; err != nil {
				return fmt.Errorf("error adding updated_at column to volumes table: %w", err)
			}
			return nil
		},
	}
}
