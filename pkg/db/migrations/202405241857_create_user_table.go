package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createUserTable() *gormigrate.Migration {
	type User struct {
		ID           string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
		CreatedAt    time.Time
		UpdatedAt    time.Time
		Name         string
		Email        string `gorm:"unique"`
		Password     string
		Organisation string
		Role         string
	}
	return &gormigrate.Migration{
		ID: "202405241857",
		Migrate: func(tx *gorm.DB) error {
			return tx.Migrator().AutoMigrate(&User{})
		},
	}
}
