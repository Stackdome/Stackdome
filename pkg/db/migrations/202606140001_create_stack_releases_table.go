package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createStackReleasesTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202606140001",
		Migrate: func(tx *gorm.DB) error {
			sql := `
CREATE TABLE stack_releases (
    id                TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    stack_id          TEXT NOT NULL REFERENCES stacks(id) ON DELETE CASCADE,
    sequence          INT  NOT NULL,
    state             TEXT NOT NULL,
    message           TEXT,
    cause             JSONB,
    snapshot          JSONB NOT NULL,
    snapshot_revision TEXT  NOT NULL,
    manifest          JSONB,
    manifest_revision TEXT,
    pins              JSONB,
    renderer_version  TEXT,
    outcome           JSONB,
    created_by        TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    rendered_at       TIMESTAMPTZ,
    completed_at      TIMESTAMPTZ,
    UNIQUE (stack_id, sequence)
);
CREATE INDEX idx_stack_releases_stack_state ON stack_releases (stack_id, state);
CREATE INDEX idx_stack_releases_manifest_revision ON stack_releases (manifest_revision);
`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to create stack_releases table: %w", err)
			}

			triggerSQL := `
CREATE OR REPLACE FUNCTION reject_manifest_rewrite() RETURNS trigger AS $$
BEGIN
    IF OLD.manifest IS NOT NULL AND NEW.manifest IS DISTINCT FROM OLD.manifest THEN
        RAISE EXCEPTION 'stack_releases.manifest is immutable once written';
    END IF;
    RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER stack_releases_manifest_frozen BEFORE UPDATE ON stack_releases
    FOR EACH ROW EXECUTE FUNCTION reject_manifest_rewrite();
`
			if err := tx.Exec(triggerSQL).Error; err != nil {
				return fmt.Errorf("failed to create manifest immutability trigger: %w", err)
			}
			return nil
		},
	}
}
