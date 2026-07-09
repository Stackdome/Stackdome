//go:generate mockgen -source=interfaces.go -destination=interfaces_mock_test.go -package=postgresaddon
package postgresaddon

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type clusterClientGetter interface {
	GetClient(clusterID string) (client.Client, error)
}

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

type referenceService interface {
	IsReferentInUse(ctx context.Context, referentType models.ReferentType, referentID string) (bool, []models.ResourceReference, *errors.ServiceError)
}
