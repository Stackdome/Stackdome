package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func dropLegacyUsageTables() *gormigrate.Migration {
	type ResourceUsage struct {
		ID           string `gorm:"primary_key;default:gen_random_uuid()"`
		ResourceType string `gorm:"not null"`
		ResourceID   string `gorm:"not null"`
		ConsumerType string `gorm:"not null"`
		ConsumerID   string `gorm:"not null"`
		StackID      string `gorm:"not null"`
		CreatedAt    time.Time
		UpdatedAt    time.Time
	}

	return &gormigrate.Migration{
		ID: "202605250002_drop_legacy_usage_tables",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().DropTable("addon_usages"); err != nil {
				return fmt.Errorf("error dropping addon_usages table: %w", err)
			}
			if err := tx.Migrator().DropTable("secret_usages"); err != nil {
				return fmt.Errorf("error dropping secret_usages table: %w", err)
			}
			if err := tx.AutoMigrate(&ResourceUsage{}); err != nil {
				return fmt.Errorf("error creating resource_usages table: %w", err)
			}
			if err := tx.Exec(
				"ALTER TABLE resource_usages ADD CONSTRAINT fk_resource_usages_stack_id FOREIGN KEY (stack_id) REFERENCES stacks(id) ON DELETE CASCADE",
			).Error; err != nil {
				return fmt.Errorf("error adding resource_usages foreign key: %w", err)
			}
			if err := tx.Exec(
				"CREATE UNIQUE INDEX IF NOT EXISTS idx_resource_usages_unique ON resource_usages(resource_type, resource_id, consumer_type, consumer_id, stack_id)",
			).Error; err != nil {
				return fmt.Errorf("error creating resource_usages unique index: %w", err)
			}
			if err := tx.Exec(
				"CREATE INDEX IF NOT EXISTS idx_resource_usages_resource ON resource_usages(resource_type, resource_id)",
			).Error; err != nil {
				return fmt.Errorf("error creating resource_usages resource index: %w", err)
			}
			return nil
		},
	}
}
