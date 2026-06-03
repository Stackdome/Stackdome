package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addRefreshTokensTable() *gormigrate.Migration {
	// Timestamps use time.Time so the columns are TIMESTAMPTZ (model reads time.Time).
	type RefreshToken struct {
		ID        string    `gorm:"primary_key;default:gen_random_uuid()"`
		UserID    string    `gorm:"not null"`
		TokenHash string    `gorm:"not null"`
		ExpiresAt time.Time `gorm:"not null"`
		CreatedAt time.Time `gorm:"not null"`
		RevokedAt *time.Time
	}
	return &gormigrate.Migration{
		ID: "202605030002_add_refresh_tokens_table",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&RefreshToken{}); err != nil {
				return fmt.Errorf("error creating refresh_tokens table: %w", err)
			}
			if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id)").Error; err != nil {
				return fmt.Errorf("error creating refresh_tokens user_id index: %w", err)
			}
			if err := tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash)").Error; err != nil {
				return fmt.Errorf("error creating refresh_tokens token_hash index: %w", err)
			}
			return nil
		},
	}
}
