package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createPostgresAddonsTable() *gormigrate.Migration {
	type PostgresAddon struct {
		ID             string `gorm:"primary_key;default:gen_random_uuid()"`
		OrganisationID string `gorm:"not null;index"`
		UserID         string `gorm:"not null;index"`
		ClusterID      string `gorm:"not null;index"`
		Name           string `gorm:"not null"`
		Namespace      string `gorm:"not null"`
		Labels         []byte `gorm:"type:jsonb"`
		Annotations    []byte `gorm:"type:jsonb"`
		Revision       string

		// Spec fields
		PostgresVersion []byte `gorm:"type:jsonb;not null"` // Contains major/minor version info
		Instances       []byte `gorm:"type:jsonb;not null"`
		Resources       []byte `gorm:"type:jsonb;not null"`
		Storage         []byte `gorm:"type:jsonb;not null"`
		Configuration   []byte `gorm:"type:jsonb"`
		Initialization  []byte `gorm:"type:jsonb"`
		BackupConfig    []byte `gorm:"type:jsonb"`

		// Lifecycle fields
		BackupRequestedAt *time.Time `gorm:"default:NULL"`
		LifecycleConfig   []byte     `gorm:"type:jsonb"`

		// Status fields
		Status    []byte `gorm:"type:jsonb"`
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	return &gormigrate.Migration{
		ID: "202507162152_create_postgres_addons_table",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&PostgresAddon{}); err != nil {
				return fmt.Errorf("error running postgres addons migration %s: %w", "202507162152_create_postgres_addons_table", err)
			}
			if err := tx.Exec(
				"ALTER TABLE postgres_addons ADD CONSTRAINT fk_postgres_addons_organisation_id FOREIGN KEY (organisation_id) REFERENCES organisations(id)").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint on postgres_addons to organisations table: %w", err)
			}
			if err := tx.Exec(
				"ALTER TABLE postgres_addons ADD CONSTRAINT fk_postgres_addons_user_id FOREIGN KEY (user_id) REFERENCES users(id)").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint on postgres_addons to users table: %w", err)
			}
			if err := tx.Exec(
				"ALTER TABLE postgres_addons ADD CONSTRAINT fk_postgres_addons_cluster_id FOREIGN KEY (cluster_id) REFERENCES clusters(id)").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint on postgres_addons to clusters table: %w", err)
			}
			return nil
		},
	}
}
