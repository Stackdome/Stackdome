package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addPreviewCommentPending() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607200002",
		Migrate: func(tx *gorm.DB) error {
			sql := `ALTER TABLE preview_stacks ADD COLUMN github_comment_pending BOOLEAN NOT NULL DEFAULT FALSE;`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to add preview_stacks.github_comment_pending: %w", err)
			}
			return nil
		},
	}
}
