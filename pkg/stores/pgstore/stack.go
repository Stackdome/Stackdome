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

type StackStoreSpec struct {
	SessionFactory db.SessionFactory
}

type stackStore struct {
	sessionFactory     db.SessionFactory
	stackResourceStore stores.StackResourceStore
	stackVolumeStore   stores.StackVolumeStore
	atomicExecutor
}

func NewStackStore(spec *StackStoreSpec) stores.StackStore {
	return &stackStore{
		sessionFactory: spec.SessionFactory,
		stackResourceStore: NewStackResourceStore(StackResourceStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		stackVolumeStore: NewStackVolumeStore(StackVolumeStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
	}
}

func (w *stackStore) Create(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	tx := w.sessionFactory.New(ctx).Begin()
	ctx = db.CtxWithTransaction(ctx, tx)
	if err := tx.Model(&models.Stack{}).Omit(clause.Associations).Create(spec).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to create stack: %s", err.Error())
	}
	for _, resource := range spec.StackResources {
		resource.StackID = spec.ID
		resource.UserID = spec.UserID
		if _, err := w.stackResourceStore.CreateWithTx(ctx, resource); err != nil {
			tx.Rollback()
			return nil, errors.GeneralError("failed to create stack: errored creating stack resource '%s': %v", resource.Name, err)
		}
	}
	tx.Commit()
	return w.GetByID(ctx, spec.ID)
}

func (w *stackStore) CreateWithTx(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}

	if err := tx.Model(&models.Stack{}).Omit(clause.Associations).Create(spec).Error; err != nil {
		return nil, errors.GeneralError("failed to create stack: %s", err.Error())
	}
	for _, resource := range spec.StackResources {
		resource.StackID = spec.ID
		if _, err := w.stackResourceStore.CreateWithTx(ctx, resource); err != nil {
			return nil, errors.GeneralError("failed to create stack: create stack resource: %v", err)
		}
	}
	return w.GetByID(ctx, spec.ID)
}

func (w *stackStore) ListByUserID(ctx context.Context, userID string) ([]*models.Stack, *errors.ServiceError) {
	var stacks []*models.Stack
	if err := w.sessionFactory.New(ctx).Omit(clause.Associations).Find(&stacks, "user_id = ?", userID).Error; err != nil {
		return nil, errors.GeneralError("failed to list stacks: %s", err.Error())
	}
	for _, stack := range stacks {
		resources, err := w.stackResourceStore.GetByStackID(ctx, stack.ID)
		if err != nil {
			return nil, errors.GeneralError("failed to get stack resources: %v", err)
		}
		stack.StackResources = resources
		stackVolumes, err := w.stackVolumeStore.ListVolumesByStackID(ctx, stack.ID)
		if err != nil {
			return nil, errors.GeneralError("failed to get stack volumes: %v", err)
		}
		stack.Volumes = stackVolumes
	}
	return stacks, nil
}

func (w *stackStore) ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.Stack, *errors.ServiceError) {
	var stacks []*models.Stack
	if err := w.sessionFactory.New(ctx).Omit(clause.Associations).Find(&stacks, "organisation_id = ?", organisationID).Error; err != nil {
		return nil, errors.GeneralError("failed to list stacks: %s", err.Error())
	}
	for _, stack := range stacks {
		resources, err := w.stackResourceStore.GetByStackID(ctx, stack.ID)
		if err != nil {
			return nil, errors.GeneralError("failed to get resources: %v", err)
		}
		stack.StackResources = resources
		stackVolumes, err := w.stackVolumeStore.ListVolumesByStackID(ctx, stack.ID)
		if err != nil {
			return nil, errors.GeneralError("failed to get stack volumes: %v", err)
		}
		stack.Volumes = stackVolumes
	}
	return stacks, nil
}

func (w *stackStore) GetByID(ctx context.Context, id string) (*models.Stack, *errors.ServiceError) {
	var stack models.Stack
	if err := w.sessionFactory.New(ctx).Omit(clause.Associations).First(&stack, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("stack with id %s not found", id)
		}
		return nil, errors.GeneralError("failed to get stack: %s", err.Error())
	}
	resources, err := w.stackResourceStore.GetByStackID(ctx, id)
	if err != nil {
		return nil, errors.GeneralError("failed to get stack resources: %v", err)
	}
	stack.StackResources = resources

	stackVolumes, err := w.stackVolumeStore.ListVolumesByStackID(ctx, id)
	if err != nil {
		return nil, errors.GeneralError("failed to get stack volumes: %v", err)
	}
	stack.Volumes = stackVolumes
	return &stack, nil
}

