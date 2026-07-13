package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addStackProjectNameUniqueIndex enforces the per-project stack name uniqueness the
// name-addressed apply endpoint relies on. Stacks are hard-deleted (no gorm
// soft-delete column), so a plain unique index is correct: a name stays taken
// until the row is actually gone, matching GetByNameAndProjectID semantics.
func addStackProjectNameUniqueIndex() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607092148",
		Migrate: func(tx *gorm.DB) error {
			sql := `CREATE UNIQUE INDEX IF NOT EXISTS idx_stacks_project_id_name ON stacks (project_id, name);`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to create unique index on stacks(project_id, name): %w", err)
			}
			return nil
		},
	}
}
