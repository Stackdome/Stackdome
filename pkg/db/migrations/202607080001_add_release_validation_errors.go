package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addReleaseValidationErrors() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607080001_add_release_validation_errors",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE stack_releases ADD COLUMN IF NOT EXISTS validation_errors jsonb`).Error; err != nil {
				return fmt.Errorf("failed to add validation_errors to stack_releases: %w", err)
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE stack_releases DROP COLUMN IF EXISTS validation_errors`).Error; err != nil {
				return fmt.Errorf("failed to drop validation_errors from stack_releases: %w", err)
			}
			return nil
		},
	}
}