func (w *stackStore) GetByName(ctx context.Context, name string, userID string) (*models.Stack, *errors.ServiceError) {
	var stack models.Stack
	if err := w.sessionFactory.New(ctx).Omit(clause.Associations).First(&stack, "name = ? AND user_id = ?", name, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("stack with name %s not found", name)
		}
		return nil, errors.GeneralError("failed to get stack: %s", err.Error())
	}
	resources, err := w.stackResourceStore.GetByStackID(ctx, stack.ID)
	if err != nil {
		return nil, errors.GeneralError("failed to get stack resources: %v", err)
	}
	stack.StackResources = resources
	stackVolumes, err := w.stackVolumeStore.ListVolumesByStackID(ctx, stack.ID)
	if err != nil {
		return nil, errors.GeneralError("failed to get stack volumes: %v", err)
	}
	stack.Volumes = stackVolumes
	return &stack, nil
}

func (w *stackStore) Update(ctx context.Context, id string, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	existingStack, err := w.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tx := w.sessionFactory.New(ctx).Begin()
	txCtx := db.CtxWithTransaction(ctx, tx)
	spec.Status = nil
	if err := tx.Model(&models.Stack{}).Omit(clause.Associations).Where("id = ?", id).Updates(spec).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to update stack: %s", err.Error())
	}
	existingResourceMap := existingStack.ResourcesMap()

	for _, resource := range spec.StackResources {
		resource.StackID = existingStack.ID
		resource.UserID = existingStack.UserID
		if currentResource, ok := existingResourceMap[resource.Name]; ok {
			if _, err := w.stackResourceStore.UpdateWithTx(txCtx, currentResource.ID, resource); err != nil {
				tx.Rollback()
				return nil, err
			}
		} else {
			if _, err := w.stackResourceStore.CreateWithTx(txCtx, resource); err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	patchResourceMap := spec.ResourcesMap()

	// Delete resources that are not in the new spec
	for _, resource := range existingStack.StackResources {
		if _, ok := patchResourceMap[resource.Name]; !ok {
			if err := w.stackResourceStore.DeleteWithTx(txCtx, resource.ID); err != nil {
				tx.Rollback()
				return nil, errors.GeneralError("failed to update stack. error deleting stack resource '%s': %v", resource.Name, err)
			}
		}
	}
	tx.Commit()
	return w.GetByID(ctx, id)
}

func (w *stackStore) UpdateWithTx(ctx context.Context, id string, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	existingStack, err := w.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	spec.Status = nil
	if err := tx.Model(&models.Stack{}).Omit(clause.Associations).Where("id = ?", id).Updates(spec).Error; err != nil {
		return nil, errors.GeneralError("failed to update stack: %s", err.Error())
	}
	existingResourceMap := existingStack.ResourcesMap()
	for _, patchResource := range spec.StackResources {
		patchResource.StackID = existingStack.ID
		patchResource.UserID = existingStack.UserID
		if existingResource, ok := existingResourceMap[patchResource.Name]; ok {
			if _, err := w.stackResourceStore.UpdateWithTx(ctx, existingResource.ID, patchResource); err != nil {
				return nil, errors.GeneralError("failed to update stack resource: %v", err)
			}
		} else {
			if _, err := w.stackResourceStore.CreateWithTx(ctx, patchResource); err != nil {
				return nil, errors.GeneralError("failed to create stack resource: %v", err)
			}
		}
	}

	patchResourceMap := spec.ResourcesMap()

	// Delete resources that are not in the new spec
	for _, resource := range existingStack.StackResources {
		if _, ok := patchResourceMap[resource.Name]; !ok {
			if err := w.stackResourceStore.DeleteWithTx(ctx, resource.ID); err != nil {
				return nil, errors.GeneralError("failed to update stack: error deleting stack resource '%s': %v", resource.Name, err)
			}
		}
	}
	return w.GetByID(ctx, id)
}

func (w *stackStore) UpdateStatus(ctx context.Context, id string, status *models.StackStatus) *errors.ServiceError {
	if err := w.sessionFactory.New(ctx).Model(&models.Stack{}).Where("id = ?", id).UpdateColumn("status", status).Error; err != nil {
		return errors.GeneralError("failed to update stack status: %s", err.Error())
	}
	return nil
}

func (w *stackStore) DeleteWithTx(ctx context.Context, id string) *errors.ServiceError {
	tx := w.sessionFactory.New(ctx)
	if err := tx.Delete(&models.Stack{}, "id = ?", id).Error; err != nil {
		return errors.GeneralError("failed to delete stack: %s", err.Error())
	}
	return nil
}

func (w *stackStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	tx := w.sessionFactory.New(ctx)
	if err := tx.Delete(&models.Stack{}, "id = ?", id).Error; err != nil {
		return errors.GeneralError("failed to delete stack: %s", err.Error())
	}
	return nil
}
