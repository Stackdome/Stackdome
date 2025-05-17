package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addSourceVolumeNameToVolumeMounts() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202505161810",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE volume_mounts ADD COLUMN source_volume_name TEXT;`).Error; err != nil {
				return fmt.Errorf("error adding source_volume_name column to volume_mounts 202505161810: %w", err)
			}
			return nil
		},
	}
}
