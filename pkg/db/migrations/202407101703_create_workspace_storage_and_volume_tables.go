package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type WorkspaceStorage struct {
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
	ID                 string `gorm:"primaryKey"`
	WorkspaceStorageID string `gorm:"primaryKey"`
	Name               string `gorm:"not null"`
	Labels             []byte `gorm:"type:jsonb"`
	Annotations        []byte `gorm:"type:jsonb"`
	Size               string
	StorageClass       string
	LocalSource        []byte `gorm:"type:jsonb"`
	BuildSource        []byte `gorm:"type:jsonb"`
	SyncBeforeUse      bool
	VolumeStatus       []byte `gorm:"type:jsonb"`
}

func createWorkspaceStorageAndVolumeTables() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202407101703",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&WorkspaceStorage{}, &Volume{}); err != nil {
				return err
			}

			if err := tx.Exec("ALTER TABLE volumes ADD FOREIGN KEY (workspace_storage_id) REFERENCES workspace_storages(id) ON DELETE CASCADE").Error; err != nil {
				return err
			}

			if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_workspace_storage_organisation_id ON workspace_storages(organisation_id)").Error; err != nil {
				return err
			}
			if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_volume_workspace_storage_id ON volumes(workspace_storage_id)").Error; err != nil {
				return err
			}
			if err := tx.Exec(`ALTER TABLE workspace_storages ADD FOREIGN KEY (user_id) REFERENCES users(id)`).Error; err != nil {
				return err
			}

			// Add foreign key for organisations
			if err := tx.Exec(`ALTER TABLE workspace_storages ADD FOREIGN KEY (organisation_id) REFERENCES organisations(id)`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_workspace_storages_user_id ON workspace_storages(user_id)
			`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_workspace_storages_organisation_id ON workspace_storages(organisation_id)`).Error; err != nil {
				return err
			}

			return nil
		},
	}
}
