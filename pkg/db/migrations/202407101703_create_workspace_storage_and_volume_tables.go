package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type StackStorage struct {
	ID                string `gorm:"primary_key;default:gen_random_uuid()"`
	OrganisationID    string `gorm:"not null"`
	UserID            string `gorm:"not null"`
	Name              string `gorm:"not null"`
	Namespace         string `gorm:"unique;not null"`
	Labels            []byte `gorm:"type:jsonb"`
	Annotations       []byte `gorm:"type:jsonb"`
	SSHConfig         []byte `gorm:"type:jsonb"`
	Status            []byte `gorm:"type:jsonb"`
	State             string `gorm:"not null"`
	DeletionTimeStamp *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Volume struct {
	ID             string `gorm:"primary_key;default:gen_random_uuid()"`
	OrganisationID string `gorm:"not null"`
	UserID         string `gorm:"not null"`
	Name           string `gorm:"not null"`
	Namespace      string `gorm:"not null"`
	Labels         []byte `gorm:"type:jsonb"`
	Annotations    []byte `gorm:"type:jsonb"`
	Size           string
	StorageClass   string
	AccessMode     string
	VolumeSource   []byte `gorm:"type:jsonb"`
	SyncBeforeUse  bool
	Status         []byte `gorm:"type:jsonb"`
}

func createStackStorageAndVolumeTables() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202407101703",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&StackStorage{}, &Volume{}); err != nil {
				return fmt.Errorf("error running stack storage and volume migration 202407101703: %w", err)
			}

			if err := tx.Exec("ALTER TABLE volumes ADD FOREIGN KEY (user_id) REFERENCES users(id)").Error; err != nil {
				return fmt.Errorf("error adding foreign key to volumes table 202407101703: %w", err)
			}

			if err := tx.Exec("ALTER TABLE volumes ADD FOREIGN KEY (organisation_id) REFERENCES organisations(id)").Error; err != nil {
				return fmt.Errorf("error adding foreign key to volumes table 202407101703: %w", err)
			}

			if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_stack_storage_organisation_id ON stack_storages(organisation_id)").Error; err != nil {
				return fmt.Errorf("error creating index on workspace_storages table 202407101703: %w", err)
			}

			if err := tx.Exec(`ALTER TABLE stack_storages ADD FOREIGN KEY (user_id) REFERENCES users(id)`).Error; err != nil {
				return fmt.Errorf("error adding foreign key to stack_storages table 202407101703: %w", err)
			}

			// Add foreign key for organisations
			if err := tx.Exec(`ALTER TABLE stack_storages ADD FOREIGN KEY (organisation_id) REFERENCES organisations(id)`).Error; err != nil {
				return fmt.Errorf("error adding foreign key to workspace_storages table 202407101703: %w", err)
			}

			if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_stack_storages_user_id ON stack_storages(user_id)`).Error; err != nil {
				return fmt.Errorf("error creating index on workspace_storages table 202407101703: %w", err)
			}

			if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_stack_storages_organisation_id ON stack_storages(organisation_id)`).Error; err != nil {
				return fmt.Errorf("error creating index on stack_storages table 202407101703: %w", err)
			}

			return nil
		},
	}
}
