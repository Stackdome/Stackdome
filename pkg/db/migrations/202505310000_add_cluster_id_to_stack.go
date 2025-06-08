package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addClusterIdToStackTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202505310000",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec("ALTER TABLE stacks ADD COLUMN cluster_id TEXT NOT NULL").Error; err != nil {
				return fmt.Errorf("error adding cluster_id column to stacks table: %w", err)
			}
			// Add foreign key constraint to cluster_id
			if err := tx.Exec(
				"ALTER TABLE stacks ADD CONSTRAINT fk_stacks_cluster_id FOREIGN KEY (cluster_id) REFERENCES clusters(id)").Error; err != nil {
				return fmt.Errorf("error adding foreign key constraint on stacks to clusters table: %w", err)
			}
			return nil
		},
	}
}
