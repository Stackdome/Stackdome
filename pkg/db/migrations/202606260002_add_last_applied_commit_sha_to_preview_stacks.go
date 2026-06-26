package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addLastAppliedCommitSHAToPreviewStacks() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202606260002",
		Migrate: func(tx *gorm.DB) error {
			sql := `ALTER TABLE preview_stacks ADD COLUMN last_applied_commit_sha TEXT NOT NULL DEFAULT '';`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to add last_applied_commit_sha column to preview_stacks: %w", err)
			}
			return nil
		},
	}
}
