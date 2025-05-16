package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createDomainsTable() *gormigrate.Migration {
	type Domain struct {
		ID        string `gorm:"primary_key;default:gen_random_uuid()"`
		Fqdn      string `gorm:"unique;not null"`
		OwnerID   string `gorm:"not null;index"`
		OwnerType string `gorm:"not null"`
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	return &gormigrate.Migration{
		ID: "202505141200",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&Domain{}); err != nil {
				return fmt.Errorf("error running domains migration 202505141200: %w", err)
			}
			return nil
		},
	}
}
