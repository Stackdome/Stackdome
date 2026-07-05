package pgstore

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"gorm.io/gorm"
)

type dbWorkspaceNamespaceStore struct {
	sessionFactory db.SessionFactory
}

type WorkspaceNamespaceStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewWorkspaceNamespaceStore(spec WorkspaceNamespaceStoreSpec) stores.WorkspaceNamespaceStore {
	return &dbWorkspaceNamespaceStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (d dbWorkspaceNamespaceStore) Create(ctx context.Context, wn *models.WorkspaceNamespace) (*models.WorkspaceNamespace, *errors.ServiceError) {
	var grm *gorm.DB
	grm = db.TxFromContext(ctx)
	if grm == nil {
		grm = d.sessionFactory.New(ctx)
	}
	err := grm.Create(wn).Error
	if err != nil {
		return nil, errors.GeneralError("failed to create workspace namespace: %s", err.Error())
	}
	return d.GetByNamespace(ctx, wn.Namespace)
}

func (d dbWorkspaceNamespaceStore) ListByUserID(ctx context.Context, userID string) ([]*models.WorkspaceNamespace, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var workspaceNamespaces []*models.WorkspaceNamespace
	err := grm.Where("user_id = ?", userID).Find(&workspaceNamespaces).Error
	if err != nil {
		return nil, errors.GeneralError("failed to list workspace namespaces: %s", err.Error())
	}
	return workspaceNamespaces, nil
}

func (d dbWorkspaceNamespaceStore) GetByNamespace(ctx context.Context, namespace string) (*models.WorkspaceNamespace, *errors.ServiceError) {
	var grm *gorm.DB
	grm = db.TxFromContext(ctx)
	if grm == nil {
		grm = d.sessionFactory.New(ctx)
	}
	var wn models.WorkspaceNamespace
	err := grm.Where("namespace = ?", namespace).First(&wn).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("workspace namespace '%s' not found", namespace)
		}
		return nil, errors.GeneralError("failed to fetch workspace namespace: %s", err.Error())
	}
	return &wn, nil
}

func (d dbWorkspaceNamespaceStore) GetByWorkspaceName(ctx context.Context, workspaceName string, userID string) (*models.WorkspaceNamespace, *errors.ServiceError) {
	var grm *gorm.DB
	grm = db.TxFromContext(ctx)
	if grm == nil {
		grm = d.sessionFactory.New(ctx)
	}
	var wn models.WorkspaceNamespace
	err := grm.Where("workspace = ? AND user_id = ?", workspaceName, userID).First(&wn).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("workspace namespace not found for workspace '%s' and user '%s'", workspaceName, userID)
		}
		return nil, errors.GeneralError("failed to fetch workspace namespace: %s", err.Error())
	}
	return &wn, nil
}

func (d dbWorkspaceNamespaceStore) Delete(ctx context.Context, namespace string) *errors.ServiceError {
	grm := d.sessionFactory.New(ctx)
	err := grm.Where("namespace = ?", namespace).Delete(&models.WorkspaceNamespace{}).Error
	if err != nil {
		return errors.GeneralError("failed to delete workspace namespace: %s", err.Error())
	}
	return nil
}

func (d dbWorkspaceNamespaceStore) Update(ctx context.Context, workspaceName string, userID string, spec *models.WorkspaceNamespace) (*models.WorkspaceNamespace, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	err := grm.Model(&models.WorkspaceNamespace{}).Where("workspace = ? AND user_id = ?", workspaceName, userID).Updates(spec).Error
	if err != nil {
		return nil, errors.GeneralError("failed to update workspace namespace: %s", err.Error())
	}
	return d.GetByWorkspaceName(ctx, workspaceName, userID)
}

func (d dbWorkspaceNamespaceStore) UpdateWithTx(ctx context.Context, workspaceName string, userID string, spec *models.WorkspaceNamespace) (*models.WorkspaceNamespace, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	err := tx.Model(&models.WorkspaceNamespace{}).Where("workspace = ? AND user_id = ?", workspaceName, userID).Updates(
		map[string]interface{}{
			"namespace": spec.Namespace,
			"enabled":   spec.Enabled,
		},
	).Error
	if err != nil {
		return nil, errors.GeneralError("failed to update workspace namespace: %s", err.Error())
	}
	return d.GetByWorkspaceName(ctx, workspaceName, userID)
}

func (d dbWorkspaceNamespaceStore) CreateBatch(ctx context.Context, wns []*models.WorkspaceNamespace) ([]*models.WorkspaceNamespace, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	err := grm.Create(&wns).Error
	if err != nil {
		return nil, errors.GeneralError("failed to create workspace namespace: %s", err.Error())
	}
	return wns, nil
}

func (d dbWorkspaceNamespaceStore) CreateBatchWithTx(ctx context.Context, wns []*models.WorkspaceNamespace) ([]*models.WorkspaceNamespace, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	err := tx.Create(&wns).Error
	if err != nil {
		return nil, errors.GeneralError("failed to create workspace namespace: %s", err.Error())
	}
	return wns, nil
}
