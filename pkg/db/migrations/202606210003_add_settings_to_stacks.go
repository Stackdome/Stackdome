package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addSettingsToStacks() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202606210003",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE stacks ADD COLUMN IF NOT EXISTS settings JSONB`).Error; err != nil {
				return fmt.Errorf("failed to add settings column to stacks: %w", err)
			}
			return nil
		},
	}
}
