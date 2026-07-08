package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createResourceValidationRecords() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607080002_create_resource_validation_records",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`CREATE TABLE IF NOT EXISTS resource_validation_records (
				stack_id text NOT NULL,
				resource_name text NOT NULL,
				check_kind text NOT NULL,
				fingerprint text NOT NULL,
				validated_at timestamptz NOT NULL,
				PRIMARY KEY (stack_id, resource_name, check_kind)
			)`).Error; err != nil {
				return fmt.Errorf("failed to create resource_validation_records: %w", err)
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Exec(`DROP TABLE IF EXISTS resource_validation_records`).Error; err != nil {
				return fmt.Errorf("failed to drop resource_validation_records: %w", err)
			}
			return nil
		},
	}
}
