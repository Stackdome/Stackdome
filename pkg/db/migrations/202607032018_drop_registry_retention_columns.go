package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func dropRegistryRetentionColumns() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607032018",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec("ALTER TABLE cluster_image_registries DROP COLUMN IF EXISTS max_repositories").Error; err != nil {
				return fmt.Errorf("drop max_repositories: %w", err)
			}
			if err := tx.Exec("ALTER TABLE cluster_image_registries DROP COLUMN IF EXISTS tags_per_repository").Error; err != nil {
				return fmt.Errorf("drop tags_per_repository: %w", err)
			}
			if err := tx.Exec("ALTER TABLE cluster_image_registries DROP COLUMN IF EXISTS delete_untagged").Error; err != nil {
				return fmt.Errorf("drop delete_untagged: %w", err)
			}
			return nil
		},
	}
}
