package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createObjectStoresTable() *gormigrate.Migration {
	type ObjectStore struct {
		ID             string `gorm:"primary_key;default:gen_random_uuid()"`
		OrganisationID string `gorm:"not null;index"`
		Name           string `gorm:"not null"`

		// Spec fields
		Configuration   []byte `gorm:"type:jsonb;not null"`
		DestinationPath string `gorm:"not null"`
		RetentionPolicy string

		// Status fields
		Status    []byte `gorm:"type:jsonb"`
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	return &gormigrate.Migration{
		ID: "202507162152_create_object_stores_table",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&ObjectStore{}); err != nil {
				return fmt.Errorf("error running object stores migration %s: %w", "202507162152_create_object_stores_table", err)
			}
			if err := tx.Exec(
				"ALTER TABLE object_stores ADD CONSTRAINT fk_object_stores_organisation_id FOREIGN KEY (organisation_id) REFERENCES organisations(id)").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint on object_stores to organisations table: %w", err)
			}
			return nil
		},
	}
}
