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

type ResourceBuildStoreSpec struct {
	SessionFactory db.SessionFactory
}

type resourceBuildStore struct {
	sessionFactory db.SessionFactory
	atomicExecutor
}

func NewResourceBuildStore(spec ResourceBuildStoreSpec) stores.ResourceBuildStore {
	return &resourceBuildStore{
		sessionFactory: spec.SessionFactory,
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
	}
}

func (r *resourceBuildStore) Create(ctx context.Context, resourceBuild *models.WorkspaceResourceBuild) (*models.WorkspaceResourceBuild, *errors.ServiceError) {
	tx := r.sessionFactory.New(ctx).Begin()
	if err := tx.Model(&models.WorkspaceResourceBuild{}).Omit(clause.Associations).Create(resourceBuild).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to create resource build: %s", err.Error())
	}
	tx.Commit()
	return r.GetByID(ctx, resourceBuild.ID)
}

func (r *resourceBuildStore) UpdateStatus(ctx context.Context, BuildID string, status *models.WorkspaceResourceBuildStatus) *errors.ServiceError {
	tx := r.sessionFactory.New(ctx).Begin()
	if err := tx.Model(&models.WorkspaceResourceBuild{}).Where("id = ?", BuildID).UpdateColumn("status", status).Error; err != nil {
		tx.Rollback()
		return errors.GeneralError("failed to update resource build status: %s", err.Error())
	}
	tx.Commit()
	return nil
}

func (r *resourceBuildStore) GetByResourceID(ctx context.Context, resourceID string) ([]*models.WorkspaceResourceBuild, *errors.ServiceError) {
	var resourceBuilds []*models.WorkspaceResourceBuild
	if err := r.sessionFactory.New(ctx).Where("workspace_resource_id = ?", resourceID).Find(&resourceBuilds).Error; err != nil {
		return nil, errors.GeneralError("failed to get resource builds: %s", err.Error())
	}
	return resourceBuilds, nil
}

func (r *resourceBuildStore) GetByID(ctx context.Context, ID string) (*models.WorkspaceResourceBuild, *errors.ServiceError) {
	var resourceBuild models.WorkspaceResourceBuild
	if err := r.sessionFactory.New(ctx).Where("id = ?", ID).Preload(clause.Associations).First(&resourceBuild).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("resource build with id '%s' not found", ID)
		}
		return nil, errors.GeneralError("failed to get resource build: %s", err.Error())
	}
	return &resourceBuild, nil
}

func (r *resourceBuildStore) ListByWorkspaceID(ctx context.Context, workspaceID string) ([]*models.WorkspaceResourceBuild, *errors.ServiceError) {
	var resourceBuilds []*models.WorkspaceResourceBuild
	if err := r.sessionFactory.New(ctx).Where("workspace_id = ?", workspaceID).Find(&resourceBuilds).Error; err != nil {
		return nil, errors.GeneralError("failed to get resource builds: %s", err.Error())
	}
	return resourceBuilds, nil
}
