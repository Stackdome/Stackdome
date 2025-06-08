package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type StackResourceStore interface {
	Create(ctx context.Context, resource *models.StackResource) (*models.StackResource, *errors.ServiceError)
	CreateWithTx(ctx context.Context, resource *models.StackResource, stack *models.Stack) (*models.StackResource, *errors.ServiceError)
	GetByStackID(ctx context.Context, stackID string) ([]*models.StackResource, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.StackResource, *errors.ServiceError)
	GetByStackIDAndResourceName(ctx context.Context, stackID, resourceName string) (*models.StackResource, *errors.ServiceError)
	Update(ctx context.Context, resourceID string, resource *models.StackResource, stack *models.Stack) (*models.StackResource, *errors.ServiceError)
	UpdateWithTx(ctx context.Context, resourceID string, resource *models.StackResource, stack *models.Stack) (*models.StackResource, *errors.ServiceError)
	UpdatePortsWithTx(ctx context.Context, resourceID string, resource *models.StackResource) *errors.ServiceError
	UpdateStatus(ctx context.Context, resourceID string, status *models.StackResourceStatus) *errors.ServiceError
	DeleteWithTx(ctx context.Context, ID string) *errors.ServiceError
	Delete(ctx context.Context, ID string) *errors.ServiceError
}
