package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// addStackTeamNameUniqueIndex enforces the per-team stack name uniqueness the
// name-addressed apply endpoint relies on. Stacks are hard-deleted (no gorm
// soft-delete column), so a plain unique index is correct: a name stays taken
// until the row is actually gone, matching GetByNameAndTeamID semantics.
func addStackTeamNameUniqueIndex() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607092148",
		Migrate: func(tx *gorm.DB) error {
			sql := `CREATE UNIQUE INDEX IF NOT EXISTS idx_stacks_team_id_name ON stacks (team_id, name);`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to create unique index on stacks(team_id, name): %w", err)
			}
			return nil
		},
	}
}
