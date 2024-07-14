package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addManagerRunningColumnToCluster() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202407142317",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec("ALTER TABLE clusters ADD COLUMN IF NOT EXISTS manager_running BOOLEAN DEFAULT false").Error; err != nil {
				return fmt.Errorf("error running cluster migration 202407142317: %w", err)
			}
			return nil
		},
	}
}
