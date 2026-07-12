package pgstore

import (
	"context"
	stderrors "errors"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StackResourceStoreSpec struct {
	SessionFactory db.SessionFactory
}

type stackResourceStore struct {
	sessionFactory db.SessionFactory
	atomicExecutor
}

func NewStackResourceStore(spec StackResourceStoreSpec) stores.StackResourceStore {
	return &stackResourceStore{
		sessionFactory: spec.SessionFactory,
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
	}
}

func (w *stackResourceStore) Create(ctx context.Context, spec *models.StackResource) (*models.StackResource, *errors.ServiceError) {
	if err := w.sessionFactory.New(ctx).Model(&models.StackResource{}).Omit(clause.Associations).Create(spec).Error; err != nil {
		return nil, errors.GeneralError("failed to create stack resource: %s", err.Error())
	}
	return w.GetByID(ctx, spec.ID)
}

func (w *stackResourceStore) CreateWithTx(ctx context.Context, spec *models.StackResource, stack *models.Stack) (*models.StackResource, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	if err := tx.Model(&models.StackResource{}).Omit(clause.Associations).Create(spec).Error; err != nil {
		return nil, errors.GeneralError("failed to create stack resource: %s", err.Error())
	}
	return w.GetByID(ctx, spec.ID)
}

func (w *stackResourceStore) Update(ctx context.Context, ID string, spec *models.StackResource, stack *models.Stack) (*models.StackResource, *errors.ServiceError) {
	if err := applyStackResourceUpdate(w.sessionFactory.New(ctx), ID, spec); err != nil {
		return nil, errors.GeneralError("failed to update stack resource: %s", err.Error())
	}
	return w.GetByID(ctx, ID)
}

func (w *stackResourceStore) UpdateWithTx(ctx context.Context, ID string, spec *models.StackResource, stack *models.Stack) (*models.StackResource, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	if err := applyStackResourceUpdate(tx, ID, spec); err != nil {
		return nil, errors.GeneralError("failed to update stack resource: %s", err.Error())
	}
	return w.GetByID(ctx, ID)
}

// applyStackResourceUpdate writes all spec columns, including ones zeroed by
// the caller, so cleared fields (e.g. image_config on an image→git switch)
// don't survive the update. Status is agent-owned and never written here.
func applyStackResourceUpdate(dbSession *gorm.DB, ID string, spec *models.StackResource) error {
	return dbSession.Model(&models.StackResource{}).
		Select("*").
		Omit(clause.Associations, "status", "version", "created_at").
		Where("id = ?", ID).
		Updates(spec).Error
}

func (w *stackResourceStore) UpdatePortsWithTx(ctx context.Context, resourceID string, resource *models.StackResource) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}
	if err := tx.Model(&models.StackResource{}).Where("id = ?", resourceID).Updates(map[string]interface{}{
		"ports": resource.Ports,
	}).Error; err != nil {
		return errors.GeneralError("failed to update stack resource ports: %s", err.Error())
	}
	return nil
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
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("stack resource not found")
		}
		return nil, errors.GeneralError("failed to get stack resource: %s", err.Error())
	}
	return &stackResource, nil
}

func (w *stackResourceStore) GetByStackIDAndResourceName(ctx context.Context, stackID, resourceName string) (*models.StackResource, *errors.ServiceError) {
	var stackResource models.StackResource
	if err := w.sessionFactory.New(ctx).Where("stack_id = ? AND name = ?", stackID, resourceName).Preload(clause.Associations).First(&stackResource).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
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
