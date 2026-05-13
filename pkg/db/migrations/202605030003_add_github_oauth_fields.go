package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addGitHubOAuthFields() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202605030003_add_github_oauth_fields",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS github_id TEXT").Error; err != nil {
				return fmt.Errorf("error adding github_id column: %w", err)
			}
			if err := tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_users_github_id ON users(github_id) WHERE github_id IS NOT NULL").Error; err != nil {
				return fmt.Errorf("error creating github_id unique index: %w", err)
			}
			if err := tx.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT").Error; err != nil {
				return fmt.Errorf("error adding avatar_url column: %w", err)
			}
			return nil
		},
	}
}
