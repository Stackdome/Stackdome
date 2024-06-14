package stores

import (
	"context"

	"github.com/ashishmax31/soradev-api-server/pkg/errors"
	"github.com/ashishmax31/soradev-api-server/pkg/models"
)

type ClusterStore interface {
	Create(ctx context.Context, spec *models.Cluster) (*models.Cluster, *errors.ServiceError)
	GetClusterForOrg(ctx context.Context, orgID int) (*models.Cluster, *errors.ServiceError)
	GetDefaultCluster(ctx context.Context) (*models.Cluster, *errors.ServiceError)
	Get(ctx context.Context, ID string) (*models.Cluster, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
}
