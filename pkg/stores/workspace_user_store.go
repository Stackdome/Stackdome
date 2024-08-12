package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type WorkspaceUserStore interface {
	Create(context.Context, *models.WorkspaceUser) (*models.WorkspaceUser, *errors.ServiceError)
	CreateWithTx(context.Context, *models.WorkspaceUser) (*models.WorkspaceUser, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.WorkspaceUser, *errors.ServiceError)
	InternalList(ctx context.Context, query string, args ...any) ([]*models.WorkspaceUser, *errors.ServiceError)
	GetByUserID(ctx context.Context, userID string) (*models.WorkspaceUser, *errors.ServiceError)
	ListByOrgID(ctx context.Context, userID string) ([]*models.WorkspaceUser, *errors.ServiceError)
	Update(ctx context.Context, id string, spec *models.WorkspaceUser) (*models.WorkspaceUser, *errors.ServiceError)
	UpdateWithTx(ctx context.Context, id string, spec *models.WorkspaceUser) (*models.WorkspaceUser, *errors.ServiceError)
	PatchStatus(ctx context.Context, id string, spec *models.WorkspaceUserStatus) (*models.WorkspaceUser, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
	DeleteWithTx(ctx context.Context, id string) *errors.ServiceError
	WithTransaction(ctx context.Context, fn func(ctx context.Context) *errors.ServiceError) *errors.ServiceError
}
