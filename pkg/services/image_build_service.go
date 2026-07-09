package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
)

//go:generate mockgen -destination=../mocks/mock_image_build_service.go -package=mocks github.com/Stackdome/stackdome/pkg/services ImageBuildService
type ImageBuildService interface {
	InternalCreate(ctx context.Context, resourceBuild *models.ImageBuild) (*models.ImageBuild, *errors.ServiceError)
	InternalUpdateStatus(ctx context.Context, BuildID string, status *models.ImageBuildStatus) *errors.ServiceError
	ListByResourceName(ctx context.Context, stackID string, resourceName string) ([]*models.ImageBuild, *errors.ServiceError)
	ListByStackID(ctx context.Context, stackID string) ([]*models.ImageBuild, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.ImageBuild, *errors.ServiceError)
	InternalGetByID(ctx context.Context, ID string) (*models.ImageBuild, *errors.ServiceError)
}

type ImageBuildServiceSpec struct {
	SessionFactory       db.SessionFactory
	StackResourceService StackResourceService
	Logger               logger.Logger
	Permissions          auth.PermissionService
	StackStore           stores.StackStore
}

type imageBuildService struct {
	imageBuildStore      stores.ImageBuildStore
	stackStore           stores.StackStore
	stackResourceService StackResourceService
	logger               logger.Logger
	permissions          auth.PermissionService
}

func NewImageBuildService(spec ImageBuildServiceSpec) ImageBuildService {
	return &imageBuildService{
		imageBuildStore: pgstore.NewImageBuildStore(pgstore.ImageBuildStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		stackStore:           spec.StackStore,
		stackResourceService: spec.StackResourceService,
		logger:               spec.Logger,
		permissions:          spec.Permissions,
	}
}

func (r *imageBuildService) InternalCreate(ctx context.Context, imageBuild *models.ImageBuild) (*models.ImageBuild, *errors.ServiceError) {
	return r.imageBuildStore.Create(ctx, imageBuild)
}

func (r *imageBuildService) InternalUpdateStatus(ctx context.Context, BuildID string, status *models.ImageBuildStatus) *errors.ServiceError {
	return r.imageBuildStore.UpdateStatus(ctx, BuildID, status)
}

func (r *imageBuildService) ListByResourceName(ctx context.Context, stackID string, resourceName string) ([]*models.ImageBuild, *errors.ServiceError) {
	stack, stackErr := r.stackStore.GetByID(ctx, stackID)
	if stackErr != nil {
		return nil, stackErr
	}
	if permErr := r.permissions.Check(ctx, stack.TeamID, auth.ResourceStacks, stackID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	resource, err := r.stackResourceService.GetByStackIDAndResourceName(ctx, stackID, resourceName)
	if err != nil {
		return nil, err
	}
	return r.imageBuildStore.GetByResourceID(ctx, resource.ID)
}

func (r *imageBuildService) ListByStackID(ctx context.Context, stackID string) ([]*models.ImageBuild, *errors.ServiceError) {
	stack, stackErr := r.stackStore.GetByID(ctx, stackID)
	if stackErr != nil {
		return nil, stackErr
	}
	if permErr := r.permissions.Check(ctx, stack.TeamID, auth.ResourceStacks, stackID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return r.imageBuildStore.ListByStackID(ctx, stackID)
}

func (r *imageBuildService) GetByID(ctx context.Context, ID string) (*models.ImageBuild, *errors.ServiceError) {
	build, err := r.imageBuildStore.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	stack, stackErr := r.stackStore.GetByID(ctx, build.StackID)
	if stackErr != nil {
		return nil, stackErr
	}
	if permErr := r.permissions.Check(ctx, stack.TeamID, auth.ResourceStacks, build.StackID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return build, nil
}

func (r *imageBuildService) InternalGetByID(ctx context.Context, ID string) (*models.ImageBuild, *errors.ServiceError) {
	return r.imageBuildStore.GetByID(ctx, ID)
}
