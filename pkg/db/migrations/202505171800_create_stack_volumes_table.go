package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createStackVolumesTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202505171800",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`CREATE TABLE IF NOT EXISTS stack_volumes (
				stack_id TEXT NOT NULL,
				volume_id TEXT NOT NULL UNIQUE,
				PRIMARY KEY (stack_id, volume_id)
			);`).Error; err != nil {
				return fmt.Errorf("error creating stack_volumes table 202505171800: %w", err)
			}

			if err := tx.Exec(`ALTER TABLE stack_volumes
				ADD FOREIGN KEY (volume_id) REFERENCES volumes(id) ON DELETE CASCADE;`).Error; err != nil {
				return fmt.Errorf("error adding foreign key constraints to stack_volumes table 202505171800: %w", err)
			}
			return nil
		},
	}
}
