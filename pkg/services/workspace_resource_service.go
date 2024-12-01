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

type WorkspaceResourceService interface {
	Create(ctx context.Context, resource *models.WorkspaceResource) (*models.WorkspaceResource, *errors.ServiceError)
	GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*models.WorkspaceResource, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.WorkspaceResource, *errors.ServiceError)
	GetByWorkspaceIDAndResourceName(ctx context.Context, workspaceID, resourceName string) (*models.WorkspaceResource, *errors.ServiceError)
	Update(ctx context.Context, resourceID string, resource *models.WorkspaceResource) (*models.WorkspaceResource, *errors.ServiceError)
	UpdateStatus(ctx context.Context, resourceID string, status *models.WorkspaceResourceStatus) *errors.ServiceError
}

type WorkspaceResourceServiceSpec struct {
	SessionFactory          db.SessionFactory
	WorkspaceUserService    WorkspaceUserService
	WorkspaceStorageService WorkspaceStorageService
	WorkspaceService        WorkspaceService
	Logger                  logger.Logger
}

type workspaceResourceService struct {
	workspaceResourceStore  stores.WorkspaceResourceStore
	logger                  logger.Logger
	sessionFactory          db.SessionFactory
	workspaceUserService    WorkspaceUserService
	workspaceStorageService WorkspaceStorageService
	workspaceService        WorkspaceService
}

func NewWorkspaceResourceService(spec WorkspaceResourceServiceSpec) WorkspaceResourceService {
	return &workspaceResourceService{
		workspaceResourceStore: pgstore.NewWorkspaceResourceStore(pgstore.WorkspaceResourceStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		workspaceUserService:    spec.WorkspaceUserService,
		workspaceStorageService: spec.WorkspaceStorageService,
		workspaceService:        spec.WorkspaceService,
		logger:                  spec.Logger,
		sessionFactory:          spec.SessionFactory,
	}
}

func (s *workspaceResourceService) Create(ctx context.Context, resource *models.WorkspaceResource) (*models.WorkspaceResource, *errors.ServiceError) {
	return s.workspaceResourceStore.Create(ctx, resource)
}

func (s *workspaceResourceService) GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*models.WorkspaceResource, *errors.ServiceError) {
	return s.workspaceResourceStore.GetByWorkspaceID(ctx, workspaceID)
}

func (s *workspaceResourceService) GetByID(ctx context.Context, ID string) (*models.WorkspaceResource, *errors.ServiceError) {
	return s.workspaceResourceStore.GetByID(ctx, ID)
}

func (s *workspaceResourceService) GetByWorkspaceIDAndResourceName(ctx context.Context, workspaceID, resourceName string) (*models.WorkspaceResource, *errors.ServiceError) {
	return s.workspaceResourceStore.GetByWorkspaceIDAndResourceName(ctx, workspaceID, resourceName)
}

func (s *workspaceResourceService) Update(ctx context.Context, resourceID string, resource *models.WorkspaceResource) (*models.WorkspaceResource, *errors.ServiceError) {
	return s.workspaceResourceStore.Update(ctx, resourceID, resource)
}

func (s *workspaceResourceService) UpdateStatus(ctx context.Context, resourceID string, status *models.WorkspaceResourceStatus) *errors.ServiceError {
	return s.workspaceResourceStore.UpdateStatus(ctx, resourceID, status)
}
