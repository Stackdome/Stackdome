package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createResourceReferencesTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202606210001_create_resource_references_table",
		Migrate: func(tx *gorm.DB) error {
			sql := `
CREATE TABLE resource_references (
    id            TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    stack_id      TEXT NOT NULL REFERENCES stacks(id)         ON DELETE CASCADE,
    release_id    TEXT          REFERENCES stack_releases(id) ON DELETE CASCADE,
    referent_type TEXT NOT NULL,
    referent_id   TEXT NOT NULL,
    relation_kind TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_resource_references_referent ON resource_references (referent_type, referent_id);
CREATE INDEX idx_resource_references_stack    ON resource_references (stack_id);
CREATE INDEX idx_resource_references_release  ON resource_references (release_id);
CREATE UNIQUE INDEX uq_resource_references_spec ON resource_references
    (stack_id, referent_type, referent_id, relation_kind) WHERE release_id IS NULL;
CREATE UNIQUE INDEX uq_resource_references_release ON resource_references
    (release_id, referent_type, referent_id, relation_kind) WHERE release_id IS NOT NULL;
`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to create resource_references table: %w", err)
			}
			return nil
		},
	}
}
