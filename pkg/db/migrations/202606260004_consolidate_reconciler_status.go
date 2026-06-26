package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func consolidateReconcilerStatus() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202606260004",
		Migrate: func(tx *gorm.DB) error {
			sql := `
				ALTER TABLE preview_stacks
					DROP COLUMN IF EXISTS stackfile_hash,
					DROP COLUMN IF EXISTS last_applied_commit_sha,
					DROP COLUMN IF EXISTS last_applied_overrides_hash,
					ADD COLUMN reconciler_status JSONB NOT NULL DEFAULT '{}';
			`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to consolidate reconciler status columns: %w", err)
			}
			return nil
		},
	}
}
