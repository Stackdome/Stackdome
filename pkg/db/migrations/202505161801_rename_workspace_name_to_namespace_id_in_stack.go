package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func renameWorkspaceNameToNamespaceIDInStack() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202505161801_rename_workspace_name_to_namespace_id_in_stack",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec("ALTER TABLE stacks DROP COLUMN IF EXISTS workspace_name;").Error; err != nil {
				return err
			}
			if err := tx.Exec("ALTER TABLE stacks ADD COLUMN namespace_id TEXT NOT NULL;").Error; err != nil {
				return err
			}
			return nil
		},
	}
}
