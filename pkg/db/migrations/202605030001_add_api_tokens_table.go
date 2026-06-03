package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addAPITokensTable() *gormigrate.Migration {
	// Timestamps use time.Time so the columns are TIMESTAMPTZ (model reads time.Time).
	type APIToken struct {
		ID          string  `gorm:"primary_key;default:gen_random_uuid()"`
		Name        string  `gorm:"not null"`
		UserID      string  `gorm:"not null"`
		TokenHash   string  `gorm:"not null"`
		TokenPrefix string  `gorm:"not null"`
		Scopes      string  `gorm:"type:jsonb;not null;default:'[]'"`
		ResourceIDs *string `gorm:"type:jsonb"`
		OrgID       string  `gorm:"not null"`
		ExpiresAt   *time.Time
		LastUsedAt  *time.Time
		CreatedAt   time.Time `gorm:"not null"`
		RevokedAt   *time.Time
	}
	return &gormigrate.Migration{
		ID: "202605030001_add_api_tokens_table",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&APIToken{}); err != nil {
				return fmt.Errorf("error creating api_tokens table: %w", err)
			}
			if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_api_tokens_user_id ON api_tokens(user_id)").Error; err != nil {
				return fmt.Errorf("error creating api_tokens user_id index: %w", err)
			}
			if err := tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_api_tokens_token_hash ON api_tokens(token_hash)").Error; err != nil {
				return fmt.Errorf("error creating api_tokens token_hash index: %w", err)
			}
			if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_api_tokens_org_id ON api_tokens(org_id)").Error; err != nil {
				return fmt.Errorf("error creating api_tokens org_id index: %w", err)
			}
			return nil
		},
	}
}
