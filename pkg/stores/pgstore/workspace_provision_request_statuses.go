package pgstore

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"gorm.io/gorm"
)

type wsProvisionRequestStatusStore struct{}

func NewWorkspaceProvisionRequestStatusStore() stores.WorkspaceProvisionRequestStatusStore {
	return &wsProvisionRequestStatusStore{}
}

func (w *wsProvisionRequestStatusStore) Create(tx *gorm.DB, spec *models.WorkspaceProvisionRequestStatus) (*models.WorkspaceProvisionRequestStatus, *errors.ServiceError) {
	if err := tx.Create(&spec).Error; err != nil {
		return nil, errors.GeneralError("failed to create workspace provision request status: %s", err.Error())
	}
	return w.GetByID(tx, spec.WorkspaceProvisionRequestID)
}

func (w *wsProvisionRequestStatusStore) GetByID(tx *gorm.DB, id string) (*models.WorkspaceProvisionRequestStatus, *errors.ServiceError) {
	var status models.WorkspaceProvisionRequestStatus
	if err := tx.Where("workspace_provision_request_id = ?", id).First(&status).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("workspace provision request status with id %s not found", id)
		}
		return nil, errors.GeneralError("failed to get workspace provision request status: %s", err.Error())
	}
	return &status, nil
}

func (w *wsProvisionRequestStatusStore) Upsert(tx *gorm.DB, spec *models.WorkspaceProvisionRequestStatus) (*models.WorkspaceProvisionRequestStatus, *errors.ServiceError) {
	if _, err := w.GetByID(tx, spec.WorkspaceProvisionRequestID); err != nil {
		if err.Code == errors.ErrorNotFound {
			return w.Create(tx, spec)
		}
		return nil, err
	}
	if err := tx.Where("workspace_provision_request_id = ?", spec.WorkspaceProvisionRequestID).Updates(&spec).Error; err != nil {
		return nil, errors.GeneralError("failed to update workspace provision request status: %s", err.Error())
	}
	return w.GetByID(tx, spec.WorkspaceProvisionRequestID)
}

func (w *wsProvisionRequestStatusStore) Delete(tx *gorm.DB, id string) *errors.ServiceError {
	if err := tx.Where("workspace_provision_request_id = ?", id).Delete(&models.WorkspaceProvisionRequestStatus{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("workspace provision request status with id %s not found", id)
		}
		return errors.GeneralError("failed to delete workspace provision request status: %s", err.Error())
	}
	return nil
}
