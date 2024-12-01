package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createResourceBuildsTable() *gormigrate.Migration {
	type WorkspaceResourceBuild struct {
		ID                  string
		Namespace           string
		WorkspaceID         string
		WorkspaceResourceID string
		BuildSourceHash     string
		ImageRegistry       string
		Status              []byte `gorm:"type:jsonb"`
		CreatedAt           time.Time
		UpdatedAt           time.Time
	}

	return &gormigrate.Migration{
		ID: "202411281957",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&WorkspaceResourceBuild{}); err != nil {
				return fmt.Errorf("error running workspace_resource_builds migration 202411281957: %w", err)
			}
			if err := tx.Exec(`ALTER TABLE workspace_resource_builds ADD FOREIGN KEY (workspace_resource_id) REFERENCES workspace_resources(id) ON DELETE CASCADE;`).Error; err != nil {
				return fmt.Errorf("error adding workpace_resource_id foreign key to workspace_resource_builds table 202411281957: %w", err)
			}

			if err := tx.Exec(`ALTER TABLE workspace_resource_builds ADD FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE;`).Error; err != nil {
				return fmt.Errorf("error adding workpace_id foreign key to workspace_resource_builds table 202411281957: %w", err)
			}
			return nil
		},
	}
}
