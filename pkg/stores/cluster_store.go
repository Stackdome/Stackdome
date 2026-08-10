package stores

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -destination=../mocks/mock_cluster_store.go -package=mocks github.com/Stackdome/stackdome/pkg/stores ClusterStore
type ClusterStore interface {
	Create(ctx context.Context, spec *models.Cluster) (*models.Cluster, *errors.ServiceError)
	CreateWithTx(ctx context.Context, spec *models.Cluster) (*models.Cluster, *errors.ServiceError)
	ListBYOCClustersForOrg(ctx context.Context, orgID string) ([]*models.Cluster, *errors.ServiceError)
	ListSharedComputeClusters(ctx context.Context) ([]*models.Cluster, *errors.ServiceError)
	ListSharedComputeClustersForOrg(ctx context.Context, orgID string) ([]*models.Cluster, *errors.ServiceError)
	Get(ctx context.Context, ID string) (*models.Cluster, *errors.ServiceError)
	PersistManagerState(ctx context.Context, ID string, running bool) *errors.ServiceError
	Delete(ctx context.Context, ID string) *errors.ServiceError
	DeleteWithTx(ctx context.Context, ID string) *errors.ServiceError
	MarkDeletingWithTx(ctx context.Context, ID string, at time.Time) *errors.ServiceError
	ListIDsForImageRegistryReconciliation(ctx context.Context) ([]string, *errors.ServiceError)
	GetByClusterUrl(ctx context.Context, clusterURL string) (*models.Cluster, *errors.ServiceError)
	UpdateSharedComputeCluster(ctx context.Context, cluster *models.Cluster) *errors.ServiceError
	UpdateClusterInfo(ctx context.Context, ID string, info *models.ClusterInfo) *errors.ServiceError
	ListAll(ctx context.Context) ([]*models.Cluster, *errors.ServiceError)
	AtomicExecutor
}
