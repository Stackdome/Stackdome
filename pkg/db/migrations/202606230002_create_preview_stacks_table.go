package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createPreviewStacksTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202606230002",
		Migrate: func(tx *gorm.DB) error {
			sql := `
CREATE TABLE preview_stacks (
    id                       TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    organisation_id          TEXT NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    project_id                  TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id                  TEXT NOT NULL REFERENCES users(id),
    stack_preview_config_id  TEXT NOT NULL REFERENCES stack_preview_configs(id) ON DELETE CASCADE,
    stack_id                 TEXT REFERENCES stacks(id) ON DELETE SET NULL,
    active_release_id        TEXT REFERENCES stack_releases(id) ON DELETE SET NULL,
    name                     TEXT NOT NULL,
    pr_number                TEXT NOT NULL,
    branch                   TEXT NOT NULL,
    commit_sha               TEXT NOT NULL DEFAULT '',
    source                   TEXT NOT NULL,
    image_overrides          JSONB,
    labels                   JSONB DEFAULT '[]',
    annotations              JSONB DEFAULT '[]',
    status                   JSONB NOT NULL,
    stackfile_content        TEXT,
    reconciler_status        JSONB NOT NULL DEFAULT '{}',
    force_sync_requested_at  TIMESTAMPTZ,
    deletion_timestamp       TIMESTAMPTZ,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (stack_preview_config_id, pr_number)
);
CREATE INDEX idx_preview_stacks_config_id ON preview_stacks (stack_preview_config_id);
CREATE INDEX idx_preview_stacks_stack_id ON preview_stacks (stack_id);
CREATE INDEX idx_preview_stacks_project_id ON preview_stacks (project_id);
CREATE INDEX idx_preview_stacks_org_id ON preview_stacks (organisation_id);
`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to create preview_stacks table: %w", err)
			}
			return nil
		},
	}
}
