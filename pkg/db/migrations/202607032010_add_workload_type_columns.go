package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addWorkloadTypeColumns() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607032010",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec("ALTER TABLE stack_resources ADD COLUMN workload_type text NOT NULL DEFAULT 'Service'").Error; err != nil {
				return fmt.Errorf("add workload_type column: %w", err)
			}
			if err := tx.Exec("ALTER TABLE stack_resources ADD COLUMN schedule text").Error; err != nil {
				return fmt.Errorf("add schedule column: %w", err)
			}
			if err := tx.Exec("ALTER TABLE stack_resources ADD COLUMN replicas integer").Error; err != nil {
				return fmt.Errorf("add replicas column: %w", err)
			}
			if err := tx.Exec("ALTER TABLE stack_resources DROP COLUMN IF EXISTS state_ful").Error; err != nil {
				return fmt.Errorf("drop state_ful column: %w", err)
			}
			return nil
		},
	}
}
