package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addGitInstallationIDUniqueIndex() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607200001",
		Migrate: func(tx *gorm.DB) error {
			sql := `CREATE UNIQUE INDEX idx_git_installations_installation_id ON git_installations (installation_id);`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to add unique index on git_installations.installation_id: %w", err)
			}
			return nil
		},
	}
}
