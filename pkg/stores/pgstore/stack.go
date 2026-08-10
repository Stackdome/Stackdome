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

type StackStoreSpec struct {
	SessionFactory db.SessionFactory
}

type stackStore struct {
	sessionFactory       db.SessionFactory
	stackResourceStore   stores.StackResourceStore
	stackVolumeStore     stores.StackVolumeStore
	stackConnectionStore stores.StackConnectionStore
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
		stackConnectionStore: NewStackConnectionStore(StackConnectionStoreSpec{
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
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.Conflict("stack with name '%s' already exists", spec.Name)
		}
		return nil, errors.GeneralError("failed to create stack: %s", err.Error())
	}
	if err := w.stackConnectionStore.ReplaceByStackIDWithTx(ctx, spec.ID, spec.Connections); err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to create stack connections: %v", err)
	}
	tx.Commit()
	return w.GetByID(ctx, spec.ID)
}

func (w *stackStore) UpdateRevision(ctx context.Context, id string, revision string) *errors.ServiceError {
	tx := w.sessionFactory.New(ctx)
	if err := tx.Model(&models.Stack{}).Where("id = ?", id).UpdateColumn("cr_revision", revision).Error; err != nil {
		return errors.GeneralError("failed to update stack revision: %s", err.Error())
	}
	return nil
}

func (w *stackStore) InternalList(ctx context.Context, query string, args ...any) ([]*models.Stack, *errors.ServiceError) {
	var stacks []*models.Stack
	if err := w.sessionFactory.New(ctx).Where(query, args...).Omit(clause.Associations).Find(&stacks).Error; err != nil {
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
		stackConnections, err := w.stackConnectionStore.ListByStackID(ctx, stack.ID)
		if err != nil {
			return nil, errors.GeneralError("failed to get stack connections: %v", err)
		}
		stack.Connections = stackConnections
	}
	return stacks, nil
}

func (w *stackStore) CreateWithTx(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}

	if err := tx.Model(&models.Stack{}).Omit(clause.Associations).Create(spec).Error; err != nil {
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.Conflict("stack with name '%s' already exists", spec.Name)
		}
		return nil, errors.GeneralError("failed to create stack: %s", err.Error())
	}
	if err := w.stackConnectionStore.ReplaceByStackIDWithTx(ctx, spec.ID, spec.Connections); err != nil {
		return nil, errors.GeneralError("failed to create stack connections: %v", err)
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
		stackConnections, err := w.stackConnectionStore.ListByStackID(ctx, stack.ID)
		if err != nil {
			return nil, errors.GeneralError("failed to get stack connections: %v", err)
		}
		stack.Connections = stackConnections
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
		stackConnections, err := w.stackConnectionStore.ListByStackID(ctx, stack.ID)
		if err != nil {
			return nil, errors.GeneralError("failed to get stack connections: %v", err)
		}
		stack.Connections = stackConnections
	}
	return stacks, nil
}

func (w *stackStore) ListByProjectID(ctx context.Context, projectID string) ([]*models.Stack, *errors.ServiceError) {
	var stacks []*models.Stack
	if err := w.sessionFactory.New(ctx).Omit(clause.Associations).
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Find(&stacks).Error; err != nil {
		return nil, errors.GeneralError("failed to list stacks by project: %s", err.Error())
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
		stackConnections, err := w.stackConnectionStore.ListByStackID(ctx, stack.ID)
		if err != nil {
			return nil, errors.GeneralError("failed to get stack connections: %v", err)
		}
		stack.Connections = stackConnections
	}
	return stacks, nil
}

func (w *stackStore) ListByProjectIDs(ctx context.Context, projectIDs []string) ([]*models.Stack, *errors.ServiceError) {
	if len(projectIDs) == 0 {
		return []*models.Stack{}, nil
	}
	var stacks []*models.Stack
	if err := w.sessionFactory.New(ctx).Omit(clause.Associations).
		Where("project_id IN ?", projectIDs).
		Order("created_at DESC").
		Find(&stacks).Error; err != nil {
		return nil, errors.GeneralError("failed to list stacks by projects: %s", err.Error())
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
		stackConnections, err := w.stackConnectionStore.ListByStackID(ctx, stack.ID)
		if err != nil {
			return nil, errors.GeneralError("failed to get stack connections: %v", err)
		}
		stack.Connections = stackConnections
	}
	return stacks, nil
}

