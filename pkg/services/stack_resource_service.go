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

type StackResourceService interface {
	Create(ctx context.Context, resource *models.StackResource) (*models.StackResource, *errors.ServiceError)
	GetByStackID(ctx context.Context, stackID string) ([]*models.StackResource, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.StackResource, *errors.ServiceError)
	GetByStackIDAndResourceName(ctx context.Context, stackID, resourceName string) (*models.StackResource, *errors.ServiceError)
	Update(ctx context.Context, resourceID string, resource *models.StackResource) (*models.StackResource, *errors.ServiceError)
	InternalUpdateWithTx(ctx context.Context, resourceID string, resource *models.StackResource) (*models.StackResource, *errors.ServiceError)
	UpdateStatus(ctx context.Context, resourceID string, status *models.StackResourceStatus) *errors.ServiceError
	InternalUpdateExposedPortDomainsWithTx(ctx context.Context, resourceID string, stackResource *models.StackResource) *errors.ServiceError
}

type StackResourceServiceSpec struct {
	SessionFactory       db.SessionFactory
	WorkspaceUserService WorkspaceUserService
	StorageService       StackStorageService
	Logger               logger.Logger
}

type stackResourceService struct {
	stackResourceStore   stores.StackResourceStore
	logger               logger.Logger
	sessionFactory       db.SessionFactory
	workspaceUserService WorkspaceUserService
	storageService       StackStorageService
}

func NewStackResourceService(spec StackResourceServiceSpec) StackResourceService {
	return &stackResourceService{
		stackResourceStore: pgstore.NewStackResourceStore(pgstore.StackResourceStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		workspaceUserService: spec.WorkspaceUserService,
		storageService:       spec.StorageService,
		logger:               spec.Logger,
		sessionFactory:       spec.SessionFactory,
	}
}

func (s *stackResourceService) Create(ctx context.Context, resource *models.StackResource) (*models.StackResource, *errors.ServiceError) {
	return s.stackResourceStore.Create(ctx, resource)
}

func (s *stackResourceService) GetByStackID(ctx context.Context, stackID string) ([]*models.StackResource, *errors.ServiceError) {
	return s.stackResourceStore.GetByStackID(ctx, stackID)
}

func (s *stackResourceService) GetByID(ctx context.Context, ID string) (*models.StackResource, *errors.ServiceError) {
	return s.stackResourceStore.GetByID(ctx, ID)
}

func (s *stackResourceService) GetByStackIDAndResourceName(ctx context.Context, stackID, resourceName string) (*models.StackResource, *errors.ServiceError) {
	return s.stackResourceStore.GetByStackIDAndResourceName(ctx, stackID, resourceName)
}

func (s *stackResourceService) Update(ctx context.Context, resourceID string, resource *models.StackResource) (*models.StackResource, *errors.ServiceError) {
	return s.stackResourceStore.Update(ctx, resourceID, resource)
}

func (s *stackResourceService) UpdateStatus(ctx context.Context, resourceID string, status *models.StackResourceStatus) *errors.ServiceError {
	return s.stackResourceStore.UpdateStatus(ctx, resourceID, status)
}

func (s *stackResourceService) InternalUpdateWithTx(ctx context.Context, resourceID string, resource *models.StackResource) (*models.StackResource, *errors.ServiceError) {
	return s.stackResourceStore.UpdateWithTx(ctx, resourceID, resource)
}

func (s *stackResourceService) InternalUpdateExposedPortDomainsWithTx(ctx context.Context, resourceID string, stackResource *models.StackResource) *errors.ServiceError {
	for _, port := range stackResource.Ports {
		if port.ExposedToPublic {
			if port.ExposedFqdn == "" {
				return errors.GeneralError("port exposed to public but fqdn is empty")
			}
			if _, err := s.InternalUpdateWithTx(ctx, resourceID, stackResource); err != nil {
				return err
			}
		}
	}
	return nil
}
