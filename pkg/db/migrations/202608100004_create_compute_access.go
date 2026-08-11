package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const createComputeAccessMigrationID = "202608100004_create_compute_access"

// Entitlements model the right to compute independently from leases, which
// continue tracking shared capacity through runtime cleanup.
func createComputeAccess() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: createComputeAccessMigrationID,
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`
CREATE TABLE compute_entitlements (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    organisation_id TEXT NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    source TEXT NOT NULL CHECK (source IN ('trial')),
    status TEXT NOT NULL CHECK (status IN ('active')),
    starts_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organisation_id, source)
)
`).Error; err != nil {
				return fmt.Errorf("failed to create compute entitlements: %w", err)
			}
			if err := tx.Exec(`
CREATE TABLE shared_compute_leases (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    organisation_id TEXT NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    entitlement_id TEXT NOT NULL UNIQUE REFERENCES compute_entitlements(id) ON DELETE CASCADE,
    state TEXT NOT NULL CHECK (state IN ('active', 'cleanup_pending', 'cleaning', 'cleaned', 'error')),
    activated_at TIMESTAMPTZ NOT NULL,
    cleanup_started_at TIMESTAMPTZ,
    cleaned_up_at TIMESTAMPTZ,
    error_at TIMESTAMPTZ,
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)
`).Error; err != nil {
				return fmt.Errorf("failed to create shared compute leases: %w", err)
			}
			if err := tx.Exec(`CREATE UNIQUE INDEX idx_shared_compute_leases_current_org ON shared_compute_leases (organisation_id) WHERE state <> 'cleaned'`).Error; err != nil {
				return fmt.Errorf("failed to constrain current shared compute leases: %w", err)
			}
			if err := tx.Exec(`CREATE INDEX idx_compute_entitlements_expiry ON compute_entitlements (expires_at) WHERE status = 'active'`).Error; err != nil {
				return fmt.Errorf("failed to index active compute entitlements: %w", err)
			}
			if err := tx.Exec(`CREATE INDEX idx_shared_compute_leases_state ON shared_compute_leases (state)`).Error; err != nil {
				return fmt.Errorf("failed to index shared compute lease state: %w", err)
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Exec(`DROP TABLE shared_compute_leases`).Error; err != nil {
				return fmt.Errorf("failed to drop shared compute leases: %w", err)
			}
			if err := tx.Exec(`DROP TABLE compute_entitlements`).Error; err != nil {
				return fmt.Errorf("failed to drop compute entitlements: %w", err)
			}
			return nil
		},
	}
}
