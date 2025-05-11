package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func removeOrganisationColumnFromUsers() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202505111557",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec("ALTER TABLE users DROP COLUMN organisation;").Error; err != nil {
				return fmt.Errorf("error removing organisation column from users table 202505111557: %w", err)
			}
			return nil
		},
	}
}
