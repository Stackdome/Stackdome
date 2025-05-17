package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addVolumeIDFkConstraintToVolumeMounts() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202505171910",
		Migrate: func(tx *gorm.DB) error {
			// Add foreign key constraint for VolumeMount
			if err := tx.Exec(
				"ALTER TABLE volume_mounts ADD CONSTRAINT fk_volume_mounts_volume_id FOREIGN KEY (source_volume_id) REFERENCES volumes(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint on volumes to volume_mounts table: %w", err)
			}
			return nil
		},
	}
}
