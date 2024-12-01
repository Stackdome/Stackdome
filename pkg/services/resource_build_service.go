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

type ResourceBuildService interface {
	Create(ctx context.Context, resourceBuild *models.WorkspaceResourceBuild) (*models.WorkspaceResourceBuild, *errors.ServiceError)
	UpdateStatus(ctx context.Context, BuildID string, status *models.WorkspaceResourceBuildStatus) *errors.ServiceError
	ListByResourceName(ctx context.Context, workspaceID string, resourceName string) ([]*models.WorkspaceResourceBuild, *errors.ServiceError)
	ListByWorkspaceID(ctx context.Context, workspaceID string) ([]*models.WorkspaceResourceBuild, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.WorkspaceResourceBuild, *errors.ServiceError)
}

type ResourceBuildServiceSpec struct {
	SessionFactory           db.SessionFactory
	WorkspaceResourceService WorkspaceResourceService
	Logger                   logger.Logger
}

type resourceBuildService struct {
	resourceBuildStore       stores.ResourceBuildStore
	workspaceResourceService WorkspaceResourceService
	logger                   logger.Logger
}

func NewResourceBuildService(spec ResourceBuildServiceSpec) ResourceBuildService {
	return &resourceBuildService{
		resourceBuildStore: pgstore.NewResourceBuildStore(pgstore.ResourceBuildStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		workspaceResourceService: spec.WorkspaceResourceService,
		logger:                   spec.Logger,
	}
}

func (r *resourceBuildService) Create(ctx context.Context, resourceBuild *models.WorkspaceResourceBuild) (*models.WorkspaceResourceBuild, *errors.ServiceError) {
	return r.resourceBuildStore.Create(ctx, resourceBuild)
}

func (r *resourceBuildService) UpdateStatus(ctx context.Context, BuildID string, status *models.WorkspaceResourceBuildStatus) *errors.ServiceError {
	return r.resourceBuildStore.UpdateStatus(ctx, BuildID, status)
}

func (r *resourceBuildService) ListByResourceName(ctx context.Context, workspaceID string, resourceName string) ([]*models.WorkspaceResourceBuild, *errors.ServiceError) {
	resource, err := r.workspaceResourceService.GetByWorkspaceIDAndResourceName(ctx, workspaceID, resourceName)
	if err != nil {
		return nil, err
	}
	return r.resourceBuildStore.GetByResourceID(ctx, resource.ID)
}

func (r *resourceBuildService) ListByWorkspaceID(ctx context.Context, workspaceID string) ([]*models.WorkspaceResourceBuild, *errors.ServiceError) {
	return r.resourceBuildStore.ListByWorkspaceID(ctx, workspaceID)
}

func (r *resourceBuildService) GetByID(ctx context.Context, ID string) (*models.WorkspaceResourceBuild, *errors.ServiceError) {
	return r.resourceBuildStore.GetByID(ctx, ID)
}
