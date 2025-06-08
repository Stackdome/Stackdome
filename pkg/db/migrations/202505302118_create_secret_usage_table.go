package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createSecretUsageTable() *gormigrate.Migration {
	type SecretUsage struct {
		SecretID string `gorm:"not null"`
		StackID  string `gorm:"not null"`
	}

	return &gormigrate.Migration{
		ID: "202505302118",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&SecretUsage{}); err != nil {
				return fmt.Errorf("error running secrets migration %s: %w", "202505302118", err)
			}
			if err := tx.Exec(
				"ALTER TABLE secret_usages ADD CONSTRAINT fk_secret_usages_stack_id FOREIGN KEY (stack_id) REFERENCES stacks(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint on secrets to organisations table: %w", err)
			}
			if err := tx.Exec(
				"ALTER TABLE secret_usages ADD CONSTRAINT fk_secret_usages_secret_id FOREIGN KEY (secret_id) REFERENCES secrets(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint on secret_usages to secrets table: %w", err)
			}
			return nil
		},
	}
}
