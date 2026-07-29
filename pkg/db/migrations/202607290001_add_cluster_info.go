package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addClusterInfo() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607290001",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE clusters ADD COLUMN cluster_info JSONB`).Error; err != nil {
				return fmt.Errorf("failed to add clusters.cluster_info: %w", err)
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE clusters DROP COLUMN cluster_info`).Error; err != nil {
				return fmt.Errorf("failed to drop clusters.cluster_info: %w", err)
			}
			return nil
		},
	}
}
