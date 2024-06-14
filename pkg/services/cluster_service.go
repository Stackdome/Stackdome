package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

type ClusterService interface {
	GetClusterForOrg(ctx context.Context, orgID int) (*models.Cluster, *errors.ServiceError)
	GetDefaultCluster(ctx context.Context) (*models.Cluster, *errors.ServiceError)
	Get(ctx context.Context, ID string) (*models.Cluster, *errors.ServiceError)
}

type clusterService struct {
	clusterStore stores.ClusterStore
	logger       logger.Logger
}

func NewClusterService(spec ClusterServiceSpec) ClusterService {
	return &clusterService{
		clusterStore: pgstore.NewClusterStore(pgstore.ClusterStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		logger: spec.Logger,
	}
}

type ClusterServiceSpec struct {
	SessionFactory db.SessionFactory
	Logger         logger.Logger
}

func (s *clusterService) GetClusterForOrg(ctx context.Context, orgID int) (*models.Cluster, *errors.ServiceError) {
	cluster, err := s.clusterStore.GetClusterForOrg(ctx, orgID)
	if err != nil {
		s.logger.Errorf("failed to get cluster for org: %v", err)
		return nil, err
	}
	return cluster, nil
}

func (s *clusterService) GetDefaultCluster(ctx context.Context) (*models.Cluster, *errors.ServiceError) {
	cluster, err := s.clusterStore.GetDefaultCluster(ctx)
	if err != nil {
		s.logger.Errorf("failed to get default cluster: %v", err)
		return nil, err
	}
	return cluster, nil
}

func (s *clusterService) Get(ctx context.Context, ID string) (*models.Cluster, *errors.ServiceError) {
	cluster, err := s.clusterStore.Get(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to get cluster: %v", err)
		return nil, err
	}
	return cluster, nil
}
