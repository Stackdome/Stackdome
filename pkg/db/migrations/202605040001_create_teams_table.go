package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createTeamsTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202605040001_create_teams_table",
		Migrate: func(tx *gorm.DB) error {
			sql := `
				CREATE TABLE IF NOT EXISTS teams (
					id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					name            VARCHAR(63) NOT NULL,
					organisation_id UUID NOT NULL REFERENCES organisations(id),
					default_team    BOOLEAN NOT NULL DEFAULT false,
					created_at      TIMESTAMP NOT NULL DEFAULT now(),
					updated_at      TIMESTAMP NOT NULL DEFAULT now(),
					UNIQUE(name, organisation_id)
				);
			`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to create teams table: %w", err)
			}

			if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_teams_organisation_id ON teams(organisation_id);`).Error; err != nil {
				return fmt.Errorf("failed to create teams organisation_id index: %w", err)
			}

			return nil
		},
	}
}
