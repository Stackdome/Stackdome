package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const createTrialAllocationsMigrationID = "202608100004_create_trial_allocations"

func createTrialAllocations() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: createTrialAllocationsMigrationID,
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`
CREATE TABLE trial_allocations (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    organisation_id TEXT NOT NULL UNIQUE REFERENCES organisations(id),
    state TEXT NOT NULL CHECK (state IN ('active', 'cleanup_pending', 'cleaning', 'cleaned', 'error')),
    activated_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    cleanup_started_at TIMESTAMPTZ,
    cleaned_up_at TIMESTAMPTZ,
    error_at TIMESTAMPTZ,
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)
`).Error; err != nil {
				return fmt.Errorf("failed to create trial allocations: %w", err)
			}
			if err := tx.Exec(`CREATE INDEX idx_trial_allocations_active_expiry ON trial_allocations (expires_at) WHERE state = 'active'`).Error; err != nil {
				return fmt.Errorf("failed to index active trial allocations: %w", err)
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Exec(`DROP TABLE trial_allocations`).Error; err != nil {
				return fmt.Errorf("failed to drop trial allocations: %w", err)
			}
			return nil
		},
	}
}
