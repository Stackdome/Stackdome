package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createWorkspaceProvisionRequestTables() *gormigrate.Migration {
	type WorkspaceProvisionRequestStatus struct {
		WorkspaceProvisionRequestID       string `gorm:"primary_key"`
		WorkspaceNamespace                *string
		WorkspaceServiceAccountName       *string
		WorkspaceServiceAccountToken      *string
		WorkspaceStorageServerSshUsername *string
		ClusterCACert                     *string
		ClusterUrl                        *string
		StatusCondition                   string
		Message                           string
	}

	type WorkspaceProvisionRequest struct {
		ID             string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
		UserID         string
		OrganisationID int
		SshPublicKey   string
		CreatedAt      time.Time
		UpdatedAt      time.Time
		Status         WorkspaceProvisionRequestStatus `gorm:"foreignkey:WorkspaceProvisionRequestID;references:ID;constraint:OnDelete:CASCADE"`
	}

	return &gormigrate.Migration{
		ID: "202406081735",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&WorkspaceProvisionRequest{}, &WorkspaceProvisionRequestStatus{}); err != nil {
				return err
			}
			return nil
		},
	}
}
