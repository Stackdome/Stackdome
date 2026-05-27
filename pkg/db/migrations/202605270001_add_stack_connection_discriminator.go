package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addStackConnectionDiscriminator() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202605270001_add_stack_connection_discriminator",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(
				"ALTER TABLE stack_connections ADD COLUMN discriminator TEXT NOT NULL DEFAULT ''",
			).Error; err != nil {
				return fmt.Errorf("error adding discriminator column: %w", err)
			}
			if err := tx.Exec(
				"DROP INDEX IF EXISTS idx_stack_connections_unique_edge",
			).Error; err != nil {
				return fmt.Errorf("error dropping old unique index: %w", err)
			}
			if err := tx.Exec(
				"CREATE UNIQUE INDEX idx_stack_connections_unique_edge ON stack_connections(stack_id, kind, from_ref, to_ref, discriminator)",
			).Error; err != nil {
				return fmt.Errorf("error creating unique index with discriminator: %w", err)
			}
			return nil
		},
	}
}
