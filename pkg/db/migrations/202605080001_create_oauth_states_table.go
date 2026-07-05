package migrations

import (
	"fmt"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createOAuthStatesTable() *gormigrate.Migration {
	tableName := models.OAuthStateTableName
	return &gormigrate.Migration{
		ID: "202605080001_create_oauth_states_table",
		Migrate: func(tx *gorm.DB) error {
			createSQL := fmt.Sprintf(`
				CREATE TABLE IF NOT EXISTS %s (
					id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
					state TEXT NOT NULL,
					provider TEXT NOT NULL,
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`, tableName)
			if err := tx.Exec(createSQL).Error; err != nil {
				return fmt.Errorf("error creating %s table: %w", tableName, err)
			}
			idx := fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS idx_%s_state ON %s(state)", tableName, tableName)
			if err := tx.Exec(idx).Error; err != nil {
				return fmt.Errorf("error creating %s state index: %w", tableName, err)
			}
			return nil
		},
	}
}
