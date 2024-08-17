package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type WorkspaceStorageStore interface {
	Create(ctx context.Context, spec *models.WorkspaceStorage) (*models.WorkspaceStorage, *errors.ServiceError)
	CreateWithTx(ctx context.Context, spec *models.WorkspaceStorage) (*models.WorkspaceStorage, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.WorkspaceStorage, *errors.ServiceError)
	GetByIDorName(ctx context.Context, idOrName string, userID string) (*models.WorkspaceStorage, *errors.ServiceError)
	Update(ctx context.Context, id string, spec *models.WorkspaceStorage) (*models.WorkspaceStorage, *errors.ServiceError)
	UpdateWithTx(ctx context.Context, id string, spec *models.WorkspaceStorage) (*models.WorkspaceStorage, *errors.ServiceError)
	UpdateStatus(ctx context.Context, id string, status *models.WorkspaceStorageStatus) (*models.WorkspaceStorage, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
	DeleteWithTx(ctx context.Context, id string) *errors.ServiceError
	DeleteByNameWithTx(ctx context.Context, name string, userID string) *errors.ServiceError
	ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.WorkspaceStorage, *errors.ServiceError)
	ListByUserID(ctx context.Context, userID string) ([]*models.WorkspaceStorage, *errors.ServiceError)
	AtomicExecutor
}
