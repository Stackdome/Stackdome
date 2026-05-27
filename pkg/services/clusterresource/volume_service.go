package clusterresource

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
	storagev1alpha1 "stackdome.io/cluster-agent/api/storage/v1alpha1"
)

//go:generate mockgen -source=volume_service.go -destination=../../mocks/mock_volume_cluster_resource_service.go -package=mocks

// VolumeClusterResourceService is an interface for managing volume resources in a cluster.
type VolumeClusterResourceService interface {
	UpdateVolumeRemoteDirRevisionInCluster(ctx context.Context, volume *models.Volume) *ClusterResourceError
	UpdateVolumeGitRevisionInCluster(ctx context.Context, volume *models.Volume) *ClusterResourceError
	CreateVolumeInCluster(ctx context.Context, volume *models.Volume) *ClusterResourceError
	DeleteVolumeInCluster(ctx context.Context, volume *models.Volume) *ClusterResourceError
}

type VolumeClusterResourceServiceSpec struct {
	ClusterService       DBClusterService
	WorkspaceUserService DBWorkspaceUserService
	ClusterManager       clustermanager.ClusterManager
	Logger               logger.Logger
}

type volumeClusterService struct {
	clusterService       DBClusterService
	workspaceUserService DBWorkspaceUserService
	clusterManager       clustermanager.ClusterManager
	logger               logger.Logger
}

func NewVolumeClusterResourceService(spec VolumeClusterResourceServiceSpec) VolumeClusterResourceService {
	return &volumeClusterService{
		clusterService:       spec.ClusterService,
		workspaceUserService: spec.WorkspaceUserService,
		clusterManager:       spec.ClusterManager,
		logger:               spec.Logger,
	}
}

func (w *volumeClusterService) CreateVolumeInCluster(ctx context.Context, volume *models.Volume) *ClusterResourceError {
	cluster, err := w.clusterService.GetClusterForOrg(ctx, volume.OrganisationID)
	if err != nil {
		w.logger.Errorf("failed to get cluster for org: %v", err)
		return newError("failed to get cluster for org", err)
	}

	volumeCR, cerr := w.desiredObjectInCluster(volume)
	if cerr != nil {
		w.logger.Errorf("failed to create desired volume object in cluster: %v", cerr)
		return newError("failed to create desired volume object in cluster", cerr)
	}

	clusterClient, clientGetErr := w.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		w.logger.Errorf("failed to get cluster client: %v", clientGetErr)
		return newError("failed to get cluster client", clientGetErr)
	}

	if err := clusterClient.Create(ctx, volumeCR); err != nil {
		w.logger.Errorf("failed to create volume  in cluster: %v", err)
		return newError("failed to create volume  in cluster", err)
	}

	return nil
}

func (w *volumeClusterService) UpdateVolumeRemoteDirRevisionInCluster(ctx context.Context, volume *models.Volume) *ClusterResourceError {
	if volume.VolumeSource == nil || volume.VolumeSource.RemoteDirSource == nil {
		return newError("remote dir source is nil", fmt.Errorf("remote dir source is nil"))
	}

	cluster, err := w.clusterService.GetClusterForOrg(ctx, volume.OrganisationID)
	if err != nil {
		w.logger.Errorf("failed to get cluster for org: %v", err)
		return newError("failed to get cluster for org", err)
	}

	clusterClient, clientGetErr := w.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		w.logger.Errorf("failed to get cluster client: %v", clientGetErr)
		return newError("failed to get cluster client", clientGetErr)
	}

	existingVolumeCR := &storagev1alpha1.Volume{}
	if err := clusterClient.Get(ctx, types.NamespacedName{
		Name:      volume.Name,
		Namespace: volume.Namespace,
	}, existingVolumeCR); err != nil {
		if k8sapierrors.IsNotFound(err) {
			w.logger.Errorf("volume missing in cluster: %v", err)
			return newError("volume missing in cluster", err)
		}
		return newError("failed to get volume from cluster", err)
	}

	existingVolumeCR.Spec.Source.RemoteDir.CurrentDirectoryHash = volume.VolumeSource.RemoteDirSource.CurrentDirectoryHash
	if err := clusterClient.Update(ctx, existingVolumeCR); err != nil {
		w.logger.Errorf("failed to update volume in cluster: %v", err)
		return newError("failed to update volume in cluster", err)
	}
	return nil
}

func (w *volumeClusterService) UpdateVolumeGitRevisionInCluster(ctx context.Context, volume *models.Volume) *ClusterResourceError {
	if volume.VolumeSource == nil || volume.VolumeSource.GitRepoSource == nil {
		return newError("git repo source is nil", fmt.Errorf("git repo source is nil"))
	}

	cluster, err := w.clusterService.GetClusterForOrg(ctx, volume.OrganisationID)
	if err != nil {
		w.logger.Errorf("failed to get cluster for org: %v", err)
		return newError("failed to get cluster for org", err)
	}

	clusterClient, clientGetErr := w.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		w.logger.Errorf("failed to get cluster client: %v", clientGetErr)
		return newError("failed to get cluster client", clientGetErr)
	}

	existingVolumeCR := &storagev1alpha1.Volume{}
	if err := clusterClient.Get(ctx, types.NamespacedName{
		Name:      volume.Name,
		Namespace: volume.Namespace,
	}, existingVolumeCR); err != nil {
		if k8sapierrors.IsNotFound(err) {
			w.logger.Errorf("volume missing in cluster: %v", err)
			return newError("volume missing in cluster", err)
		}
		return newError("failed to get volume from cluster", err)
	}
	updatedRevision := corev1alpha1.GitRepoRevision{}
	switch volume.VolumeSource.GitRepoSource.Revision.Type() {
	case models.Branch:
		updatedRevision.Branch = &corev1alpha1.GitBranch{
			Name: volume.VolumeSource.GitRepoSource.Revision.Branch.Name,
		}
		if volume.VolumeSource.GitRepoSource.Revision.Branch.HeadSha != "" {
			updatedRevision.Branch.HeadSha = volume.VolumeSource.GitRepoSource.Revision.Branch.HeadSha
		}
	case models.Tag:
		updatedRevision.Tag = volume.VolumeSource.GitRepoSource.Revision.Tag
	case models.Commit:
		updatedRevision.Commit = volume.VolumeSource.GitRepoSource.Revision.Commit
	default:
		return newError("unknown git revision type", fmt.Errorf("unknown git revision type: %s", volume.VolumeSource.GitRepoSource.Revision.Type()))
	}

	existingVolumeCR.Spec.Source.GitRepo.Revision = updatedRevision
	if err := clusterClient.Update(ctx, existingVolumeCR); err != nil {
		w.logger.Errorf("failed to update volume in cluster: %v", err)
		return newError("failed to update volume in cluster", err)
	}
	return nil
}

