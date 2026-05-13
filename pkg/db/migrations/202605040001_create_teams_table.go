package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createTeamsTable() *gormigrate.Migration {
	type Team struct {
		ID             string `gorm:"primary_key;default:gen_random_uuid()"`
		Name           string `gorm:"not null;uniqueIndex:idx_teams_org_name"`
		OrganisationID string `gorm:"not null;uniqueIndex:idx_teams_org_name"`
		DefaultTeam    bool   `gorm:"not null;default:false"`
		CreatedAt      time.Time `gorm:"not null;default:now()"`
		UpdatedAt      time.Time `gorm:"not null;default:now()"`
	}
	return &gormigrate.Migration{
		ID: "202605040001_create_teams_table",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&Team{}); err != nil {
				return fmt.Errorf("error creating teams table: %w", err)
			}
			if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_teams_organisation_id ON teams(organisation_id)").Error; err != nil {
				return fmt.Errorf("error creating teams organisation_id index: %w", err)
			}
			if err := tx.Exec("ALTER TABLE teams ADD CONSTRAINT fk_teams_organisation FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE RESTRICT").Error; err != nil {
				return fmt.Errorf("error adding teams foreign key: %w", err)
			}
			return nil
		},
	}
}
