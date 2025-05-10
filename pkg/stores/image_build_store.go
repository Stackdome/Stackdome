package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type ImageBuildStore interface {
	Create(ctx context.Context, resourceBuild *models.ImageBuild) (*models.ImageBuild, *errors.ServiceError)
	UpdateStatus(ctx context.Context, BuildID string, status *models.ImageBuildStatus) *errors.ServiceError
	GetByResourceID(ctx context.Context, resourceID string) ([]*models.ImageBuild, *errors.ServiceError)
	ListByStackID(ctx context.Context, stackID string) ([]*models.ImageBuild, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.ImageBuild, *errors.ServiceError)
}
