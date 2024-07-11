package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createUserAndOrganisationTable() *gormigrate.Migration {
	type Organisation struct {
		ID         string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
		Name       string
		DomainName string // TLD + SLD. Ex: example.test
		Default    bool
		CreatedAt  time.Time
		UpdatedAt  time.Time
	}

	type User struct {
		ID             string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
		InternalID     int    `gorm:"autoIncrement"`
		CreatedAt      time.Time
		UpdatedAt      time.Time
		Name           string
		Email          string `gorm:"unique"`
		Password       string
		Organisation   string
		Role           string
		OrganisationID string
	}

	return &gormigrate.Migration{
		ID: "202405241857",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&User{}); err != nil {
				return err
			}
			if err := tx.Migrator().AutoMigrate(&Organisation{}); err != nil {
				return err
			}

			if err := tx.Exec(`ALTER TABLE users ADD FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE;`).Error; err != nil {
				return err
			}
			return nil
		},
	}
}
