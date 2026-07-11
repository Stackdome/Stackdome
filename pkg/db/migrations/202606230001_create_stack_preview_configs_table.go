package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createStackPreviewConfigsTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202606230001",
		Migrate: func(tx *gorm.DB) error {
			sql := `
CREATE TABLE stack_preview_configs (
    id                  TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    organisation_id     TEXT NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    project_id             TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id             TEXT NOT NULL REFERENCES users(id),
    name                TEXT NOT NULL,
    description         TEXT,
    git_repository      JSONB NOT NULL,
    stackfile_path      TEXT NOT NULL DEFAULT 'stackfile.yaml',
    max_active_previews INT NOT NULL DEFAULT 10,
    labels              JSONB DEFAULT '[]',
    annotations         JSONB DEFAULT '[]',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, name)
);
CREATE INDEX idx_stack_preview_configs_project_id ON stack_preview_configs (project_id);
CREATE INDEX idx_stack_preview_configs_organisation_id ON stack_preview_configs (organisation_id);
`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to create stack_preview_configs table: %w", err)
			}
			return nil
		},
	}
}
