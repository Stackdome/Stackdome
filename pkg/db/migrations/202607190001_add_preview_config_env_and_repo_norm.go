package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addPreviewConfigEnvAndRepoNorm() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607190001",
		Migrate: func(tx *gorm.DB) error {
			sql := `
ALTER TABLE stack_preview_configs ADD COLUMN env JSONB NOT NULL DEFAULT '[]';
ALTER TABLE stack_preview_configs ADD COLUMN repo_url_normalized TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX idx_preview_configs_org_repo ON stack_preview_configs (organisation_id, repo_url_normalized);
`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to add preview config env/repo_norm: %w", err)
			}
			return nil
		},
	}
}