func (w *volumeClusterService) DeleteVolumeInCluster(ctx context.Context, volume *models.Volume) *ClusterResourceError {
	cluster, err := w.clusterService.GetClusterForOrg(ctx, volume.OrganisationID)
	if err != nil {
		w.logger.Errorf("failed to get cluster for org: %v", err)
		return newError("failed to get cluster for org", err)
	}

	clusterClient, clientGetErr := w.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		w.logger.Errorf("failed to get cluster client: %v", clientGetErr)
		return newError("failed to get cluster client", clientGetErr)
	}

	existingVolumeCR := &storagev1alpha1.Volume{}
	if err := clusterClient.Get(ctx, types.NamespacedName{
		Name:      volume.Name,
		Namespace: volume.Namespace,
	}, existingVolumeCR); err != nil {
		if k8sapierrors.IsNotFound(err) {
			w.logger.Errorf("volume missing in cluster: %v", err)
			return newError("volume missing in cluster", err)
		}
		return newError("failed to get volume from cluster", err)
	}

	if err := clusterClient.Delete(ctx, existingVolumeCR); err != nil {
		if k8sapierrors.IsNotFound(err) {
			w.logger.Warn(ctx, "volume '%s' not found in cluster", volume.ID)
			return nil
		}
		w.logger.Errorf("failed to delete  volume in cluster: %v", err)
		return newError("failed to delete  volume in cluster", err)
	}
	return nil
}

func (w *volumeClusterService) desiredObjectInCluster(volume *models.Volume) (*storagev1alpha1.Volume, error) {
	volumeCRLabels := volume.Labels.ToMap()
	volumeCRLabels[models.VolumeIDLabel] = volume.ID
	volumeCRLabels[models.CreatedForUserLabel] = volume.UserID
	// storageCRLabels[models.ObjectServerGeneration] = fmt.Sprintf("%d", volume.Version)
	res := storagev1alpha1.Volume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        volume.Name,
			Namespace:   volume.Namespace,
			Labels:      volumeCRLabels,
			Annotations: volume.Annotations.ToMap(),
		},
		Spec: storagev1alpha1.VolumeSpec{
			Annotations:        volume.Annotations.ToMap(),
			Labels:             volumeCRLabels,
			Size:               volume.Size,
			StorageClass:       volume.StorageClass,
			AccessMode:         corev1.PersistentVolumeAccessMode(volume.AccessMode),
			NeedsSyncBeforeUse: volume.SyncBeforeUse,
		},
	}

	if volume.VolumeSource != nil {
		switch {
		case volume.VolumeSource.RemoteDirSource != nil:
			res.Spec.Source = &storagev1alpha1.VolumeSource{
				RemoteDir: &storagev1alpha1.RemoteDirSource{
					Path:                 volume.VolumeSource.RemoteDirSource.Path,
					CurrentDirectoryHash: volume.VolumeSource.RemoteDirSource.CurrentDirectoryHash,
				},
			}
		case len(volume.VolumeSource.BuildSource) > 0:
			buildSrcs := make([]storagev1alpha1.BuildArtifactSource, len(volume.VolumeSource.BuildSource))
			for i, buildSrc := range volume.VolumeSource.BuildSource {
				buildSrcs[i] = storagev1alpha1.BuildArtifactSource{
					BuildSource: storagev1alpha1.StackResourceReference{
						Name: buildSrc.ResourceName,
					},
					SourcePath:      buildSrc.SourcePath,
					DestinationPath: buildSrc.DestinationPath,
				}
			}
			res.Spec.Source = &storagev1alpha1.VolumeSource{
				BuildArtifacts: buildSrcs,
			}
		case volume.VolumeSource.GitRepoSource != nil:
			gitRevision := corev1alpha1.GitRepoRevision{}

			switch volume.VolumeSource.GitRepoSource.Revision.Type() {
			case models.Branch:
				gitRevision.Branch = &corev1alpha1.GitBranch{
					Name: volume.VolumeSource.GitRepoSource.Revision.Branch.Name,
				}
			case models.Tag:
				gitRevision.Tag = volume.VolumeSource.GitRepoSource.Revision.Tag
			case models.Commit:
				gitRevision.Commit = volume.VolumeSource.GitRepoSource.Revision.Commit
			default:
				return nil, fmt.Errorf("unknown git revision type: %s", volume.VolumeSource.GitRepoSource.Revision.Type())
			}

			res.Spec.Source = &storagev1alpha1.VolumeSource{
				GitRepo: &storagev1alpha1.GitRepoSource{
					RepoUrl:  volume.VolumeSource.GitRepoSource.RepoUrl,
					Revision: gitRevision,
				},
			}
		}
	}

	return &res, nil
}
