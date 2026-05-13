package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createOrgInvitesTable() *gormigrate.Migration {
	type OrgInvite struct {
		ID             string    `gorm:"primary_key;default:gen_random_uuid()"`
		Email          string    `gorm:"not null;index:idx_org_invites_email"`
		OrganisationID string    `gorm:"not null;index:idx_org_invites_org_id"`
		TeamID         string    `gorm:"not null"`
		TeamRole       string    `gorm:"not null"`
		TokenHash      string    `gorm:"not null;uniqueIndex:idx_org_invites_token_hash"`
		EncryptedToken string    `gorm:"not null"`
		Status         string    `gorm:"not null;default:pending"`
		ExpiresAt      time.Time `gorm:"not null"`
		InvitedByID    string    `gorm:"not null"`
		AcceptedAt     *time.Time
		EmailSent      bool `gorm:"not null;default:false"`
		EmailSentAt    *time.Time
		EmailError     *string
		CreatedAt      time.Time `gorm:"not null;default:now()"`
	}
	return &gormigrate.Migration{
		ID: "202605100001_create_org_invites_table",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&OrgInvite{}); err != nil {
				return fmt.Errorf("error creating org_invites table: %w", err)
			}
			if err := tx.Exec(
				"ALTER TABLE org_invites ADD CONSTRAINT fk_org_invites_org FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE",
			).Error; err != nil {
				return fmt.Errorf("error adding org FK on org_invites: %w", err)
			}
			if err := tx.Exec(
				"ALTER TABLE org_invites ADD CONSTRAINT fk_org_invites_team FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE",
			).Error; err != nil {
				return fmt.Errorf("error adding team FK on org_invites: %w", err)
			}
			if err := tx.Exec(
				"ALTER TABLE org_invites ADD CONSTRAINT fk_org_invites_invited_by FOREIGN KEY (invited_by_id) REFERENCES users(id) ON DELETE CASCADE",
			).Error; err != nil {
				return fmt.Errorf("error adding invited_by FK on org_invites: %w", err)
			}
			if err := tx.Exec(
				"CREATE UNIQUE INDEX IF NOT EXISTS idx_org_invites_email_org_pending ON org_invites(email, organisation_id) WHERE status = 'pending'",
			).Error; err != nil {
				return fmt.Errorf("error creating partial unique index on org_invites: %w", err)
			}
			return nil
		},
	}
}
