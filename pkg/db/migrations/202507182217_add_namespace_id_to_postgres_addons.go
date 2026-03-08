package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addNamespaceIdToPostgresAddons() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202507182217_add_namespace_id_to_postgres_addons",
		Migrate: func(tx *gorm.DB) error {
			// Add namespace_id column to postgres_addons table
			if err := tx.Exec("ALTER TABLE postgres_addons ADD COLUMN namespace_id TEXT NOT NULL").Error; err != nil {
				return fmt.Errorf("error adding namespace_id column to postgres_addons table: %w", err)
			}

			// Create index on namespace_id
			if err := tx.Exec("CREATE INDEX idx_postgres_addons_namespace_id ON postgres_addons(namespace_id)").Error; err != nil {
				return fmt.Errorf("error creating index on namespace_id: %w", err)
			}

			// Add foreign key constraint to namespaces table
			if err := tx.Exec("ALTER TABLE postgres_addons ADD CONSTRAINT fk_postgres_addons_namespace FOREIGN KEY (namespace_id) REFERENCES namespaces(id)").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint on postgres_addons to namespaces table: %w", err)
			}

			return nil
		},
	}
}
