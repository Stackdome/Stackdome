package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type ClusterImageRegistry struct {
	ID                  string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	ClusterID           string `gorm:"not null;index:idx_cluster_image_registry_cluster_id" json:"cluster_id"`
	OrganisationID      string `gorm:"not null;index:idx_cluster_image_registry_organisation_id" json:"organisation_id"`
	Name                string `gorm:"not null;check:name <> ''" json:"name"`
	BackendStorageSize  string `gorm:"not null;check:backend_storage_size <> ''" json:"backend_storage_size"`
	BackendStorageClass string
	MaxRepositories     int
	TagsPerRepository   int
	DeleteUntagged      bool
	Status              []byte `gorm:"type:jsonb" json:"status"`
}

func createClusterImageRegistriesTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202505131837",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&ClusterImageRegistry{}); err != nil {
				return err
			}
			if err := tx.Exec("ALTER TABLE cluster_image_registries ADD FOREIGN KEY (cluster_id) REFERENCES clusters(id)").Error; err != nil {
				return err
			}
			if err := tx.Exec("ALTER TABLE cluster_image_registries ADD FOREIGN KEY (organisation_id) REFERENCES organisations(id)").Error; err != nil {
				return err
			}
			return nil
		},
	}
}
