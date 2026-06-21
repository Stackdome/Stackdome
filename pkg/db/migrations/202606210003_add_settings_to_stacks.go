package migrations

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addSettingsToStacks() *gormigrate.Migration {
	type Stack struct {
		Settings *models.StackSettings `gorm:"type:jsonb"`
	}
	return &gormigrate.Migration{
		ID: "202606210003",
		Migrate: func(tx *gorm.DB) error {
			return tx.Migrator().AddColumn(&Stack{}, "Settings")
		},
	}
}
