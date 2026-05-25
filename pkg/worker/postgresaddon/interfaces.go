package postgresaddon

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type postgresAddonService interface {
	InternalGetPostgresAddon(ctx context.Context, id string) (*models.PostgresAddon, *errors.ServiceError)
	UpdatePostgresAddonStatus(ctx context.Context, id string, status *models.PostgresAddonStatus) *errors.ServiceError
	InternalList(ctx context.Context, query string, args ...any) ([]*models.PostgresAddon, *errors.ServiceError)
	InternalDeleteFromDB(ctx context.Context, id string) *errors.ServiceError
}

type objectStoreService interface {
	InternalGetByID(ctx context.Context, ID string) (*models.ObjectStore, *errors.ServiceError)
	UpdateStatus(ctx context.Context, id string, status models.ObjectStoreStatus) *errors.ServiceError
}

type namespaceService interface {
	Get(ctx context.Context, id string) (*models.Namespace, *errors.ServiceError)
}

type secretService interface {
	InternalGetByID(ctx context.Context, id string) (*models.Secret, *errors.ServiceError)
}

type connectionUsageChecker interface {
	IsResourceReferencedAsSource(ctx context.Context, resourceType, resourceID string) (bool, error)
}
