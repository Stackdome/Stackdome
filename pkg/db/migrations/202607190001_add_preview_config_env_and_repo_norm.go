package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/Stackdome/stackdome/pkg/models"
	"gorm.io/gorm"
)

func addPreviewConfigEnvAndRepoNorm() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607190001",
		Migrate: func(tx *gorm.DB) error {
			sql := `
ALTER TABLE stack_preview_configs ADD COLUMN env JSONB NOT NULL DEFAULT '[]';
ALTER TABLE stack_preview_configs ADD COLUMN repo_url_normalized TEXT NOT NULL DEFAULT '';
`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to add preview config env/repo_norm: %w", err)
			}

			var rows []struct {
				ID      string
				RepoURL string
			}
			if err := tx.Raw(`SELECT id, git_repository->>'repo_url' AS repo_url FROM stack_preview_configs`).Scan(&rows).Error; err != nil {
				return fmt.Errorf("failed to read preview configs for repo_norm backfill: %w", err)
			}
			for _, row := range rows {
				if err := tx.Exec(`UPDATE stack_preview_configs SET repo_url_normalized = ? WHERE id = ?`, models.NormalizeRepoURL(row.RepoURL), row.ID).Error; err != nil {
					return fmt.Errorf("failed to backfill repo_url_normalized for %s: %w", row.ID, err)
				}
			}

			if err := tx.Exec(`CREATE UNIQUE INDEX idx_preview_configs_org_repo ON stack_preview_configs (organisation_id, repo_url_normalized);`).Error; err != nil {
				return fmt.Errorf("failed to create preview config org/repo index: %w", err)
			}
			return nil
		},
	}
}
