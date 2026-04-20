package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createPostgresBackupsTable() *gormigrate.Migration {
	type PostgresBackup struct {
		ID              string `gorm:"primary_key;default:gen_random_uuid()"`
		PostgresAddonID string `gorm:"not null;index"`
		Name            string
		Description     string
		Type            string
		Phase           string
		StartedAt       *time.Time
		CompletedAt     *time.Time
		Error           string
		SizeBytes       *int32
		CreatedAt       time.Time
		UpdatedAt       time.Time
	}
	return &gormigrate.Migration{
		ID: "202507162152_create_postgres_backups_table",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&PostgresBackup{}); err != nil {
				return fmt.Errorf("error running postgres backups migration %s: %w", "202507162152_create_postgres_backups_table", err)
			}
			if err := tx.Exec(
				"ALTER TABLE postgres_backups ADD CONSTRAINT fk_postgres_backups_postgres_addon_id FOREIGN KEY (postgres_addon_id) REFERENCES postgres_addons(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint on postgres_backups to postgres_addons table: %w", err)
			}
			return nil
		},
	}
}
