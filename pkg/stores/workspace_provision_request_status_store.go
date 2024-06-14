package stores

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"gorm.io/gorm"
)

type WorkspaceProvisionRequestStatusStore interface {
	Create(tx *gorm.DB, spec *models.WorkspaceProvisionRequestStatus) (*models.WorkspaceProvisionRequestStatus, *errors.ServiceError)
	GetByID(tx *gorm.DB, id string) (*models.WorkspaceProvisionRequestStatus, *errors.ServiceError)
	Upsert(tx *gorm.DB, spec *models.WorkspaceProvisionRequestStatus) (*models.WorkspaceProvisionRequestStatus, *errors.ServiceError)
	Delete(tx *gorm.DB, id string) *errors.ServiceError
}
