package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addSourceVolumeTypeToVolumeMounts() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202412091922",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE volume_mounts ADD COLUMN source_volume_type TEXT;`).Error; err != nil {
				return fmt.Errorf("error adding source_volume_type column to volume_mounts 202412091922: %w", err)
			}
			return nil
		},
	}
}
