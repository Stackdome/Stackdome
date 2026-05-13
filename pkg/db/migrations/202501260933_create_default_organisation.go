package migrations

import (
	"errors"
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createDefaultOrganisation() *gormigrate.Migration {
	type Organisation struct {
		Name    string
		Default bool
	}
	return &gormigrate.Migration{
		ID: "202501260933",
		Migrate: func(tx *gorm.DB) error {
			defaultOrg := &Organisation{
				Name:    "DefaultOrganisation",
				Default: true,
			}
			var org Organisation
			if err := tx.Table("organisations").Where("\"default\" = ?", true).First(&org).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("error checking for default organisation: %w", err)
				}
				if err := tx.Table("organisations").Create(defaultOrg).Error; err != nil {
					return fmt.Errorf("error creating default organisation: %w", err)
				}
			}
			return nil
		},
	}
}
