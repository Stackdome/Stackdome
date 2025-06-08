package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addDeletionTimestampToStackTable() *gormigrate.Migration {
	type Stack struct {
		DeletionTimestamp *time.Time `gorm:"default:NULL"`
	}
	return &gormigrate.Migration{
		ID: "202505311229",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AddColumn(&Stack{}, "DeletionTimestamp"); err != nil {
				return err
			}
			return nil
		},
	}
}
