package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// deleteOrphanPlatformOrg removes the migration-seeded platform organisation
// when nothing references it. The platform org is now created by the env-driven
// boot bootstrap, so installs without PLATFORM_* config carry no platform org.
func deleteOrphanPlatformOrg() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607230001",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`
				DELETE FROM organisations o
				WHERE o.platform = true
				  AND NOT EXISTS (SELECT 1 FROM users u WHERE u.organisation_id = o.id)
				  AND NOT EXISTS (SELECT 1 FROM projects p WHERE p.organisation_id = o.id)
				  AND NOT EXISTS (SELECT 1 FROM clusters c WHERE c.organisation_id = o.id)
				  AND NOT EXISTS (SELECT 1 FROM organisation_domains d WHERE d.organisation_id = o.id)
				  AND NOT EXISTS (SELECT 1 FROM cluster_image_registries r WHERE r.organisation_id = o.id)
			`).Error; err != nil {
				return fmt.Errorf("delete orphan platform organisation: %w", err)
			}
			return nil
		},
	}
}
