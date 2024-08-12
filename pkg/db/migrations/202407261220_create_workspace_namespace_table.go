package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createWorkspaceNamespaceTable() *gormigrate.Migration {
	type WorkspaceNamespace struct {
		UserID          string `gorm:"type:text;primaryKey;not null"`
		WorkspaceUserID string `gorm:"type:text;not null"`
		Namespace       string `gorm:"type:text;uniqueIndex"`
		Workspace       string `gorm:"type:text;primaryKey;not null"`
		Enabled         bool   `gorm:"type:boolean;not null;default:true"`
		CreatedAt       time.Time
		UpdatedAt       time.Time
	}

	return &gormigrate.Migration{
		ID: "202307261200",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&WorkspaceNamespace{}); err != nil {
				return fmt.Errorf("error running workspace namespace migration 202307261200: %w", err)
			}

			// Add foreign key constraint
			if err := tx.Exec(
				"ALTER TABLE workspace_namespaces ADD CONSTRAINT fk_workspace_namespaces_workspace_user FOREIGN KEY (workspace_user_id) REFERENCES workspace_users(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint to workspace namespace table: %w", err)
			}

			return nil
		},
	}
}
