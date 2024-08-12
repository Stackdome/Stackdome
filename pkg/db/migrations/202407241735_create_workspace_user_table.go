package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createWorkspaceUserTable() *gormigrate.Migration {
	type WorkspaceUser struct {
		ID                string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
		UserID            string `gorm:"type:text;uniqueIndex"`
		ClusterID         string
		OrganisationID    string
		SshPublicKey      string
		Status            []byte `gorm:"type:jsonb"`
		State             string
		Message           string
		CreatedAt         time.Time
		UpdatedAt         time.Time
		DeletionTimeStamp *time.Time
	}
	return &gormigrate.Migration{
		ID: "202407241735",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&WorkspaceUser{}); err != nil {
				return fmt.Errorf("error running workspace user migration 202407241735: %w", err)
			}

			if err := tx.Exec(
				"ALTER TABLE workspace_users ADD CONSTRAINT fk_workspace_users_cluster_id FOREIGN KEY (cluster_id) REFERENCES clusters(id)").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint to workspace namespace table: %w", err)
			}
			return nil
		},
	}
}
