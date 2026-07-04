package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

type StorageStore interface {
	Create(ctx context.Context, spec *models.StackStorage) (*models.StackStorage, *errors.ServiceError)
	CreateWithTx(ctx context.Context, spec *models.StackStorage) (*models.StackStorage, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.StackStorage, *errors.ServiceError)
	GetByWorkspaceName(ctx context.Context, workspaceName string, userID string) (*models.StackStorage, *errors.ServiceError)
	GetByIDorName(ctx context.Context, idOrName string, userID string) (*models.StackStorage, *errors.ServiceError)
	Update(ctx context.Context, id string, spec *models.StackStorage) (*models.StackStorage, *errors.ServiceError)
	UpdateWithTx(ctx context.Context, id string, spec *models.StackStorage) (*models.StackStorage, *errors.ServiceError)
	UpdateStatus(ctx context.Context, id string, status *models.StackStorageStatus) (*models.StackStorage, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
	DeleteWithTx(ctx context.Context, id string) *errors.ServiceError
	DeleteByNameWithTx(ctx context.Context, name string, userID string) *errors.ServiceError
	ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.StackStorage, *errors.ServiceError)
	ListByUserID(ctx context.Context, userID string) ([]*models.StackStorage, *errors.ServiceError)
	AtomicExecutor
}
