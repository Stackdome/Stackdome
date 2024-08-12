package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func alterWorkspaceUserTableAddVersioningSupport() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202408111920",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`
				ALTER TABLE workspace_users
				ADD COLUMN IF NOT EXISTS version int DEFAULT 1;
			`).Error; err != nil {
				return fmt.Errorf("error running workspace user migration 202408111920: %w", err)
			}

			// Trigger function to increment version on update
			if err := tx.Exec(`
				CREATE OR REPLACE FUNCTION increment_version()
				RETURNS TRIGGER AS $$
				DECLARE
        			new_status text;
				BEGIN
				-- Store the new status
				new_status := NEW.status;
				
				-- Temporarily set NEW.status to OLD.status
				NEW.status := OLD.status;
				
				-- Check if any other field has changed
				IF NEW.* IS DISTINCT FROM OLD.* THEN
					NEW.version := OLD.version + 1;
				END IF;
				
				-- Restore the new status
				NEW.status := new_status;
				
				RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;
			`).Error; err != nil {
				return fmt.Errorf("error running workspace user migration 202408111920: %w", err)
			}

			// Trigger to call the function on update
			if err := tx.Exec(`
				DROP TRIGGER IF EXISTS increment_version_trigger ON workspace_users;
				CREATE TRIGGER increment_version_trigger
				BEFORE UPDATE ON workspace_users
				FOR EACH ROW EXECUTE FUNCTION increment_version();
			`).Error; err != nil {
				return fmt.Errorf("error running workspace user migration 202408111920: %w", err)
			}
			return nil
		},
	}
}
