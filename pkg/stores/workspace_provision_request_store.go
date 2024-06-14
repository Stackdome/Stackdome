package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type WorkspaceProvisionRequestStore interface {
	Create(context.Context, *models.WorkspaceProvisionRequest) (*models.WorkspaceProvisionRequest, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.WorkspaceProvisionRequest, *errors.ServiceError)
	InternalList(ctx context.Context, query string, args ...any) ([]*models.WorkspaceProvisionRequest, *errors.ServiceError)
	GetByUserID(ctx context.Context, userID string) (*models.WorkspaceProvisionRequest, *errors.ServiceError)
	ListByOrgID(ctx context.Context, userID string) ([]*models.WorkspaceProvisionRequest, *errors.ServiceError)
	Update(ctx context.Context, id string, spec *models.WorkspaceProvisionRequest) (*models.WorkspaceProvisionRequest, *errors.ServiceError)
	PatchStatus(ctx context.Context, id string, spec *models.WorkspaceProvisionRequestStatus) (*models.WorkspaceProvisionRequest, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
}
