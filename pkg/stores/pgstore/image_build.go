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

type ImageBuildStoreSpec struct {
	SessionFactory db.SessionFactory
}

type imageBuildStore struct {
	sessionFactory db.SessionFactory
	atomicExecutor
}

func NewImageBuildStore(spec ImageBuildStoreSpec) stores.ImageBuildStore {
	return &imageBuildStore{
		sessionFactory: spec.SessionFactory,
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
	}
}

func (r *imageBuildStore) Create(ctx context.Context, imageBuild *models.ImageBuild) (*models.ImageBuild, *errors.ServiceError) {
	tx := r.sessionFactory.New(ctx).Begin()
	if err := tx.Model(&models.ImageBuild{}).Omit(clause.Associations).Create(imageBuild).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to create resource build: %s", err.Error())
	}
	tx.Commit()
	return r.GetByID(ctx, imageBuild.ID)
}

func (r *imageBuildStore) UpdateStatus(ctx context.Context, BuildID string, status *models.ImageBuildStatus) *errors.ServiceError {
	tx := r.sessionFactory.New(ctx).Begin()
	if err := tx.Model(&models.ImageBuild{}).Where("id = ?", BuildID).UpdateColumn("status", status).Error; err != nil {
		tx.Rollback()
		return errors.GeneralError("failed to update resource build status: %s", err.Error())
	}
	tx.Commit()
	return nil
}

func (r *imageBuildStore) GetByResourceID(ctx context.Context, resourceID string) ([]*models.ImageBuild, *errors.ServiceError) {
	var imageBuilds []*models.ImageBuild
	if err := r.sessionFactory.New(ctx).Where("stack_resource_id = ?", resourceID).Find(&imageBuilds).Error; err != nil {
		return nil, errors.GeneralError("failed to get image builds: %s", err.Error())
	}
	return imageBuilds, nil
}

func (r *imageBuildStore) GetByID(ctx context.Context, ID string) (*models.ImageBuild, *errors.ServiceError) {
	var imageBuild models.ImageBuild
	if err := r.sessionFactory.New(ctx).Where("id = ?", ID).Preload(clause.Associations).First(&imageBuild).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("image build with id '%s' not found", ID)
		}
		return nil, errors.GeneralError("failed to get image build: %s", err.Error())
	}
	return &imageBuild, nil
}

func (r *imageBuildStore) ListByStackID(ctx context.Context, stackID string) ([]*models.ImageBuild, *errors.ServiceError) {
	var imageBuilds []*models.ImageBuild
	if err := r.sessionFactory.New(ctx).Where("stack_id = ?", stackID).Find(&imageBuilds).Error; err != nil {
		return nil, errors.GeneralError("failed to get image builds: %s", err.Error())
	}
	return imageBuilds, nil
}
