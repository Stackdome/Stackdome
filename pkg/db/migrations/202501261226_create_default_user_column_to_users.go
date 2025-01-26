package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addDefaultUserColumnToUsersTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202501261226",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec("ALTER TABLE users ADD COLUMN default_user BOOLEAN DEFAULT FALSE;").Error; err != nil {
				return fmt.Errorf("error adding default_user column to users table 202501261226: %w", err)
			}
			return nil
		},
	}
}
