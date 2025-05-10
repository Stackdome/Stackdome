package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

type ImageBuildService interface {
	Create(ctx context.Context, resourceBuild *models.ImageBuild) (*models.ImageBuild, *errors.ServiceError)
	UpdateStatus(ctx context.Context, BuildID string, status *models.ImageBuildStatus) *errors.ServiceError
	ListByResourceName(ctx context.Context, stackID string, resourceName string) ([]*models.ImageBuild, *errors.ServiceError)
	ListByStackID(ctx context.Context, stackID string) ([]*models.ImageBuild, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.ImageBuild, *errors.ServiceError)
}

type ImageBuildServiceSpec struct {
	SessionFactory       db.SessionFactory
	StackResourceService StackResourceService
	Logger               logger.Logger
}

type imageBuildService struct {
	imageBuildStore      stores.ImageBuildStore
	stackResourceService StackResourceService
	logger               logger.Logger
}

func NewImageBuildService(spec ImageBuildServiceSpec) ImageBuildService {
	return &imageBuildService{
		imageBuildStore: pgstore.NewImageBuildStore(pgstore.ImageBuildStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		stackResourceService: spec.StackResourceService,
		logger:               spec.Logger,
	}
}

func (r *imageBuildService) Create(ctx context.Context, imageBuild *models.ImageBuild) (*models.ImageBuild, *errors.ServiceError) {
	return r.imageBuildStore.Create(ctx, imageBuild)
}

func (r *imageBuildService) UpdateStatus(ctx context.Context, BuildID string, status *models.ImageBuildStatus) *errors.ServiceError {
	return r.imageBuildStore.UpdateStatus(ctx, BuildID, status)
}

func (r *imageBuildService) ListByResourceName(ctx context.Context, stackID string, resourceName string) ([]*models.ImageBuild, *errors.ServiceError) {
	resource, err := r.stackResourceService.GetByStackIDAndResourceName(ctx, stackID, resourceName)
	if err != nil {
		return nil, err
	}
	return r.imageBuildStore.GetByResourceID(ctx, resource.ID)
}

func (r *imageBuildService) ListByStackID(ctx context.Context, stackID string) ([]*models.ImageBuild, *errors.ServiceError) {
	return r.imageBuildStore.ListByStackID(ctx, stackID)
}

func (r *imageBuildService) GetByID(ctx context.Context, ID string) (*models.ImageBuild, *errors.ServiceError) {
	return r.imageBuildStore.GetByID(ctx, ID)
}
