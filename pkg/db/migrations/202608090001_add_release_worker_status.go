package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addReleaseWorkerStatus() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202608090001",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE stack_releases ADD COLUMN worker_status JSONB`).Error; err != nil {
				return fmt.Errorf("failed to add stack_releases.worker_status: %w", err)
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE stack_releases DROP COLUMN worker_status`).Error; err != nil {
				return fmt.Errorf("failed to drop stack_releases.worker_status: %w", err)
			}
			return nil
		},
	}
}
