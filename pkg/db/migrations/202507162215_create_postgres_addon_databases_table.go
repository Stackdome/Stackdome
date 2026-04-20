package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createPostgresAddonDatabasesTable() *gormigrate.Migration {
	type PostgresAddonDatabase struct {
		ID              string `gorm:"primary_key;default:gen_random_uuid()"`
		PostgresAddonID string `gorm:"not null;index"`
		Name            string `gorm:"not null"`
		Extensions      []byte `gorm:"type:jsonb"` // Array of strings for PostgreSQL extensions
		CreatedAt       time.Time
		UpdatedAt       time.Time
	}
	return &gormigrate.Migration{
		ID: "202507162215_create_postgres_addon_databases_table",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&PostgresAddonDatabase{}); err != nil {
				return fmt.Errorf("error running postgres addon databases migration %s: %w", "202507162215_create_postgres_addon_databases_table", err)
			}
			if err := tx.Exec(
				"ALTER TABLE postgres_addon_databases ADD CONSTRAINT fk_postgres_addon_databases_postgres_addon_id FOREIGN KEY (postgres_addon_id) REFERENCES postgres_addons(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint on postgres_addon_databases to postgres_addons table: %w", err)
			}
			return nil
		},
	}
}
