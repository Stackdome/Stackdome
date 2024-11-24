package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type WorkspaceResourceStore interface {
	Create(ctx context.Context, resource *models.WorkspaceResource) (*models.WorkspaceResource, *errors.ServiceError)
	CreateWithTx(ctx context.Context, resource *models.WorkspaceResource) (*models.WorkspaceResource, *errors.ServiceError)
	GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*models.WorkspaceResource, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.WorkspaceResource, *errors.ServiceError)
	GetByWorkspaceIDAndResourceName(ctx context.Context, workspaceID, resourceName string) (*models.WorkspaceResource, *errors.ServiceError)
	Update(ctx context.Context, resourceID string, resource *models.WorkspaceResource) (*models.WorkspaceResource, *errors.ServiceError)
	UpdateWithTx(ctx context.Context, resourceID string, resource *models.WorkspaceResource) (*models.WorkspaceResource, *errors.ServiceError)
	UpdateStatus(ctx context.Context, resourceID string, status *models.WorkspaceResourceStatus) *errors.ServiceError
	DeleteWithTx(ctx context.Context, ID string) *errors.ServiceError
	Delete(ctx context.Context, ID string) *errors.ServiceError
}
