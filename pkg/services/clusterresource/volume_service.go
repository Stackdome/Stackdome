package clusterresource

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	storagev1alpha1 "stackdome.io/cluster-agent/api/storage/v1alpha1"
)

//go:generate mockgen -source=volume_service.go -destination=../../mocks/mock_volume_cluster_resource_service.go -package=mocks

// VolumeClusterResourceService is an interface for managing volume resources in a cluster.
type VolumeClusterResourceService interface {
	DeleteVolumeInCluster(ctx context.Context, volume *models.Volume) *ClusterResourceError
}

type VolumeClusterResourceServiceSpec struct {
	ClusterService DBClusterService
	ClusterManager clustermanager.ClusterManager
	Logger         logger.Logger
}

type volumeClusterService struct {
	clusterService DBClusterService
	clusterManager clustermanager.ClusterManager
	logger         logger.Logger
}

func NewVolumeClusterResourceService(spec VolumeClusterResourceServiceSpec) VolumeClusterResourceService {
	return &volumeClusterService{
		clusterService: spec.ClusterService,
		clusterManager: spec.ClusterManager,
		logger:         spec.Logger,
	}
}

func (w *volumeClusterService) DeleteVolumeInCluster(ctx context.Context, volume *models.Volume) *ClusterResourceError {
	cluster, err := w.clusterService.GetClusterForOrg(ctx, volume.OrganisationID)
	if err != nil {
		w.logger.Error(ctx, "failed to get cluster for org: %v", err)
		return newError("failed to get cluster for org", err)
	}

	clusterClient, clientGetErr := w.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		w.logger.Error(ctx, "failed to get cluster client: %v", clientGetErr)
		return newError("failed to get cluster client", clientGetErr)
	}

	existingVolumeCR := &storagev1alpha1.Volume{}
	if err := clusterClient.Get(ctx, types.NamespacedName{
		Name:      volume.Name,
		Namespace: volume.Namespace,
	}, existingVolumeCR); err != nil {
		if k8sapierrors.IsNotFound(err) {
			w.logger.Error(ctx, "volume missing in cluster: %v", err)
			return newError("volume missing in cluster", err)
		}
		return newError("failed to get volume from cluster", err)
	}

	if err := clusterClient.Delete(ctx, existingVolumeCR); err != nil {
		if k8sapierrors.IsNotFound(err) {
			w.logger.Warn(ctx, "volume '%s' not found in cluster", volume.ID)
			return nil
		}
		w.logger.Error(ctx, "failed to delete  volume in cluster: %v", err)
		return newError("failed to delete  volume in cluster", err)
	}
	return nil
}
