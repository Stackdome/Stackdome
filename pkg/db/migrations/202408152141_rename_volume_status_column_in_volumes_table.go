package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func renameVolumeStatusColumnInVolumesTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202408152141",
		Migrate: func(tx *gorm.DB) error {
			return tx.Exec(`
				ALTER TABLE volumes
				RENAME COLUMN volume_status TO status;
			`).Error
		},
	}
}
