package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addWorkspaceResourceNameToWorkspaceResourceBuilds() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "2023121141555",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE workspace_resource_builds ADD COLUMN workspace_resource_name TEXT;`).Error; err != nil {
				return fmt.Errorf("err addding workspace_resource_name column to workspace_resource_builds 2023121141555: %w", err)
			}
			return nil
		},
	}
}
