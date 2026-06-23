package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func dropResourceUsagesTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202606210002_drop_resource_usages_table",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().DropTable("resource_usages"); err != nil {
				return fmt.Errorf("failed to drop resource_usages table: %w", err)
			}
			return nil
		},
	}
}
