package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addProjectIDToResources() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202605040003_add_project_id_to_resources",
		Migrate: func(tx *gorm.DB) error {
			tables := []string{
				"stacks",
				"secrets",
				"volumes",
				"postgres_addons",
				"object_stores",
				"workspace_users",
			}

			for _, table := range tables {
				addCol := fmt.Sprintf(
					"ALTER TABLE %s ADD COLUMN IF NOT EXISTS project_id TEXT NOT NULL REFERENCES projects(id);",
					table,
				)
				if err := tx.Exec(addCol).Error; err != nil {
					return fmt.Errorf("failed to add project_id to %s: %w", table, err)
				}

				addIdx := fmt.Sprintf(
					"CREATE INDEX IF NOT EXISTS idx_%s_project_id ON %s(project_id);",
					table, table,
				)
				if err := tx.Exec(addIdx).Error; err != nil {
					return fmt.Errorf("failed to create project_id index on %s: %w", table, err)
				}
			}

			return nil
		},
	}
}
