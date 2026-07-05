package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createRegistryCredentialsTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607040454",
		Migrate: func(tx *gorm.DB) error {
			sql := `
CREATE TABLE registry_credentials (
    id                 TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    organisation_id    TEXT NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    host               TEXT NOT NULL,
    purpose            TEXT NOT NULL DEFAULT 'both',
    username           TEXT NOT NULL,
    encrypted_password TEXT NOT NULL,
    data_hash          TEXT NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (organisation_id, host, purpose)
);
CREATE INDEX idx_registry_credentials_organisation_id ON registry_credentials (organisation_id);
`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to create registry_credentials table: %w", err)
			}
			return nil
		},
	}
}
