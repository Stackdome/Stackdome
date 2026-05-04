package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
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
	UpdateStatus(ctx context.Context, resourceID string, status *models.StackResourceStatus) *errors.ServiceError
	InternalUpdateExposedPortDomainsWithTx(ctx context.Context, resourceID string, stackResource *models.StackResource) *errors.ServiceError
}

type StackResourceServiceSpec struct {
	SessionFactory       db.SessionFactory
	WorkspaceUserService WorkspaceUserService
	StorageService       StackStorageService
	Logger               logger.Logger
	Permissions          auth.PermissionService
	StackStore           stores.StackStore
}

type stackResourceService struct {
	stackResourceStore   stores.StackResourceStore
	stackStore           stores.StackStore
	logger               logger.Logger
	sessionFactory       db.SessionFactory
	workspaceUserService WorkspaceUserService
	storageService       StackStorageService
	permissions          auth.PermissionService
}

func NewStackResourceService(spec StackResourceServiceSpec) StackResourceService {
	return &stackResourceService{
		stackResourceStore: pgstore.NewStackResourceStore(pgstore.StackResourceStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		stackStore:           spec.StackStore,
		workspaceUserService: spec.WorkspaceUserService,
		storageService:       spec.StorageService,
		logger:               spec.Logger,
		sessionFactory:       spec.SessionFactory,
		permissions:          spec.Permissions,
	}
}

func (s *stackResourceService) Create(ctx context.Context, resource *models.StackResource) (*models.StackResource, *errors.ServiceError) {
	return s.stackResourceStore.Create(ctx, resource)
}

func (s *stackResourceService) GetByStackID(ctx context.Context, stackID string) ([]*models.StackResource, *errors.ServiceError) {
	stack, stackErr := s.stackStore.GetByID(ctx, stackID)
	if stackErr != nil {
		return nil, stackErr
	}
	if permErr := auth.CheckServicePermission(s.permissions, ctx, stack.TeamID, auth.ResourceStacks, stackID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return s.stackResourceStore.GetByStackID(ctx, stackID)
}

func (s *stackResourceService) GetByID(ctx context.Context, ID string) (*models.StackResource, *errors.ServiceError) {
	resource, err := s.stackResourceStore.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	stack, stackErr := s.stackStore.GetByID(ctx, resource.StackID)
	if stackErr != nil {
		return nil, stackErr
	}
	if permErr := auth.CheckServicePermission(s.permissions, ctx, stack.TeamID, auth.ResourceStacks, resource.StackID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return resource, nil
}

func (s *stackResourceService) GetByStackIDAndResourceName(ctx context.Context, stackID, resourceName string) (*models.StackResource, *errors.ServiceError) {
	stack, stackErr := s.stackStore.GetByID(ctx, stackID)
	if stackErr != nil {
		return nil, stackErr
	}
	if permErr := auth.CheckServicePermission(s.permissions, ctx, stack.TeamID, auth.ResourceStacks, stackID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return s.stackResourceStore.GetByStackIDAndResourceName(ctx, stackID, resourceName)
}

func (s *stackResourceService) UpdateStatus(ctx context.Context, resourceID string, status *models.StackResourceStatus) *errors.ServiceError {
	return s.stackResourceStore.UpdateStatus(ctx, resourceID, status)
}

func (s *stackResourceService) InternalUpdateExposedPortDomainsWithTx(ctx context.Context, resourceID string, stackResource *models.StackResource) *errors.ServiceError {
	for _, port := range stackResource.Ports {
		if port.ExposedToPublic {
			if port.ExposedFqdn == "" {
				return errors.GeneralError("port exposed to public but fqdn is empty")
			}
			if err := s.stackResourceStore.UpdatePortsWithTx(ctx, resourceID, stackResource); err != nil {
				return err
			}
		}
	}
	return nil
}
