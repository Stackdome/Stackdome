package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addStackResourceNameToImageBuilds() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "2023121141555",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE image_builds ADD COLUMN stack_resource_name TEXT;`).Error; err != nil {
				return fmt.Errorf("err addding stack_resource_name column to image_builds 2023121141555: %w", err)
			}
			return nil
		},
	}
}
