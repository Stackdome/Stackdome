package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -destination=../mocks/mock_cluster_store.go -package=mocks github.com/Stackdome/stackdome/pkg/stores ClusterStore
type ClusterStore interface {
	Create(ctx context.Context, spec *models.Cluster) (*models.Cluster, *errors.ServiceError)
	CreateWithTx(ctx context.Context, spec *models.Cluster) (*models.Cluster, *errors.ServiceError)
	GetClusterForOrg(ctx context.Context, orgID string) (*models.Cluster, *errors.ServiceError)
	GetPlatformCluster(ctx context.Context) (*models.Cluster, *errors.ServiceError)
	Get(ctx context.Context, ID string) (*models.Cluster, *errors.ServiceError)
	PersistManagerState(ctx context.Context, ID string, running bool) *errors.ServiceError
	Delete(ctx context.Context, ID string) *errors.ServiceError
	GetByClusterUrl(ctx context.Context, clusterURL string) (*models.Cluster, *errors.ServiceError)
	UpdateCredentials(ctx context.Context, id, encToken, encCAData string) *errors.ServiceError
	UpdateNameAndPlatform(ctx context.Context, id, name string) *errors.ServiceError
	UpdateClusterInfo(ctx context.Context, ID string, info *models.ClusterInfo) *errors.ServiceError
	ListAll(ctx context.Context) ([]*models.Cluster, *errors.ServiceError)
	AtomicExecutor
}
