package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addResourceValidationRecordsStackForeignKey() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202609190001_add_resource_validation_records_stack_fk",
		Migrate: func(tx *gorm.DB) error {
			// Orphaned rows would block the constraint, so clear them first.
			if err := deleteOrphanedResourceValidationRecords(tx); err != nil {
				return err
			}
			if err := tx.Exec(`ALTER TABLE resource_validation_records
				ADD CONSTRAINT fk_resource_validation_records_stack_id
				FOREIGN KEY (stack_id) REFERENCES stacks(id) ON DELETE CASCADE`).Error; err != nil {
				return fmt.Errorf("failed to add stack foreign key to resource_validation_records: %w", err)
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE resource_validation_records
				DROP CONSTRAINT IF EXISTS fk_resource_validation_records_stack_id`).Error; err != nil {
				return fmt.Errorf("failed to drop stack foreign key from resource_validation_records: %w", err)
			}
			return nil
		},
	}
}

func deleteOrphanedResourceValidationRecords(tx *gorm.DB) error {
	if err := tx.Exec(`DELETE FROM resource_validation_records
		WHERE NOT EXISTS (
			SELECT 1 FROM stacks s
			WHERE s.id = resource_validation_records.stack_id
		)`).Error; err != nil {
		return fmt.Errorf("failed to delete orphaned resource_validation_records: %w", err)
	}
	return nil
}
