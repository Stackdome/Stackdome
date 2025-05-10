package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addWorkspaceNameColumnToStacksTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202505101427",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec("ALTER TABLE stacks ADD COLUMN workspace_name TEXT;").Error; err != nil {
				return fmt.Errorf("error adding workspace_name column to stacks table 202505101427: %w", err)
			}
			return nil
		},
	}
}
