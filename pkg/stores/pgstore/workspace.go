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

type WorkspaceStoreSpec struct {
	SessionFactory db.SessionFactory
}

type workspaceStore struct {
	sessionFactory         db.SessionFactory
	workspaceResourceStore stores.WorkspaceResourceStore
	atomicExecutor
}

func NewWorkspaceStore(spec *WorkspaceStoreSpec) stores.WorkspaceStore {
	return &workspaceStore{
		sessionFactory: spec.SessionFactory,
		workspaceResourceStore: NewWorkspaceResourceStore(WorkspaceResourceStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
	}
}

func (w *workspaceStore) Create(ctx context.Context, spec *models.Workspace) (*models.Workspace, *errors.ServiceError) {
	tx := w.sessionFactory.New(ctx).Begin()
	txCtx := db.CtxWithTransaction(ctx, tx)
	if err := tx.Model(&models.Workspace{}).Omit(clause.Associations).Create(spec).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to create workspace: %s", err.Error())
	}
	for _, resource := range spec.WorkspaceResources {
		resource.WorkspaceID = spec.ID
		resource.UserID = spec.UserID
		if _, err := w.workspaceResourceStore.CreateWithTx(txCtx, resource); err != nil {
			tx.Rollback()
			return nil, errors.GeneralError("failed to create workpace: errored creating workspace resource '%s': %v", resource.Name, err)
		}
	}
	tx.Commit()
	return w.GetByID(ctx, spec.ID)
}

func (w *workspaceStore) CreateWithTx(ctx context.Context, spec *models.Workspace) (*models.Workspace, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}

	if err := tx.Model(&models.Workspace{}).Omit(clause.Associations).Create(spec).Error; err != nil {
		return nil, errors.GeneralError("failed to create workspace: %s", err.Error())
	}
	for _, resource := range spec.WorkspaceResources {
		resource.WorkspaceID = spec.ID
		if _, err := w.workspaceResourceStore.CreateWithTx(ctx, resource); err != nil {
			return nil, errors.GeneralError("failed to create workspace. Error create workspace resource: %v", err)
		}
	}
	return w.GetByID(ctx, spec.ID)
}

func (w *workspaceStore) ListByUserID(ctx context.Context, userID string) ([]*models.Workspace, *errors.ServiceError) {
	var workspaces []*models.Workspace
	if err := w.sessionFactory.New(ctx).Omit(clause.Associations).Find(&workspaces, "user_id = ?", userID).Error; err != nil {
		return nil, errors.GeneralError("failed to list workspaces: %s", err.Error())
	}
	for _, workspace := range workspaces {
		resources, err := w.workspaceResourceStore.GetByWorkspaceID(ctx, workspace.ID)
		if err != nil {
			return nil, errors.GeneralError("failed to get workspace resources: %v", err)
		}
		workspace.WorkspaceResources = resources
	}
	return workspaces, nil
}

func (w *workspaceStore) ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.Workspace, *errors.ServiceError) {
	var workspaces []*models.Workspace
	if err := w.sessionFactory.New(ctx).Omit(clause.Associations).Find(&workspaces, "organisation_id = ?", organisationID).Error; err != nil {
		return nil, errors.GeneralError("failed to list workspaces: %s", err.Error())
	}
	for _, workspace := range workspaces {
		resources, err := w.workspaceResourceStore.GetByWorkspaceID(ctx, workspace.ID)
		if err != nil {
			return nil, errors.GeneralError("failed to get workspace resources: %v", err)
		}
		workspace.WorkspaceResources = resources
	}
	return workspaces, nil
}

func (w *workspaceStore) GetByID(ctx context.Context, id string) (*models.Workspace, *errors.ServiceError) {
	var workspace models.Workspace
	if err := w.sessionFactory.New(ctx).Omit(clause.Associations).First(&workspace, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("workspace with id %s not found", id)
		}
		return nil, errors.GeneralError("failed to get workspace: %s", err.Error())
	}
	resources, err := w.workspaceResourceStore.GetByWorkspaceID(ctx, id)
	if err != nil {
		return nil, errors.GeneralError("failed to get workspace resources: %v", err)
	}
	workspace.WorkspaceResources = resources
	return &workspace, nil
}

func (w *workspaceStore) GetByName(ctx context.Context, Name string, userID string) (*models.Workspace, *errors.ServiceError) {
	var workspace models.Workspace
	if err := w.sessionFactory.New(ctx).Omit(clause.Associations).First(&workspace, "name = ? AND user_id = ?", Name, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("workspace with name %s not found", Name)
		}
		return nil, errors.GeneralError("failed to get workspace: %s", err.Error())
	}
	resources, err := w.workspaceResourceStore.GetByWorkspaceID(ctx, workspace.ID)
	if err != nil {
		return nil, errors.GeneralError("failed to get workspace resources: %v", err)
	}
	workspace.WorkspaceResources = resources
	return &workspace, nil
}

