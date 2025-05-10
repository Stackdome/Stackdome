package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func alterWorkspaceStorageTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202408122225",
		Migrate: func(tx *gorm.DB) error {
			// Add version and workspace column to workspace_storage table
			if err := tx.Exec(`
				ALTER TABLE stack_storages
				ADD COLUMN IF NOT EXISTS version int DEFAULT 1,
				ADD COLUMN IF NOT EXISTS workspace_name TEXT,
				DROP COLUMN IF EXISTS state;
			`).Error; err != nil {
				return fmt.Errorf("error running stack storage migration 202408122225: %w", err)
			}

			// Trigger to call the function on update
			if err := tx.Exec(`
				DROP TRIGGER IF EXISTS increment_version_trigger ON stack_storages;
				CREATE TRIGGER increment_version_trigger
				BEFORE UPDATE ON stack_storages
				FOR EACH ROW EXECUTE FUNCTION increment_version();
			`).Error; err != nil {
				return fmt.Errorf("error creating trigger for stack storage: 202408122225: %w", err)
			}

			return nil
		},
	}
}
