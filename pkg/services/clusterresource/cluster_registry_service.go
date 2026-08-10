package clusterresource

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	registryv1alpha1 "stackdome.io/cluster-agent/api/registry/v1alpha1"
)

//go:generate mockgen -destination=../../mocks/mock_cluster_resource_image_registry_service.go -package=mocks -mock_names=ClusterImageRegistryService=MockClusterResourceImageRegistryService github.com/Stackdome/stackdome/pkg/services/clusterresource ClusterImageRegistryService

type ClusterImageRegistryService interface {
	EnsureImageRegistryInCluster(ctx context.Context, registry *models.ClusterImageRegistry) *ClusterResourceError
	DeleteImageRegistryInCluster(ctx context.Context, registry *models.ClusterImageRegistry) (bool, *ClusterResourceError)
}

type clusterImageRegistryService struct {
	clusterManager clustermanager.ClusterManager
	logger         logger.Logger
}

type ClusterImageRegistryServiceSpec struct {
	ClusterManager clustermanager.ClusterManager
	Logger         logger.Logger
}

func NewClusterImageRegistryService(spec ClusterImageRegistryServiceSpec) ClusterImageRegistryService {
	return &clusterImageRegistryService{
		clusterManager: spec.ClusterManager,
		logger:         spec.Logger,
	}
}

func (s *clusterImageRegistryService) EnsureImageRegistryInCluster(ctx context.Context, registry *models.ClusterImageRegistry) *ClusterResourceError {
	clusterClient, clientErr := s.clientForRegistry(ctx, registry)
	if clientErr != nil {
		return clientErr
	}

	desired := desiredImageRegistryObject(registry)
	existing := &registryv1alpha1.ClusterRegistry{}
	if err := clusterClient.Get(ctx, client.ObjectKeyFromObject(desired), existing); err != nil {
		if !k8sapierrors.IsNotFound(err) {
			return newError("failed to inspect image registry in cluster", err)
		}
		if err := clusterClient.Create(ctx, desired); err != nil && !k8sapierrors.IsAlreadyExists(err) {
			return newError("failed to create image registry in cluster", err)
		}
		return nil
	}

	if existing.Labels[models.ImageRegistryIDLabel] != registry.ID {
		return newError(
			"refusing to manage mismatched image registry",
			fmt.Errorf("ClusterRegistry %q does not belong to registry %q", existing.Name, registry.ID),
		)
	}

	if apiequality.Semantic.DeepEqual(existing.Spec, desired.Spec) {
		return nil
	}
	existing.Spec = desired.Spec
	if err := clusterClient.Update(ctx, existing); err != nil {
		return newError("failed to update image registry in cluster", err)
	}
	return nil
}

// DeleteImageRegistryInCluster returns true only once the Kubernetes object is
// absent. A delete accepted by Kubernetes is retried until a later Get returns
// NotFound, keeping the database row as durable deletion intent meanwhile.
func (s *clusterImageRegistryService) DeleteImageRegistryInCluster(ctx context.Context, registry *models.ClusterImageRegistry) (bool, *ClusterResourceError) {
	clusterClient, clientErr := s.clientForRegistry(ctx, registry)
	if clientErr != nil {
		return false, clientErr
	}

	desired := desiredImageRegistryObject(registry)
	existing := &registryv1alpha1.ClusterRegistry{}
	if err := clusterClient.Get(ctx, client.ObjectKeyFromObject(desired), existing); err != nil {
		if k8sapierrors.IsNotFound(err) {
			return true, nil
		}
		return false, newError("failed to inspect image registry in cluster", err)
	}
	if existing.Labels[models.ImageRegistryIDLabel] != registry.ID {
		return false, newError(
			"refusing to manage mismatched image registry",
			fmt.Errorf("ClusterRegistry %q does not belong to registry %q", existing.Name, registry.ID),
		)
	}
	if !existing.DeletionTimestamp.IsZero() {
		return false, nil
	}
	if err := clusterClient.Delete(ctx, existing); err != nil && !k8sapierrors.IsNotFound(err) {
		return false, newError("failed to delete image registry in cluster", err)
	}
	return false, nil
}

func (s *clusterImageRegistryService) clientForRegistry(ctx context.Context, registry *models.ClusterImageRegistry) (client.Client, *ClusterResourceError) {
	if registry == nil {
		return nil, newError("failed to get cluster client", fmt.Errorf("image registry is nil"))
	}
	if registry.ClusterID == "" {
		return nil, newError("failed to get cluster client", fmt.Errorf("image registry cluster ID is empty"))
	}
	clusterClient, err := s.clusterManager.GetClient(registry.ClusterID)
	if err != nil {
		s.logger.Error(ctx, "failed to get cluster client: %v", err)
		return nil, newError("failed to get cluster client", err)
	}
	return clusterClient, nil
}

func desiredImageRegistryObject(registry *models.ClusterImageRegistry) *registryv1alpha1.ClusterRegistry {
	result := &registryv1alpha1.ClusterRegistry{
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
			Port: 5000,
		},
	}
	if registry.BackendStorageClass != "" {
		result.Spec.Storage.StorageClass = ptr.To(registry.BackendStorageClass)
	}
	return result
}
