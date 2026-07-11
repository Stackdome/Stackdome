package clusterresource

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	registryv1alpha1 "stackdome.io/cluster-agent/api/registry/v1alpha1"
)

type ClusterImageRegistryService interface {
	CreateImageRegistryInCluster(ctx context.Context, registry *models.ClusterImageRegistry) *ClusterResourceError
	DeleteImageRegistryInCluster(ctx context.Context, registry *models.ClusterImageRegistry) *ClusterResourceError
}

type clusterImageRegistryService struct {
	clusterService DBClusterService
	clusterManager clustermanager.ClusterManager
	logger         logger.Logger
}

type ClusterImageRegistryServiceSpec struct {
	ClusterService DBClusterService
	ClusterManager clustermanager.ClusterManager
	Logger         logger.Logger
}

func NewClusterImageRegistryService(spec ClusterImageRegistryServiceSpec) ClusterImageRegistryService {
	return &clusterImageRegistryService{
		clusterService: spec.ClusterService,
		clusterManager: spec.ClusterManager,
		logger:         spec.Logger,
	}
}

func (s *clusterImageRegistryService) CreateImageRegistryInCluster(ctx context.Context, registry *models.ClusterImageRegistry) *ClusterResourceError {
	cluster, err := s.clusterService.GetClusterForOrg(ctx, registry.OrganisationID)
	if err != nil {
		return newError("failed to get cluster for organisation", err)
	}
	clusterClient, clientGetErr := s.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		s.logger.Error(ctx, "failed to get cluster client: %v", clientGetErr)
		return newError("failed to get cluster client", clientGetErr)
	}

	desiredObjectInCluster := s.desiredObjectInCluster(registry)

	if err := clusterClient.Create(ctx, desiredObjectInCluster); err != nil {
		s.logger.Error(ctx, "failed to create imageregistry in cluster: %v", err)
		return newError("failed to create image registry in cluster", err)
	}
	return nil
}

func (s *clusterImageRegistryService) DeleteImageRegistryInCluster(ctx context.Context, registry *models.ClusterImageRegistry) *ClusterResourceError {
	cluster, err := s.clusterService.GetClusterForOrg(ctx, registry.OrganisationID)
	if err != nil {
		return newError("failed to get cluster for organisation", err)
	}
	clusterClient, clientGetErr := s.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		s.logger.Error(ctx, "failed to get cluster client: %v", clientGetErr)
		return newError("failed to get cluster client", clientGetErr)
	}

	desiredObjectInCluster := s.desiredObjectInCluster(registry)

	if err := clusterClient.Delete(ctx, desiredObjectInCluster); err != nil {
		if k8sapierrors.IsNotFound(err) {
			s.logger.Warn(ctx, "image registry '%s' not found in cluster", registry.ID)
			return nil
		}
		s.logger.Error(ctx, "failed to delete imageregistry in cluster: %v", err)
		return newError("failed to delete image registry in cluster", err)
	}
	return nil
}

func (s *clusterImageRegistryService) desiredObjectInCluster(registry *models.ClusterImageRegistry) *registryv1alpha1.ClusterRegistry {
	res := &registryv1alpha1.ClusterRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name: registry.Name,
			Labels: map[string]string{
				models.ImageRegistryIDLabel: registry.ID,
			},
		},
		Spec: registryv1alpha1.ClusterRegistrySpec{
			Owner: registryv1alpha1.RegistryOwner{
				Type: "Organization",
				ID:   registry.OrganisationID,
			},
			Storage: registryv1alpha1.RegistryStorageSpec{
				Size: registry.BackendStorageSize,
			},
			Port: int32(5000),
		},
	}

	if registry.BackendStorageClass != "" {
		res.Spec.Storage.StorageClass = ptr.To(registry.BackendStorageClass)
	}
	return res
}
