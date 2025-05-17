package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type ClusterImageRegistryStore interface {
	GetByID(ctx context.Context, ID string) (*models.ClusterImageRegistry, *errors.ServiceError)
	GetForOrg(ctx context.Context, orgID string) (*models.ClusterImageRegistry, *errors.ServiceError)
	ListByClusterID(ctx context.Context, clusterID string) ([]*models.ClusterImageRegistry, *errors.ServiceError)
	Create(ctx context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError)
	CreateWithTx(ctx context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError)
	UpdateStatus(ctx context.Context, ID string, status *models.ClusterImageRegistryStatus) *errors.ServiceError
	Delete(ctx context.Context, ID string) *errors.ServiceError
	DeleteWithTx(ctx context.Context, ID string) *errors.ServiceError
	AtomicExecutor
}
