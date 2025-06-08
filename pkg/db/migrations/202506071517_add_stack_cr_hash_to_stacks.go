package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addStackCRHashToStackTable() *gormigrate.Migration {
	type Stack struct {
		CrRevision string `gorm:"default:''"`
	}
	return &gormigrate.Migration{
		ID: "202506071517",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AddColumn(&Stack{}, "CrRevision"); err != nil {
				return err
			}
			return nil
		},
	}
}
