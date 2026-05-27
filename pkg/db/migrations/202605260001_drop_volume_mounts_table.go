package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func dropVolumeMountsTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202605260001_drop_volume_mounts_table",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec("DROP TABLE IF EXISTS volume_mounts").Error; err != nil {
				return fmt.Errorf("failed to drop volume_mounts table: %w", err)
			}
			return nil
		},
	}
}
