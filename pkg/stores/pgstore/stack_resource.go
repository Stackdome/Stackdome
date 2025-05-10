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

type StackResourceStoreSpec struct {
	SessionFactory db.SessionFactory
}

type stackResourceStore struct {
	sessionFactory   db.SessionFactory
	volumeMountStore stores.VolumeMountStore
	atomicExecutor
}

func NewStackResourceStore(spec StackResourceStoreSpec) stores.StackResourceStore {
	return &stackResourceStore{
		sessionFactory: spec.SessionFactory,
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
		volumeMountStore: NewVolumeMountStore(VolumeMountStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
	}
}

func (w *stackResourceStore) Create(ctx context.Context, spec *models.StackResource) (*models.StackResource, *errors.ServiceError) {
	tx := w.sessionFactory.New(ctx).Begin()
	if err := tx.Model(&models.StackResource{}).Omit(clause.Associations).Create(spec).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to create stack resource: %s", err.Error())
	}
	if len(spec.VolumeMounts) > 0 {
		for _, volumeMount := range spec.VolumeMounts {
			volumeMount.StackID = spec.StackID
			volumeMount.StackResourceID = spec.ID
		}
		if _, err := w.volumeMountStore.BulkCreateWithTx(ctx, spec.VolumeMounts); err != nil {
			tx.Rollback()
			return nil, errors.GeneralError("failed to create stack resource: %s", err.Error())
		}
	}
	tx.Commit()
	return w.GetByID(ctx, spec.ID)
}

func (w *stackResourceStore) CreateWithTx(ctx context.Context, spec *models.StackResource) (*models.StackResource, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	if err := tx.Model(&models.StackResource{}).Omit(clause.Associations).Create(spec).Error; err != nil {
		return nil, errors.GeneralError("failed to create stack resource: %s", err.Error())
	}

	if len(spec.VolumeMounts) > 0 {
		for _, volumeMount := range spec.VolumeMounts {
			volumeMount.StackResourceID = spec.ID
			volumeMount.StackID = spec.StackID
		}
		if _, err := w.volumeMountStore.BulkCreateWithTx(ctx, spec.VolumeMounts); err != nil {
			return nil, errors.GeneralError("failed to create stack resource: %s", err.Error())
		}
	}
	return w.GetByID(ctx, spec.ID)
}

func (w *stackResourceStore) Update(ctx context.Context, ID string, spec *models.StackResource) (*models.StackResource, *errors.ServiceError) {
	tx := w.sessionFactory.New(ctx).Begin()
	ctx = db.CtxWithTransaction(ctx, tx)
	// Update all fields except status
	spec.Status = nil
	if err := tx.Model(&models.StackResource{}).Omit(clause.Associations).Where("id = ?", ID).Updates(spec).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to update stack resource: %s", err.Error())
	}

	if err := w.volumeMountStore.DeleteForStackResourceWithTx(ctx, ID); err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to update stack resource: %s", err.Error())
	}
	for _, volumeMount := range spec.VolumeMounts {
		volumeMount.StackID = spec.StackID
		volumeMount.StackResourceID = ID
	}
	if _, err := w.volumeMountStore.BulkCreateWithTx(ctx, spec.VolumeMounts); err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to update stack resource: %s", err.Error())
	}
	tx.Commit()
	return w.GetByID(ctx, ID)
}

func (w *stackResourceStore) UpdateWithTx(ctx context.Context, ID string, spec *models.StackResource) (*models.StackResource, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	// Update all fields except status
	spec.Status = nil
	if err := tx.Model(&models.StackResource{}).Omit(clause.Associations).Where("id = ?", ID).Updates(spec).Error; err != nil {
		return nil, errors.GeneralError("failed to update stack resource: %s", err.Error())
	}
	if err := w.volumeMountStore.DeleteForStackResourceWithTx(ctx, ID); err != nil {
		return nil, errors.GeneralError("failed to update stack resource: %s", err.Error())
	}
	if len(spec.VolumeMounts) > 0 {
		for _, volumeMount := range spec.VolumeMounts {
			volumeMount.StackResourceID = ID
			volumeMount.StackID = spec.StackID
		}
		if _, err := w.volumeMountStore.BulkCreateWithTx(ctx, spec.VolumeMounts); err != nil {
			return nil, errors.GeneralError("failed to update stack resource: %s", err.Error())
		}
	}
	return w.GetByID(ctx, ID)
}

func (w *stackResourceStore) UpdateStatus(ctx context.Context, ID string, status *models.StackResourceStatus) *errors.ServiceError {
	if err := w.sessionFactory.New(ctx).Model(&models.StackResource{}).Where("id = ?", ID).UpdateColumn("status", status).Error; err != nil {
		return errors.GeneralError("failed to update stack resource status: %s", err.Error())
	}
	return nil
}

func (w *stackResourceStore) GetByID(ctx context.Context, ID string) (*models.StackResource, *errors.ServiceError) {
	var stackResource models.StackResource
	if err := w.sessionFactory.New(ctx).Where("id = ?", ID).Preload(clause.Associations).First(&stackResource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("stack resource not found")
		}
		return nil, errors.GeneralError("failed to get stack resource: %s", err.Error())
	}
	return &stackResource, nil
}

func (w *stackResourceStore) GetByStackIDAndResourceName(ctx context.Context, stackID, resourceName string) (*models.StackResource, *errors.ServiceError) {
	var stackResource models.StackResource
	if err := w.sessionFactory.New(ctx).Where("stack_id = ? AND name = ?", stackID, resourceName).Preload(clause.Associations).First(&stackResource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("stack resource not found")
		}
		return nil, errors.GeneralError("failed to get stack resource: %s", err.Error())
	}
	return &stackResource, nil
}

func (w *stackResourceStore) GetByStackID(ctx context.Context, stackID string) ([]*models.StackResource, *errors.ServiceError) {
	var stackResources []*models.StackResource
	if err := w.sessionFactory.New(ctx).Where("stack_id = ?", stackID).Preload(clause.Associations).Find(&stackResources).Error; err != nil {
		return nil, errors.GeneralError("failed to get stack resources: %s", err.Error())
	}
	return stackResources, nil
}

func (w *stackResourceStore) DeleteWithTx(ctx context.Context, ID string) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}
	if err := tx.Where("id = ?", ID).Delete(&models.StackResource{}).Error; err != nil {
		return errors.GeneralError("failed to delete stack resource: %s", err.Error())
	}
	return nil
}

func (w *stackResourceStore) Delete(ctx context.Context, ID string) *errors.ServiceError {
	if err := w.sessionFactory.New(ctx).Where("id = ?", ID).Delete(&models.StackResource{}).Error; err != nil {
		return errors.GeneralError("failed to delete stack resource: %s", err.Error())
	}
	return nil
}
