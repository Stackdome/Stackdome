package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type ResourceBuildStore interface {
	Create(ctx context.Context, resourceBuild *models.WorkspaceResourceBuild) (*models.WorkspaceResourceBuild, *errors.ServiceError)
	UpdateStatus(ctx context.Context, BuildID string, status *models.WorkspaceResourceBuildStatus) *errors.ServiceError
	GetByResourceID(ctx context.Context, resourceID string) ([]*models.WorkspaceResourceBuild, *errors.ServiceError)
	ListByWorkspaceID(ctx context.Context, workspaceID string) ([]*models.WorkspaceResourceBuild, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.WorkspaceResourceBuild, *errors.ServiceError)
}
