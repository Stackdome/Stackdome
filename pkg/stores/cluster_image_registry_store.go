package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -destination=../mocks/mock_cluster_image_registry_store.go -package=mocks github.com/Stackdome/stackdome/pkg/stores ClusterImageRegistryStore

type ClusterImageRegistryStore interface {
	GetByID(ctx context.Context, ID string) (*models.ClusterImageRegistry, *errors.ServiceError)
	GetForOrgAndCluster(ctx context.Context, orgID, clusterID string) (*models.ClusterImageRegistry, *errors.ServiceError)
	ListForOrg(ctx context.Context, orgID string) ([]*models.ClusterImageRegistry, *errors.ServiceError)
	ListByClusterID(ctx context.Context, orgID, clusterID string) ([]*models.ClusterImageRegistry, *errors.ServiceError)
	ListByClusterIDInternal(ctx context.Context, clusterID string) ([]*models.ClusterImageRegistry, *errors.ServiceError)
	Create(ctx context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError)
	CreateWithTx(ctx context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError)
	UpdateStatus(ctx context.Context, ID string, status *models.ClusterImageRegistryStatus) *errors.ServiceError
	Delete(ctx context.Context, ID string) *errors.ServiceError
	DeleteWithTx(ctx context.Context, ID string) *errors.ServiceError
	MarkDeletingWithTx(ctx context.Context, ID string) *errors.ServiceError
	MarkAllDeletingByClusterIDWithTx(ctx context.Context, clusterID string) *errors.ServiceError
	AtomicExecutor
}
