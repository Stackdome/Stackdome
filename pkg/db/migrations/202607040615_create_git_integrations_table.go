package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createGitIntegrationsTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607040615",
		Migrate: func(tx *gorm.DB) error {
			sql := `
CREATE TABLE git_integrations (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    organisation_id TEXT NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    type            TEXT NOT NULL,
    host            TEXT,
    encrypted_auth  TEXT,
    status          TEXT NOT NULL DEFAULT 'active',
    data_hash       TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_git_integrations_organisation_id ON git_integrations (organisation_id);
CREATE UNIQUE INDEX idx_git_integrations_org_host_credentials
    ON git_integrations (organisation_id, host) WHERE type = 'git_credentials';
CREATE UNIQUE INDEX idx_git_integrations_org_github_app
    ON git_integrations (organisation_id) WHERE type = 'github_app';
`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to create git_integrations table: %w", err)
			}
			return nil
		},
	}
}
