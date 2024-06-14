package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type wsProvisionRequestStore struct {
	sessionFactory db.SessionFactory
	statusStore    stores.WorkspaceProvisionRequestStatusStore
}

type WorkspaceProvisionRequestStoreSpec struct {
	SessionFactory              db.SessionFactory
	ProvisionRequestStatusStore stores.WorkspaceProvisionRequestStatusStore
}

func NewWorkspaceProvisionRequestStore(spec WorkspaceProvisionRequestStoreSpec) stores.WorkspaceProvisionRequestStore {
	return &wsProvisionRequestStore{
		sessionFactory: spec.SessionFactory,
		statusStore:    spec.ProvisionRequestStatusStore,
	}
}

func (w *wsProvisionRequestStore) Create(ctx context.Context, spec *models.WorkspaceProvisionRequest) (*models.WorkspaceProvisionRequest, *errors.ServiceError) {
	tx := w.sessionFactory.New(ctx)
	tx.Begin()
	if err := tx.Omit(clause.Associations).Create(&spec).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to create workspace provision request: %s", err.Error())
	}
	tx.Commit()
	return w.GetByID(ctx, spec.ID)
}

func (w *wsProvisionRequestStore) GetByID(ctx context.Context, id string) (*models.WorkspaceProvisionRequest, *errors.ServiceError) {
	grm := w.sessionFactory.New(ctx)
	var res models.WorkspaceProvisionRequest
	if err := grm.Model(&models.WorkspaceProvisionRequest{}).Preload(clause.Associations).Where("id = ?", id).First(&res).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("workspace provision request with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch workspace provision request: %s", err.Error())
	}
	return &res, nil
}
func (w *wsProvisionRequestStore) InternalList(ctx context.Context, query string, args ...any) ([]*models.WorkspaceProvisionRequest, *errors.ServiceError) {
	grm := w.sessionFactory.New(ctx)
	var requests []*models.WorkspaceProvisionRequest
	if err := grm.Model(&models.WorkspaceProvisionRequest{}).
		Preload(clause.Associations).
		Where(query, args...).
		Find(&requests).Error; err != nil {
		return nil, errors.GeneralError("failed to fetch workspace provision requests: %s", err.Error())
	}
	return requests, nil
}

func (w *wsProvisionRequestStore) GetByUserID(ctx context.Context, userID string) (*models.WorkspaceProvisionRequest, *errors.ServiceError) {
	grm := w.sessionFactory.New(ctx)
	var res models.WorkspaceProvisionRequest
	if err := grm.Model(&models.WorkspaceProvisionRequest{}).Preload(clause.Associations).Where("user_id = ?", userID).First(&res).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("workspace provision request with user_id '%s' not found", userID)
		}
		return nil, errors.GeneralError("failed to fetch workspace provision request: %s", err.Error())
	}
	return &res, nil
}

func (w *wsProvisionRequestStore) ListByOrgID(ctx context.Context, orgID string) ([]*models.WorkspaceProvisionRequest, *errors.ServiceError) {
	grm := w.sessionFactory.New(ctx)
	var res []*models.WorkspaceProvisionRequest

	if err := grm.Model(&models.WorkspaceProvisionRequest{}).Preload(clause.Associations).Where("organisation_id = ?", orgID).Find(&res).Error; err != nil {
		return nil, errors.GeneralError("failed to list workspace provision requests for orgID '%s': %s", orgID, err.Error())
	}
	return res, nil
}

func (w *wsProvisionRequestStore) Update(ctx context.Context, id string, spec *models.WorkspaceProvisionRequest) (*models.WorkspaceProvisionRequest, *errors.ServiceError) {
	existingObj, err := w.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tx := w.sessionFactory.New(ctx).Begin()

	if err := tx.Model(&existingObj).Updates(map[string]interface{}{"ssh_public_key": spec.SshPublicKey, "state": spec.State, "message": spec.Message}).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to update object: %s", err.Error())
	}

	if spec.Status != nil {
		spec.Status.WorkspaceProvisionRequestID = id
		if _, err := w.statusStore.Upsert(tx, spec.Status); err != nil {
			tx.Rollback()
			return nil, errors.GeneralError("failed to update object: %s", err.Error())
		}
	}
	tx.Commit()
	return w.GetByID(ctx, id)
}

func (w *wsProvisionRequestStore) PatchStatus(ctx context.Context, id string, status *models.WorkspaceProvisionRequestStatus) (*models.WorkspaceProvisionRequest, *errors.ServiceError) {
	_, err := w.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	grm := w.sessionFactory.New(ctx)
	if status == nil {
		return nil, errors.GeneralError("status is nil")
	}
	status.WorkspaceProvisionRequestID = id
	if _, err := w.statusStore.Upsert(grm, status); err != nil {
		return nil, err
	}
	return w.GetByID(ctx, id)
}

func (w *wsProvisionRequestStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	if _, err := w.GetByID(ctx, id); err != nil {
		return err
	}
	grm := w.sessionFactory.New(ctx)
	if err := grm.Where("id = ?", id).Delete(&models.WorkspaceProvisionRequest{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("workspace provision request with id '%s' not found", id)
		}
		return errors.GeneralError("failed to delete object: %s", err.Error())
	}
	return nil
}
