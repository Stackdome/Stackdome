package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createOrganisationDomainsTable() *gormigrate.Migration {
	type OrganisationDomain struct {
		ID             string `gorm:"primary_key;default:gen_random_uuid()"`
		Domain         string `gorm:"unique;not null"`
		OrganisationID string `gorm:"not null;index"`
		CreatedAt      time.Time
		UpdatedAt      time.Time
	}
	return &gormigrate.Migration{
		ID: "202505202153",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&OrganisationDomain{}); err != nil {
				return fmt.Errorf("error running domains migration 202505202153: %w", err)
			}
			// Add foreign key constraint for OrganisationDomain
			if err := tx.Exec(
				"ALTER TABLE organisation_domains ADD CONSTRAINT fk_organisation_domains_organisation_id FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint on organisation_domains to organisations table: %w", err)
			}
			return nil
		},
	}
}
