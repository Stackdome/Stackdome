package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addStackfileContentToPreviewStacks() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202606260001",
		Migrate: func(tx *gorm.DB) error {
			sql := `ALTER TABLE preview_stacks ADD COLUMN stackfile_content TEXT;`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to add stackfile_content column to preview_stacks: %w", err)
			}
			return nil
		},
	}
}
