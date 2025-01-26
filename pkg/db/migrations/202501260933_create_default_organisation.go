package migrations

import (
	"errors"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createDefaultOrganisation() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202501260933",
		Migrate: func(tx *gorm.DB) error {
			defaultOrg := &models.Organisation{
				Name:    "DefaultOrganisation",
				Default: true,
			}
			// check if default organisation already exists
			var org models.Organisation
			if err := tx.Where("\"default\" = ?", true).First(&org).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("error checking for default organisation: %w", err)
				}
				// create default organisation
				if err := tx.Create(defaultOrg).Error; err != nil {
					return fmt.Errorf("error creating default organisation: %w", err)
				}
			}
			return nil
		},
	}
}
