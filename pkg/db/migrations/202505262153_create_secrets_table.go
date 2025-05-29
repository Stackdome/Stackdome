// filepath: pkg/db/migrations/202505262153_create_table_secrets.go
package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createSecretsTable() *gormigrate.Migration {
	type Secret struct {
		ID             string `gorm:"primary_key;default:gen_random_uuid()"`
		OrganisationID string `gorm:"not null;index"`
		UserID         string `gorm:"not null;index"`
		Name           string `gorm:"not null"`
		Description    string
		Type           string   `gorm:"not null"`
		EncryptedData  string   `gorm:"type:text;not null"`
		Keys           []string `gorm:"type:jsonb" json:"keys"`
		DataHash       string   `gorm:"not null"`
		CreatedAt      time.Time
		UpdatedAt      time.Time
	}
	return &gormigrate.Migration{
		ID: "202505262153",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&Secret{}); err != nil {
				return fmt.Errorf("error running secrets migration %s: %w", "202505262153", err)
			}
			if err := tx.Exec(
				"ALTER TABLE secrets ADD CONSTRAINT fk_secrets_organisation_id FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint on secrets to organisations table: %w", err)
			}
			return nil
		},
	}
}
