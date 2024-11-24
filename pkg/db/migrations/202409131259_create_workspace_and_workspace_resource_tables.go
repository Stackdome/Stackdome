package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createWorkspaceAndWorkspaceResourceTables() *gormigrate.Migration {
	type Workspace struct {
		ID             string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
		OrganisationID string `gorm:"not null"`
		UserID         string `gorm:"not null"`
		Name           string `gorm:"not null;<-:create"`
		Namespace      string `gorm:"unique;not null;<-:create"`
		Labels         []byte `gorm:"type:jsonb"`
		Annotations    []byte `gorm:"type:jsonb"`
		Version        int
		Status         []byte `gorm:"type:jsonb"`
		CreatedAt      time.Time
		UpdatedAt      time.Time
	}

	type WorkspaceResource struct {
		ID              string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
		UserID          string `gorm:"not null"`
		WorkspaceID     string `gorm:"not null"`
		Name            string `gorm:"not null;<-:create"`
		Labels          []byte `gorm:"type:jsonb"`
		Annotations     []byte `gorm:"type:jsonb"`
		Version         int
		ImageRegistry   *string
		Build           []byte `gorm:"type:jsonb"`
		Prebuilt        []byte `gorm:"type:jsonb"`
		Init            []byte `gorm:"type:jsonb"`
		ExecutionConfig []byte `gorm:"type:jsonb"`
		DependsOn       []byte `gorm:"type:jsonb"`
		LifecycleConfig []byte `gorm:"type:jsonb"`
		Ports           []byte `gorm:"type:jsonb"`
		StateFul        bool
		Status          []byte `gorm:"type:jsonb"`
		CreatedAt       time.Time
		UpdatedAt       time.Time
	}

	type VolumeMount struct {
		WorkspaceStorageID  string
		WorkspaceResourceID string
		SourceVolumeID      string
		SourceSubPath       string
		TargetPath          string
	}

	return &gormigrate.Migration{
		ID: "202409131259",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&Workspace{}); err != nil {
				return fmt.Errorf("error migrating Workspace table: %w", err)
			}

			if err := tx.AutoMigrate(&WorkspaceResource{}); err != nil {
				return fmt.Errorf("error migrating WorkspaceResource table: %w", err)
			}

			if err := tx.AutoMigrate(&VolumeMount{}); err != nil {
				return fmt.Errorf("error migrating VolumeMount table: %w", err)
			}

			// Add foreign key constraint
			if err := tx.Exec(
				"ALTER TABLE workspace_resources ADD CONSTRAINT fk_workspace_resources_workspace_id FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint to workspace_resources table: %w", err)
			}

			// Add unique constraint for WorkspaceID and Name
			if err := tx.Exec(
				"ALTER TABLE workspace_resources ADD CONSTRAINT uq_workspace_resources_workspace_id_name UNIQUE (workspace_id, name)").Error; err != nil {
				return fmt.Errorf("error adding unique constraint to workspace_resources table: %w", err)
			}

			// Add foreign key constraint for VolumeMount
			if err := tx.Exec(
				"ALTER TABLE volume_mounts ADD CONSTRAINT fk_volume_mounts_workspace_resource_id FOREIGN KEY (workspace_resource_id) REFERENCES workspace_resources(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint to volume_mounts table: %w", err)
			}

			// Add indexes
			if err := tx.Exec(
				"CREATE INDEX idx_workspaces_organisation_id ON workspaces(organisation_id)").Error; err != nil {
				return fmt.Errorf("error creating index on workspaces(organisation_id): %w", err)
			}

			if err := tx.Exec(
				"CREATE INDEX idx_workspaces_user_id ON workspaces(user_id)").Error; err != nil {
				return fmt.Errorf("error creating index on workspaces(user_id): %w", err)
			}

			if err := tx.Exec(
				"CREATE INDEX idx_workspace_resources_user_id ON workspace_resources(user_id)").Error; err != nil {
				return fmt.Errorf("error creating index on workspace_resources(user_id): %w", err)
			}

			// Set default value of 1 for version column in Workspace and WorkspaceResource tables
			if err := tx.Exec(
				"ALTER TABLE workspaces ALTER COLUMN version SET DEFAULT 1").Error; err != nil {
				return fmt.Errorf("error setting default value for version column in workspaces table: %w", err)
			}

			if err := tx.Exec(
				"ALTER TABLE workspace_resources ALTER COLUMN version SET DEFAULT 1").Error; err != nil {
				return fmt.Errorf("error setting default value for version column in workspace_resources table: %w", err)
			}

			if err := tx.Exec(`
				DROP TRIGGER IF EXISTS increment_version_trigger ON workspaces;
				CREATE TRIGGER increment_version_trigger
				BEFORE UPDATE ON workspaces
				FOR EACH ROW EXECUTE FUNCTION increment_version();
			`).Error; err != nil {
				return fmt.Errorf("error creating trigger for workspace table: %w", err)
			}

			// Add trigger to increment version on update for WorkspaceResource table
			if err := tx.Exec(`
				DROP TRIGGER IF EXISTS increment_version_trigger ON workspace_resources;
				CREATE TRIGGER increment_version_trigger
				BEFORE UPDATE ON workspace_resources
				FOR EACH ROW EXECUTE FUNCTION increment_version();
			`).Error; err != nil {
				return fmt.Errorf("error creating trigger for workspace_resources table: %w", err)
			}

			return nil
		},
	}
}
