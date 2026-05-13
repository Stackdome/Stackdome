package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createTeamMembershipsTable() *gormigrate.Migration {
	type TeamMembership struct {
		ID        string `gorm:"primary_key;default:gen_random_uuid()"`
		TeamID    string `gorm:"not null;uniqueIndex:idx_team_memberships_team_user"`
		UserID    string `gorm:"not null;uniqueIndex:idx_team_memberships_team_user"`
		Role      string `gorm:"not null"`
		CreatedAt time.Time `gorm:"not null;default:now()"`
	}
	return &gormigrate.Migration{
		ID: "202605040002_create_team_memberships_table",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&TeamMembership{}); err != nil {
				return fmt.Errorf("error creating team_memberships table: %w", err)
			}
			if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_team_memberships_team_id ON team_memberships(team_id)").Error; err != nil {
				return fmt.Errorf("error creating team_memberships team_id index: %w", err)
			}
			if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_team_memberships_user_id ON team_memberships(user_id)").Error; err != nil {
				return fmt.Errorf("error creating team_memberships user_id index: %w", err)
			}
			if err := tx.Exec("ALTER TABLE team_memberships ADD CONSTRAINT fk_team_memberships_team FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding team_memberships team foreign key: %w", err)
			}
			if err := tx.Exec("ALTER TABLE team_memberships ADD CONSTRAINT fk_team_memberships_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding team_memberships user foreign key: %w", err)
			}
			return nil
		},
	}
}
