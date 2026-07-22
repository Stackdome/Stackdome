package migrations

import (
	"fmt"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func renameDefaultToPlatform() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607220001",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE organisations RENAME COLUMN "default" TO platform`).Error; err != nil {
				return fmt.Errorf("rename organisations.default to platform: %w", err)
			}
			if err := tx.Exec(`ALTER TABLE clusters RENAME COLUMN "default" TO platform`).Error; err != nil {
				return fmt.Errorf("rename clusters.default to platform: %w", err)
			}
			if err := tx.Exec(
				`UPDATE organisations SET name = ? WHERE platform = true AND name = ?`,
				models.PlatformOrganisationName, "DefaultOrganisation",
			).Error; err != nil {
				return fmt.Errorf("rename seeded platform organisation: %w", err)
			}
			return nil
		},
	}
}
