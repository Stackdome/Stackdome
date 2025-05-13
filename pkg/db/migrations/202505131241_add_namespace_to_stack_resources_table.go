package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addNamespaceToStackResourcesTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202505131241",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE stack_resources ADD COLUMN namespace TEXT;`).Error; err != nil {
				return fmt.Errorf("error adding namespace column to stack_resources 202505131241: %w", err)
			}
			return nil
		},
	}
}
