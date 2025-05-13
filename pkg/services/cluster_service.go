package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
	certutil "k8s.io/client-go/util/cert"
)

type ClusterService interface {
	GetClusterForOrg(ctx context.Context, orgID string) (*models.Cluster, *errors.ServiceError)
	GetDefaultCluster(ctx context.Context) (*models.Cluster, *errors.ServiceError)
	Get(ctx context.Context, ID string) (*models.Cluster, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	AddCluster(ctx context.Context, cluster *models.Cluster) (*models.Cluster, *errors.ServiceError)
	InternalListAllClusters(ctx context.Context) ([]*models.Cluster, *errors.ServiceError)
	InjectClusterManager(clusterManager clustermanager.ClusterManager)
}

type clusterService struct {
	clusterStore   stores.ClusterStore
	logger         logger.Logger
	clusterManager clustermanager.ClusterManager
}

func NewClusterService(spec ClusterServiceSpec) ClusterService {
	return &clusterService{
		clusterStore: pgstore.NewClusterStore(pgstore.ClusterStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		clusterManager: spec.ClusterManager,
		logger:         spec.Logger,
	}
}

type ClusterServiceSpec struct {
	SessionFactory db.SessionFactory
	ClusterManager clustermanager.ClusterManager
	Logger         logger.Logger
}

// inject cluster manager
func (s *clusterService) InjectClusterManager(clusterManager clustermanager.ClusterManager) {
	s.clusterManager = clusterManager
}

// InternalListAllClusters lists all clusters in the database
func (s *clusterService) InternalListAllClusters(ctx context.Context) ([]*models.Cluster, *errors.ServiceError) {
	clusters, err := s.clusterStore.ListAll(ctx)
	if err != nil {
		s.logger.Errorf("failed to list all clusters: %v", err)
		return nil, err
	}
	return clusters, nil
}

func (s *clusterService) AddCluster(ctx context.Context, cluster *models.Cluster) (*models.Cluster, *errors.ServiceError) {
	// Check if the cluster already exists for the org
	existingCluster, err := s.clusterStore.GetClusterForOrg(ctx, cluster.OrganisationID)
	if err != nil && err.Code != errors.ErrorNotFound {
		s.logger.Errorf("failed to get existing cluster for org: %v", err)
		return nil, err
	}
	if existingCluster != nil {
		return nil, errors.Conflict("cluster already exists for org")
	}
	var (
		createdCluster *models.Cluster
		createdErr     *errors.ServiceError
	)

	// Validate the cluster
	err = s.validateCluster(cluster)
	if err != nil {
		s.logger.Errorf("failed to validate cluster: %v", err)
		return nil, err
	}

	cerr := s.clusterStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		// Create the cluster in the database
		createdCluster, createdErr = s.clusterStore.CreateWithTx(ctx, cluster)
		if createdErr != nil {
			return createdErr
		}

		// Register the cluster with the cluster manager
		merr := s.clusterManager.RegisterCluster(createdCluster)
		if merr != nil {
			return errors.GeneralError("failed to register cluster with manager")
		}
		return nil
	})
	if cerr != nil {
		s.logger.Errorf("failed to create cluster: %v", cerr)
		return nil, cerr
	}
	return createdCluster, nil
}

func (s *clusterService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	// Get the cluster from the database
	cluster, err := s.clusterStore.Get(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to get cluster: %v", err)
		return err
	}

	// Unregister the cluster from the cluster manager
	cerr := s.clusterManager.UnregisterCluster(cluster.ID)
	if cerr != nil {
		s.logger.Errorf("failed to unregister cluster from manager: %v", cerr)
		return errors.GeneralError("failed to unregister cluster from manager: %v", cerr)
	}
	// Delete the cluster from the database
	err = s.clusterStore.Delete(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to delete cluster: %v", err)
		return err
	}
	return nil
}

func (s *clusterService) validateCluster(cluster *models.Cluster) *errors.ServiceError {
	if cluster == nil {
		return errors.BadRequest("cluster cannot be nil")
	}

	if cluster.Name == "" {
		return errors.BadRequest("cluster name cannot be empty")
	}

	if cluster.OrganisationID == "" {
		return errors.BadRequest("organisation ID cannot be empty")
	}

	if cluster.ClusterCAData == "" {
		return errors.BadRequest("cluster CA data cannot be empty")
	}

	if cluster.ClusterURL == "" {
		return errors.BadRequest("cluster URL cannot be empty")
	}
	if cluster.Token == "" {
		return errors.BadRequest("cluster token cannot be empty")
	}

	// validation for cluster url
	url, err := url.Parse(cluster.ClusterURL)
	if err != nil {
		return errors.BadRequest("cluster URL is not valid: %s", err.Error())
	}
	if url.Scheme != "https" {
		return errors.BadRequest("cluster URL must use https scheme")
	}

	existingCluster, serr := s.clusterStore.GetByClusterUrl(context.Background(), cluster.ClusterURL)
	if serr != nil && serr.Code != errors.ErrorNotFound {
		s.logger.Errorf("failed to get cluster by URL: %v", serr)
		return serr
	}

	if existingCluster != nil {
		return errors.Conflict("cluster with this api URL already exists")
	}

	var (
		clusterCADataDecoded []byte
		derr                 error
	)
	if IsBase64(cluster.ClusterCAData) {
		clusterCADataDecoded, derr = base64.StdEncoding.DecodeString(cluster.ClusterCAData)
		if derr != nil {
			return errors.BadRequest("cluster CA data is not valid base64: %s", derr.Error())
		}
		if len(clusterCADataDecoded) == 0 {
			return errors.BadRequest("cluster CA data is empty after decoding")
		}
	} else {
		cluster.ClusterCAData = base64.StdEncoding.EncodeToString([]byte(cluster.ClusterCAData))
	}

	if IsBase64(cluster.Token) {
		tokenDecoded, derr := base64.StdEncoding.DecodeString(cluster.Token)
		if derr != nil {
			return errors.BadRequest("cluster token is not valid base64: %s", derr.Error())
		}
		if len(tokenDecoded) == 0 {
			return errors.BadRequest("cluster token is empty after decoding")
		}
	} else {
		cluster.Token = base64.StdEncoding.EncodeToString([]byte(cluster.Token))
	}
	if _, err := certutil.NewPoolFromBytes(clusterCADataDecoded); err != nil {
		return errors.BadRequest("cluster CA data is not valid: %s", err.Error())
	}
	return nil
}

func (s *clusterService) PersistManagerState(ctx context.Context, clusterID string, running bool) error {
	err := s.clusterStore.PersistManagerState(ctx, clusterID, running)
	if err != nil {
		return fmt.Errorf("failed to persist cluster manager state: %w", err)
	}
	return nil
}

func (s *clusterService) GetClusterForOrg(ctx context.Context, orgID string) (*models.Cluster, *errors.ServiceError) {
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

func IsBase64(s string) bool {
	// Base64 string must be a multiple of 4
	if len(s)%4 != 0 {
		return false
	}

	// Basic character check: must only contain valid base64 characters
	if strings.ContainsAny(s, " \t\r\n") {
		return false
	}

	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}
