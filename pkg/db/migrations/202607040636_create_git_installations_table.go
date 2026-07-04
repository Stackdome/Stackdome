package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createGitInstallationsTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607040636",
		Migrate: func(tx *gorm.DB) error {
			sql := `
CREATE TABLE git_installations (
    id                   TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    git_integration_id   TEXT NOT NULL REFERENCES git_integrations(id) ON DELETE CASCADE,
    installation_id      BIGINT NOT NULL,
    account_login        TEXT NOT NULL,
    account_type         TEXT,
    repository_selection TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (git_integration_id, installation_id)
);
CREATE INDEX idx_git_installations_integration_id ON git_installations (git_integration_id);
`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to create git_installations table: %w", err)
			}
			return nil
		},
	}
}
