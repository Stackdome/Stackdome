package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createAddonUsagesTable() *gormigrate.Migration {
	type AddonUsage struct {
		ID              string `gorm:"primary_key;default:gen_random_uuid()"`
		AddonType       string `gorm:"not null"`
		AddonID         string `gorm:"not null"`
		StackID         string `gorm:"not null"`
		StackResourceID string `gorm:"not null"`
	}
	return &gormigrate.Migration{
		ID: "202603161001_create_addon_usages_table",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&AddonUsage{}); err != nil {
				return fmt.Errorf("error creating addon_usages table: %w", err)
			}
			if err := tx.Exec(
				"ALTER TABLE addon_usages ADD CONSTRAINT uq_addon_usages UNIQUE (addon_type, addon_id, stack_id, stack_resource_id)",
			).Error; err != nil {
				return fmt.Errorf("error adding unique constraint on addon_usages: %w", err)
			}
			if err := tx.Exec(
				"CREATE INDEX IF NOT EXISTS idx_addon_usages_addon ON addon_usages(addon_type, addon_id)",
			).Error; err != nil {
				return fmt.Errorf("error creating addon_usages addon index: %w", err)
			}
			if err := tx.Exec(
				"CREATE INDEX IF NOT EXISTS idx_addon_usages_stack ON addon_usages(stack_id)",
			).Error; err != nil {
				return fmt.Errorf("error creating addon_usages stack index: %w", err)
			}
			return nil
		},
	}
}
