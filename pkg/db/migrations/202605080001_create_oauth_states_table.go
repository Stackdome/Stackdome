package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createOAuthStatesTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202605080001_create_oauth_states_table",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`
				CREATE TABLE IF NOT EXISTS oauth_states (
					id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
					state TEXT NOT NULL,
					provider TEXT NOT NULL,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`).Error; err != nil {
				return fmt.Errorf("failed to create oauth_states table: %w", err)
			}

			if err := tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth_states_state ON oauth_states (state)`).Error; err != nil {
				return fmt.Errorf("failed to create index on oauth_states: %w", err)
			}

			return nil
		},
	}
}
