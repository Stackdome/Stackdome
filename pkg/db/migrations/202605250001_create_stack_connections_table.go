package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createStackConnectionsTable() *gormigrate.Migration {
	type StackConnection struct {
		ID        string `gorm:"primary_key;default:gen_random_uuid()"`
		StackID   string `gorm:"not null;index"`
		Kind      string `gorm:"not null"`
		FromRef   []byte `gorm:"column:from_ref;type:jsonb;not null"`
		ToRef     []byte `gorm:"column:to_ref;type:jsonb;not null"`
		Mappings  []byte `gorm:"type:jsonb"`
		Config    []byte `gorm:"type:jsonb"`
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	return &gormigrate.Migration{
		ID: "202605250001_create_stack_connections_table",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&StackConnection{}); err != nil {
				return fmt.Errorf("error creating stack_connections table: %w", err)
			}
			if err := tx.Exec(
				"ALTER TABLE stack_connections ADD CONSTRAINT fk_stack_connections_stack_id FOREIGN KEY (stack_id) REFERENCES stacks(id) ON DELETE CASCADE",
			).Error; err != nil {
				return fmt.Errorf("error adding stack_connections foreign key: %w", err)
			}
			if err := tx.Exec(
				"CREATE INDEX IF NOT EXISTS idx_stack_connections_stack_id ON stack_connections(stack_id)",
			).Error; err != nil {
				return fmt.Errorf("error creating stack_connections stack index: %w", err)
			}
			return nil
		},
	}
}
