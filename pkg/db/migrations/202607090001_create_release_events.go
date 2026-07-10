package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createReleaseEvents() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607090001_create_release_events",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`
CREATE TABLE IF NOT EXISTS release_events (
    id             TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    release_id     TEXT NOT NULL REFERENCES stack_releases(id) ON DELETE CASCADE,
    stack_id       TEXT NOT NULL REFERENCES stacks(id) ON DELETE CASCADE,
    sequence       INT NOT NULL,
    occurred_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    source         TEXT NOT NULL,
    scope          TEXT NOT NULL,
    resource_name  TEXT,
    type           TEXT NOT NULL,
    level          TEXT NOT NULL,
    message        TEXT NOT NULL,
    dedupe_key     TEXT NOT NULL,
    links          JSONB,
    metadata       JSONB,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (release_id, sequence),
    UNIQUE (release_id, dedupe_key)
);
CREATE INDEX IF NOT EXISTS idx_release_events_release_sequence
    ON release_events (release_id, sequence);
CREATE INDEX IF NOT EXISTS idx_release_events_stack
    ON release_events (stack_id);
`).Error; err != nil {
				return fmt.Errorf("failed to create release_events table: %w", err)
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Exec(`DROP TABLE IF EXISTS release_events`).Error; err != nil {
				return fmt.Errorf("failed to drop release_events table: %w", err)
			}
			return nil
		},
	}
}
