package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func updateUserRoles() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202605040004_update_user_roles",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`UPDATE users SET role = 'OrgAdmin' WHERE role IN ('OrganisationAdmin', 'PlatformAdmin');`).Error; err != nil {
				return fmt.Errorf("failed to migrate admin roles: %w", err)
			}

			if err := tx.Exec(`UPDATE users SET role = '' WHERE role = 'User';`).Error; err != nil {
				return fmt.Errorf("failed to clear User role: %w", err)
			}

			if err := tx.Exec(`ALTER TABLE users DROP COLUMN IF EXISTS default_user;`).Error; err != nil {
				return fmt.Errorf("failed to drop default_user column: %w", err)
			}

			return nil
		},
	}
}
