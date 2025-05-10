package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createStackAndStackResourceTables() *gormigrate.Migration {
	type Stack struct {
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

	type StackResource struct {
		ID              string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
		UserID          string `gorm:"not null"`
		StackID         string `gorm:"not null"`
		Name            string `gorm:"not null;<-:create"`
		Labels          []byte `gorm:"type:jsonb"`
		Annotations     []byte `gorm:"type:jsonb"`
		Version         int
		BuildConfig     []byte `gorm:"type:jsonb"`
		ImageConfig     []byte `gorm:"type:jsonb"`
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
		StackID         string
		StackResourceID string
		SourceVolumeID  string
		SourceSubPath   string
		TargetPath      string
	}

	return &gormigrate.Migration{
		ID: "202409131259",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&Stack{}); err != nil {
				return fmt.Errorf("error migrating Stack table: %w", err)
			}

			if err := tx.AutoMigrate(&StackResource{}); err != nil {
				return fmt.Errorf("error migrating StackResource table: %w", err)
			}

			if err := tx.AutoMigrate(&VolumeMount{}); err != nil {
				return fmt.Errorf("error migrating VolumeMount table: %w", err)
			}

			// Add foreign key constraint
			if err := tx.Exec(
				"ALTER TABLE stack_resources ADD CONSTRAINT fk_stack_resources_stack_id FOREIGN KEY (stack_id) REFERENCES stacks(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint to stack_resources table: %w", err)
			}

			// Add unique constraint for StackID and Name
			if err := tx.Exec(
				"ALTER TABLE stack_resources ADD CONSTRAINT stack_resources_stack_id_name UNIQUE (stack_id, name)").Error; err != nil {
				return fmt.Errorf("error adding unique constraint to workspace_resources table: %w", err)
			}

			// Add foreign key constraint for VolumeMount
			if err := tx.Exec(
				"ALTER TABLE volume_mounts ADD CONSTRAINT fk_volume_mounts_stack_resource_id FOREIGN KEY (stack_resource_id) REFERENCES stack_resources(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint on stack_resources to volume_mounts table: %w", err)
			}

			// Add foreign key constraint for VolumeMount
			if err := tx.Exec(
				"ALTER TABLE volume_mounts ADD CONSTRAINT fk_volume_mounts_stack_id FOREIGN KEY (stack_id) REFERENCES stacks(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint on stacks to volume_mounts table: %w", err)
			}

			// Add indexes
			if err := tx.Exec(
				"CREATE INDEX idx_stacks_organisation_id ON stacks(organisation_id)").Error; err != nil {
				return fmt.Errorf("error creating index on stacks(organisation_id): %w", err)
			}

			if err := tx.Exec(
				"CREATE INDEX idx_stacks_user_id ON stacks(user_id)").Error; err != nil {
				return fmt.Errorf("error creating index on workspaces(user_id): %w", err)
			}

			if err := tx.Exec(
				"CREATE INDEX idx_stack_resources_user_id ON stack_resources(user_id)").Error; err != nil {
				return fmt.Errorf("error creating index on stack_resources(user_id): %w", err)
			}

			// Set default value of 1 for version column in Stack and StackResource tables
			if err := tx.Exec(
				"ALTER TABLE stacks ALTER COLUMN version SET DEFAULT 1").Error; err != nil {
				return fmt.Errorf("error setting default value for version column in stacks table: %w", err)
			}

			if err := tx.Exec(
				"ALTER TABLE stack_resources ALTER COLUMN version SET DEFAULT 1").Error; err != nil {
				return fmt.Errorf("error setting default value for version column in stack_resources table: %w", err)
			}

			if err := tx.Exec(`
				DROP TRIGGER IF EXISTS increment_version_trigger ON stacks;
				CREATE TRIGGER increment_version_trigger
				BEFORE UPDATE ON stacks
				FOR EACH ROW EXECUTE FUNCTION increment_version();
			`).Error; err != nil {
				return fmt.Errorf("error creating trigger for stacks table: %w", err)
			}

			// Add trigger to increment version on update for StackResource table
			if err := tx.Exec(`
				DROP TRIGGER IF EXISTS increment_version_trigger ON stack_resources;
				CREATE TRIGGER increment_version_trigger
				BEFORE UPDATE ON stack_resources
				FOR EACH ROW EXECUTE FUNCTION increment_version();
			`).Error; err != nil {
				return fmt.Errorf("error creating trigger for stack_resources table: %w", err)
			}

			return nil
		},
	}
}
