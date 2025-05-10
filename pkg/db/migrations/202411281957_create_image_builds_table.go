package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createImageBuildsTable() *gormigrate.Migration {
	type ImageBuild struct {
		ID              string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
		Name            string
		Namespace       string
		StackID         string
		StackResourceID string
		Spec            []byte `gorm:"type:jsonb"`
		Status          []byte `gorm:"type:jsonb"`
		CreatedAt       time.Time
		UpdatedAt       time.Time
	}

	return &gormigrate.Migration{
		ID: "202411281957",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&ImageBuild{}); err != nil {
				return fmt.Errorf("error running image_builds migration 202411281957: %w", err)
			}
			if err := tx.Exec(`ALTER TABLE image_builds ADD FOREIGN KEY (stack_resource_id) REFERENCES stack_resources(id) ON DELETE CASCADE;`).Error; err != nil {
				return fmt.Errorf("error adding stack_resource_id foreign key to image_builds table 202411281957: %w", err)
			}

			if err := tx.Exec(`ALTER TABLE image_builds ADD FOREIGN KEY (stack_id) REFERENCES stacks(id) ON DELETE CASCADE;`).Error; err != nil {
				return fmt.Errorf("error adding stack_id foreign key to image_builds table 202411281957: %w", err)
			}
			return nil
		},
	}
}
