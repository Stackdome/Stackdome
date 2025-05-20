package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func replaceDomainWithStackDomainsTable() *gormigrate.Migration {
	type StackDomain struct {
		ID                string `gorm:"primary_key;default:gen_random_uuid()"`
		OrganisationID    string `gorm:"not null"`
		Fqdn              string `gorm:"unique;not null"`
		StackID           string `gorm:"not null;index"`
		StackResourceID   string `gorm:"not null;index"`
		StackResourceName string `gorm:"not null"`
		TargetPort        int    `gorm:"not null"`
		CreatedAt         time.Time
		UpdatedAt         time.Time
	}

	return &gormigrate.Migration{
		ID: "202505202237", // Current timestamp
		Migrate: func(tx *gorm.DB) error {
			if tx.Migrator().HasTable("domains") {
				if err := tx.Migrator().DropTable("domains"); err != nil {
					return fmt.Errorf("error dropping domains table: %w", err)
				}
			}
			if err := tx.Migrator().AutoMigrate(&StackDomain{}); err != nil {
				return fmt.Errorf("error creating stack_domains table: %w", err)
			}
			if err := tx.Exec(
				"ALTER TABLE stack_domains ADD CONSTRAINT fk_stack_domains_stack_id FOREIGN KEY (stack_id) REFERENCES stacks(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint on stack_domains to stacks table: %w", err)
			}
			if err := tx.Exec(
				"ALTER TABLE stack_domains ADD CONSTRAINT fk_stack_domains_stack_resource_id FOREIGN KEY (stack_resource_id) REFERENCES stack_resources(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint on stack_domains to stack_resources table: %w", err)
			}
			if err := tx.Exec(
				"ALTER TABLE stack_domains ADD CONSTRAINT fk_stack_domains_organisation_id FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint on stack_domains to organisations table: %w", err)
			}
			return nil
		},
	}
}