func (w *workspaceStore) Update(ctx context.Context, id string, spec *models.Workspace) (*models.Workspace, *errors.ServiceError) {
	currentWorkspace, err := w.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tx := w.sessionFactory.New(ctx).Begin()
	txCtx := db.CtxWithTransaction(ctx, tx)
	spec.Status = nil
	if err := tx.Model(&models.Workspace{}).Omit(clause.Associations).Where("id = ?", id).Updates(spec).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to update workspace: %s", err.Error())
	}
	currentResourceMap := currentWorkspace.ResourcesMap()
	for _, resource := range spec.WorkspaceResources {
		resource.WorkspaceID = currentWorkspace.ID
		resource.UserID = spec.UserID
		if currentResource, ok := currentResourceMap[resource.Name]; ok {
			if _, err := w.workspaceResourceStore.UpdateWithTx(txCtx, currentResource.ID, resource); err != nil {
				tx.Rollback()
				return nil, errors.GeneralError("failed to update workspace resource: %v", err)
			}
		} else {
			if _, err := w.workspaceResourceStore.CreateWithTx(txCtx, resource); err != nil {
				tx.Rollback()
				return nil, errors.GeneralError("failed to create workspace resource: %v", err)
			}
		}
	}

	specResourceMap := spec.ResourcesMap()

	// Delete resources that are not in the new spec
	for _, resource := range currentWorkspace.WorkspaceResources {
		if _, ok := specResourceMap[resource.Name]; !ok {
			if err := w.workspaceResourceStore.DeleteWithTx(txCtx, resource.ID); err != nil {
				tx.Rollback()
				return nil, errors.GeneralError("failed to update workspace. error deleting workspace resource: %v", err)
			}
		}
	}
	tx.Commit()
	return w.GetByID(ctx, id)
}

func (w *workspaceStore) UpdateWithTx(ctx context.Context, id string, spec *models.Workspace) (*models.Workspace, *errors.ServiceError) {
	currentWorkspace, err := w.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	spec.Status = nil
	if err := tx.Model(&models.Workspace{}).Omit(clause.Associations).Where("id = ?", id).Updates(spec).Error; err != nil {
		return nil, errors.GeneralError("failed to update workspace: %s", err.Error())
	}
	currentResourceMap := currentWorkspace.ResourcesMap()
	for _, resource := range spec.WorkspaceResources {
		resource.WorkspaceID = spec.ID
		resource.UserID = spec.UserID
		if currentResource, ok := currentResourceMap[resource.Name]; ok {
			if _, err := w.workspaceResourceStore.UpdateWithTx(ctx, currentResource.ID, resource); err != nil {
				return nil, errors.GeneralError("failed to update workspace resource: %v", err)
			}
		} else {
			if _, err := w.workspaceResourceStore.CreateWithTx(ctx, resource); err != nil {
				return nil, errors.GeneralError("failed to create workspace resource: %v", err)
			}
		}
	}

	specResourceMap := spec.ResourcesMap()

	// Delete resources that are not in the new spec
	for _, resource := range currentWorkspace.WorkspaceResources {
		if _, ok := specResourceMap[resource.Name]; !ok {
			if err := w.workspaceResourceStore.DeleteWithTx(ctx, resource.ID); err != nil {
				return nil, errors.GeneralError("failed to update workspace. error deleting workspace resource: %v", err)
			}
		}
	}
	return w.GetByID(ctx, id)
}

func (w *workspaceStore) UpdateStatus(ctx context.Context, id string, status *models.WorkspaceStatus) *errors.ServiceError {
	if err := w.sessionFactory.New(ctx).Model(&models.Workspace{}).Where("id = ?", id).UpdateColumn("status", status).Error; err != nil {
		return errors.GeneralError("failed to update workspace status: %s", err.Error())
	}
	return nil
}

func (w *workspaceStore) DeleteWithTx(ctx context.Context, id string) *errors.ServiceError {
	tx := w.sessionFactory.New(ctx)
	if err := tx.Delete(&models.Workspace{}, "id = ?", id).Error; err != nil {
		return errors.GeneralError("failed to delete workspace: %s", err.Error())
	}
	return nil
}

func (w *workspaceStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	tx := w.sessionFactory.New(ctx)
	if err := tx.Delete(&models.Workspace{}, "id = ?", id).Error; err != nil {
		return errors.GeneralError("failed to delete workspace: %s", err.Error())
	}
	return nil
}
