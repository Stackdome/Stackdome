package migrations

import (
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addInviteTokenToOAuthStates() *gormigrate.Migration {
	tableName := models.OAuthStateTableName
	return &gormigrate.Migration{
		ID: "202605100002_add_invite_token_to_oauth_states",
		Migrate: func(tx *gorm.DB) error {
			sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS invite_token TEXT", tableName)
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("error adding invite_token to %s: %w", tableName, err)
			}
			return nil
		},
	}
}
