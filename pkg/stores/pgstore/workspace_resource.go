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

type WorkspaceResourceStoreSpec struct {
	SessionFactory db.SessionFactory
}

type workspaceResourceStore struct {
	sessionFactory db.SessionFactory
	atomicExecutor
}

func NewWorkspaceResourceStore(spec WorkspaceResourceStoreSpec) stores.WorkspaceResourceStore {
	return &workspaceResourceStore{
		sessionFactory: spec.SessionFactory,
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
	}
}

func (w *workspaceResourceStore) Create(ctx context.Context, spec *models.WorkspaceResource) (*models.WorkspaceResource, *errors.ServiceError) {
	tx := w.sessionFactory.New(ctx).Begin()
	if err := tx.Model(&models.WorkspaceResource{}).Omit(clause.Associations).Create(spec).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to create workspace resource: %s", err.Error())
	}
	for _, volumeMount := range spec.VolumeMounts {
		volumeMount.WorkspaceResourceID = spec.ID
	}
	if err := tx.Model(&models.VolumeMount{}).Omit(clause.Associations).Create(spec.VolumeMounts).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to create workspace resource: %s", err.Error())
	}
	tx.Commit()
	return w.GetByID(ctx, spec.ID)
}

func (w *workspaceResourceStore) CreateWithTx(ctx context.Context, spec *models.WorkspaceResource) (*models.WorkspaceResource, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	if err := tx.Model(&models.WorkspaceResource{}).Omit(clause.Associations).Create(spec).Error; err != nil {
		return nil, errors.GeneralError("failed to create workspace resource: %s", err.Error())
	}

	if len(spec.VolumeMounts) > 0 {
		for _, volumeMount := range spec.VolumeMounts {
			volumeMount.WorkspaceResourceID = spec.ID
		}
		for _, volumeMount := range spec.VolumeMounts {
			volumeMount.WorkspaceResourceID = spec.ID
			// create volume mount
			if err := tx.Model(&models.VolumeMount{}).Omit(clause.Associations).Create(volumeMount).Error; err != nil {
				return nil, errors.GeneralError("failed to create workspace resource: %s", err.Error())
			}
		}
	}

	return w.GetByID(ctx, spec.ID)
}

func (w *workspaceResourceStore) Update(ctx context.Context, ID string, spec *models.WorkspaceResource) (*models.WorkspaceResource, *errors.ServiceError) {
	tx := w.sessionFactory.New(ctx).Begin()
	// Update all fields except status
	spec.Status = nil
	if err := tx.Model(&models.WorkspaceResource{}).Omit(clause.Associations).Where("id = ?", ID).Updates(spec).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to update workspace resource: %s", err.Error())
	}

	if err := tx.Where("workspace_resource_id = ?", ID).Delete(&models.VolumeMount{}).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to update workspace resource: %s", err.Error())
	}
	for _, volumeMount := range spec.VolumeMounts {
		volumeMount.WorkspaceResourceID = ID
	}
	if err := tx.Model(&models.VolumeMount{}).Omit(clause.Associations).Create(spec.VolumeMounts).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to update workspace resource: %s", err.Error())
	}
	tx.Commit()
	return w.GetByID(ctx, ID)
}

func (w *workspaceResourceStore) UpdateWithTx(ctx context.Context, ID string, spec *models.WorkspaceResource) (*models.WorkspaceResource, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	// Update all fields except status
	spec.Status = nil
	if err := tx.Model(&models.WorkspaceResource{}).Omit(clause.Associations).Where("id = ?", ID).Updates(spec).Error; err != nil {
		return nil, errors.GeneralError("failed to update workspace resource: %s", err.Error())
	}
	if err := tx.Where("workspace_resource_id = ?", ID).Delete(&models.VolumeMount{}).Error; err != nil {
		return nil, errors.GeneralError("failed to update workspace resource: %s", err.Error())
	}
	if len(spec.VolumeMounts) > 0 {
		for _, volumeMount := range spec.VolumeMounts {
			volumeMount.WorkspaceResourceID = ID
			// create volume mount
			if err := tx.Model(&models.VolumeMount{}).Omit(clause.Associations).Create(volumeMount).Error; err != nil {
				return nil, errors.GeneralError("failed to create workspace resource: %s", err.Error())
			}
		}
	}
	return w.GetByID(ctx, ID)
}

func (w *workspaceResourceStore) UpdateStatus(ctx context.Context, ID string, status *models.WorkspaceResourceStatus) *errors.ServiceError {
	if err := w.sessionFactory.New(ctx).Model(&models.WorkspaceResource{}).Where("id = ?", ID).UpdateColumn("status", status).Error; err != nil {
		return errors.GeneralError("failed to update workspace resource status: %s", err.Error())
	}
	return nil
}

func (w *workspaceResourceStore) GetByID(ctx context.Context, ID string) (*models.WorkspaceResource, *errors.ServiceError) {
	var workspaceResource models.WorkspaceResource
	if err := w.sessionFactory.New(ctx).Where("id = ?", ID).Preload(clause.Associations).First(&workspaceResource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("workspace resource not found")
		}
		return nil, errors.GeneralError("failed to get workspace resource: %s", err.Error())
	}
	return &workspaceResource, nil
}

func (w *workspaceResourceStore) GetByWorkspaceIDAndResourceName(ctx context.Context, workspaceID, resourceName string) (*models.WorkspaceResource, *errors.ServiceError) {
	var workspaceResource models.WorkspaceResource
	if err := w.sessionFactory.New(ctx).Where("workspace_id = ? AND name = ?", workspaceID, resourceName).Preload(clause.Associations).First(&workspaceResource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("workspace resource not found")
		}
		return nil, errors.GeneralError("failed to get workspace resource: %s", err.Error())
	}
	return &workspaceResource, nil
}

func (w *workspaceResourceStore) GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*models.WorkspaceResource, *errors.ServiceError) {
	var workspaceResources []*models.WorkspaceResource
	if err := w.sessionFactory.New(ctx).Where("workspace_id = ?", workspaceID).Preload(clause.Associations).Find(&workspaceResources).Error; err != nil {
		return nil, errors.GeneralError("failed to get workspace resources: %s", err.Error())
	}
	return workspaceResources, nil
}

func (w *workspaceResourceStore) DeleteWithTx(ctx context.Context, ID string) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}
	if err := tx.Where("id = ?", ID).Delete(&models.WorkspaceResource{}).Error; err != nil {
		return errors.GeneralError("failed to delete workspace resource: %s", err.Error())
	}
	return nil
}

func (w *workspaceResourceStore) Delete(ctx context.Context, ID string) *errors.ServiceError {
	if err := w.sessionFactory.New(ctx).Where("id = ?", ID).Delete(&models.WorkspaceResource{}).Error; err != nil {
		return errors.GeneralError("failed to delete workspace resource: %s", err.Error())
	}
	return nil
}
