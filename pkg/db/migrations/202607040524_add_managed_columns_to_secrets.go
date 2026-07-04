package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addManagedColumnsToSecrets() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607040524",
		Migrate: func(tx *gorm.DB) error {
			sql := `
ALTER TABLE secrets ADD COLUMN managed BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE secrets ADD COLUMN managed_by_kind TEXT;
ALTER TABLE secrets ADD COLUMN managed_by_id TEXT;
ALTER TABLE secrets ADD COLUMN managed_slot TEXT;
CREATE INDEX idx_secrets_managed_by ON secrets (managed_by_kind, managed_by_id);
`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to add managed columns to secrets: %w", err)
			}
			return nil
		},
	}
}
