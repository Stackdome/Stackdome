package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createTeamMembershipsTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202605040002_create_team_memberships_table",
		Migrate: func(tx *gorm.DB) error {
			sql := `
				CREATE TABLE IF NOT EXISTS team_memberships (
					id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					team_id    UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
					user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					role       VARCHAR(50) NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT now(),
					UNIQUE(team_id, user_id)
				);
			`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to create team_memberships table: %w", err)
			}

			if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_team_memberships_team_id ON team_memberships(team_id);`).Error; err != nil {
				return fmt.Errorf("failed to create team_memberships team_id index: %w", err)
			}
			if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_team_memberships_user_id ON team_memberships(user_id);`).Error; err != nil {
				return fmt.Errorf("failed to create team_memberships user_id index: %w", err)
			}

			return nil
		},
	}
}
