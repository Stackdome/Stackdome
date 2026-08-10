package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func renameClusterPlatformToSharedCompute() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202608100001",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE clusters RENAME COLUMN platform TO shared_compute`).Error; err != nil {
				return fmt.Errorf("rename clusters.platform to shared_compute: %w", err)
			}
			return nil
		},
	}
}