func (w *stackStore) GetByID(ctx context.Context, id string) (*models.Stack, *errors.ServiceError) {
	var stack models.Stack
	if err := w.sessionFactory.New(ctx).Omit(clause.Associations).First(&stack, "id = ?", id).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
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
	stackConnections, err := w.stackConnectionStore.ListByStackID(ctx, id)
	if err != nil {
		return nil, errors.GeneralError("failed to get stack connections: %v", err)
	}
	stack.Connections = stackConnections
	return &stack, nil
}

func (w *stackStore) LockByID(ctx context.Context, id string) *errors.ServiceError {
	var stack models.Stack
	if err := w.sessionFactory.New(ctx).
		Clauses(clause.Locking{Strength: rowLockStrengthUpdate}).
		Select("id").
		First(&stack, "id = ?", id).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFound("stack with id %s not found", id)
		}
		return errors.GeneralError("failed to lock stack: %s", err.Error())
	}
	return nil
}

func (w *stackStore) GetByNameAndProjectID(ctx context.Context, name string, projectID string) (*models.Stack, *errors.ServiceError) {
	var stack models.Stack
	if err := w.sessionFactory.New(ctx).Omit(clause.Associations).First(&stack, "name = ? AND project_id = ?", name, projectID).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
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
	stackConnections, err := w.stackConnectionStore.ListByStackID(ctx, stack.ID)
	if err != nil {
		return nil, errors.GeneralError("failed to get stack connections: %v", err)
	}
	stack.Connections = stackConnections
	return &stack, nil
}

func (w *stackStore) Update(ctx context.Context, id string, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	_, err := w.GetByID(ctx, id)
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
	if err := w.stackConnectionStore.ReplaceByStackIDWithTx(txCtx, id, spec.Connections); err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to update stack connections: %v", err)
	}
	tx.Commit()
	return w.GetByID(ctx, id)
}

func (w *stackStore) UpdateForDelete(ctx context.Context, id string, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	_, err := w.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := w.sessionFactory.New(ctx).Model(
		&models.Stack{}).Omit(clause.Associations).Where("id = ?", id).Updates(spec).Error; err != nil {
		return nil, errors.GeneralError("failed to update stack: %s", err.Error())
	}
	return w.GetByID(ctx, id)
}

func (w *stackStore) UpdateWithTx(ctx context.Context, id string, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	_, err := w.GetByID(ctx, id)
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
	if err := w.stackConnectionStore.ReplaceByStackIDWithTx(ctx, id, spec.Connections); err != nil {
		return nil, errors.GeneralError("failed to update stack connections: %v", err)
	}
	return w.GetByID(ctx, id)
}

func (w *stackStore) UpdateShellWithTx(ctx context.Context, id string, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	_, err := w.GetByID(ctx, id)
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
	return w.GetByID(ctx, id)
}

func (w *stackStore) UpdateStatus(ctx context.Context, id string, status *models.StackStatus) *errors.ServiceError {
	if err := w.sessionFactory.New(ctx).Model(&models.Stack{}).Where("id = ?", id).UpdateColumn("status", status).Error; err != nil {
		return errors.GeneralError("failed to update stack status: %s", err.Error())
	}
	return nil
}

func (w *stackStore) CreateConnectionWithTx(ctx context.Context, id string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError) {
	return w.stackConnectionStore.CreateWithTx(ctx, id, connection)
}

func (w *stackStore) UpdateConnectionWithTx(ctx context.Context, id string, connectionID string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError) {
	return w.stackConnectionStore.UpdateWithTx(ctx, id, connectionID, connection)
}

func (w *stackStore) DeleteConnectionWithTx(ctx context.Context, id string, connectionID string) *errors.ServiceError {
	return w.stackConnectionStore.DeleteWithTx(ctx, id, connectionID)
}

func (w *stackStore) UpdateConnectionsWithTx(ctx context.Context, id string, connections models.StackConnections) *errors.ServiceError {
	return w.stackConnectionStore.ReplaceByStackIDWithTx(ctx, id, connections)
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
