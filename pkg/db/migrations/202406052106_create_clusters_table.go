package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createClustersTable() *gormigrate.Migration {
	type Cluster struct {
		ID             string `gorm:"primary_key;default:gen_random_uuid()"`
		OrganisationID int    `gorm:"unique;not null"`
		Name           string `gorm:"not null;check:name <> ''"`
		CreatedAt      time.Time
		UpdatedAt      time.Time
		Default        bool
		ClusterURL     string `gorm:"not null;check:cluster_url <> ''"`
		ClusterCAData  string `gorm:"not null;check:cluster_ca_data <> ''"`
		Token          string `gorm:"not null;check:token <> ''"`
	}

	return &gormigrate.Migration{
		ID: "202406052106",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&Cluster{}); err != nil {
				return err
			}
			if err := tx.Exec(`ALTER TABLE clusters ADD FOREIGN KEY (organisation_id) REFERENCES organisations(id) ON DELETE CASCADE;`).Error; err != nil {
				return err
			}
			return nil
		},
	}
}
