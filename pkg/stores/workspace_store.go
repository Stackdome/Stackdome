package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type WorkspaceStore interface {
	Create(ctx context.Context, spec *models.Workspace) (*models.Workspace, *errors.ServiceError)
	CreateWithTx(ctx context.Context, spec *models.Workspace) (*models.Workspace, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.Workspace, *errors.ServiceError)
	GetByName(ctx context.Context, Name string, userID string) (*models.Workspace, *errors.ServiceError)
	Update(ctx context.Context, id string, spec *models.Workspace) (*models.Workspace, *errors.ServiceError)
	UpdateWithTx(ctx context.Context, id string, spec *models.Workspace) (*models.Workspace, *errors.ServiceError)
	UpdateStatus(ctx context.Context, id string, status *models.WorkspaceStatus) *errors.ServiceError
	DeleteWithTx(ctx context.Context, id string) *errors.ServiceError
	Delete(ctx context.Context, id string) *errors.ServiceError
	ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.Workspace, *errors.ServiceError)
	ListByUserID(ctx context.Context, userID string) ([]*models.Workspace, *errors.ServiceError)
	AtomicExecutor
}
