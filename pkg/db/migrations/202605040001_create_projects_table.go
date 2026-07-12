package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createProjectsTable() *gormigrate.Migration {
	type Project struct {
		ID             string    `gorm:"primary_key;default:gen_random_uuid()"`
		Name           string    `gorm:"not null;uniqueIndex:idx_projects_org_name"`
		OrganisationID string    `gorm:"not null;uniqueIndex:idx_projects_org_name"`
		DefaultProject bool      `gorm:"not null;default:false"`
		CreatedAt      time.Time `gorm:"not null;default:now()"`
		UpdatedAt      time.Time `gorm:"not null;default:now()"`
	}
	return &gormigrate.Migration{
		ID: "202605040001_create_projects_table",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&Project{}); err != nil {
				return fmt.Errorf("error creating projects table: %w", err)
			}
			if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_projects_organisation_id ON projects(organisation_id)").Error; err != nil {
				return fmt.Errorf("error creating projects organisation_id index: %w", err)
			}
			if err := tx.Exec("ALTER TABLE projects ADD CONSTRAINT fk_projects_organisation FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE RESTRICT").Error; err != nil {
				return fmt.Errorf("error adding projects foreign key: %w", err)
			}
			return nil
		},
	}
}
