package clusterresource

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	workspacev1alpha1 "soradev.io/cluster-agent/api/v1alpha1"
)

// WorkspaceStorageClusterResourceService is an interface for managing workspace storage resources in a cluster.
type WorkspaceStorageClusterResourceService interface {
	UpsertWorkspaceStorageInCluster(ctx context.Context, workspaceStorage *models.WorkspaceStorage) *ClusterResourceError
	DeleteWorkspaceStorageInCluster(ctx context.Context, workspaceStorage *models.WorkspaceStorage) *ClusterResourceError
}

type WorkspaceStorageClusterResourceServiceSpec struct {
	ClusterService       DBClusterService
	WorkspaceUserService DBWorkspaceUserService
	ClusterManager       clustermanager.ClusterManager
	Logger               logger.Logger
}

type workspaceStorageClusterService struct {
	clusterService       DBClusterService
	workspaceUserService DBWorkspaceUserService
	clusterManager       clustermanager.ClusterManager
	logger               logger.Logger
}

func NewWorkspaceStorageClusterResourceService(spec WorkspaceStorageClusterResourceServiceSpec) WorkspaceStorageClusterResourceService {
	return &workspaceStorageClusterService{
		clusterService:       spec.ClusterService,
		workspaceUserService: spec.WorkspaceUserService,
		clusterManager:       spec.ClusterManager,
		logger:               spec.Logger,
	}
}

func (w *workspaceStorageClusterService) CreateWorkspaceStorageInCluster(ctx context.Context, workspaceStorage *models.WorkspaceStorage) *ClusterResourceError {
	cluster, err := w.clusterService.GetClusterForOrg(ctx, workspaceStorage.OrganisationID)
	if err != nil {
		w.logger.Errorf("failed to get cluster for org: %v", err)
		return newError("failed to get cluster for org", err)
	}

	workspaceUser, err := w.workspaceUserService.GetWorkspaceUser(ctx, workspaceStorage.UserID)
	if err != nil {
		w.logger.Errorf("failed to get workspace user: %v", err)
		return newError("failed to get workspace user", err)
	}

	workspaceStorageCR := w.desiredObjectInCluster(workspaceStorage, workspaceUser)

	clusterClient, clientGetErr := w.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		w.logger.Errorf("failed to get cluster client: %v", clientGetErr)
		return newError("failed to get cluster client", clientGetErr)
	}

	if err := clusterClient.Create(ctx, workspaceStorageCR); err != nil {
		w.logger.Errorf("failed to create workspace storage in cluster: %v", err)
		return newError("failed to create workspace storage in cluster", err)
	}

	return nil
}

func (w *workspaceStorageClusterService) UpsertWorkspaceStorageInCluster(ctx context.Context, workspaceStorage *models.WorkspaceStorage) *ClusterResourceError {
	cluster, err := w.clusterService.GetClusterForOrg(ctx, workspaceStorage.OrganisationID)
	if err != nil {
		w.logger.Errorf("failed to get cluster for org: %v", err)
		return newError("failed to get cluster for org", err)
	}

	workspaceUser, err := w.workspaceUserService.GetWorkspaceUser(ctx, workspaceStorage.UserID)
	if err != nil {
		w.logger.Errorf("failed to get workspace user: %v", err)
		return newError("failed to get workspace user", err)
	}

	desiredWorkspaceStorageCR := w.desiredObjectInCluster(workspaceStorage, workspaceUser)

	clusterClient, clientGetErr := w.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		w.logger.Errorf("failed to get cluster client: %v", clientGetErr)
		return newError("failed to get cluster client", clientGetErr)
	}

	existingWorkspaceStorageCR := &workspacev1alpha1.WorkspaceStorage{}
	if err := clusterClient.Get(ctx, client.ObjectKeyFromObject(desiredWorkspaceStorageCR), existingWorkspaceStorageCR); err != nil {
		if k8sapierrors.IsNotFound(err) {
			if err := clusterClient.Create(ctx, desiredWorkspaceStorageCR); err != nil {
				w.logger.Errorf("failed to create workspace storage in cluster: %v", err)
				return newError("failed to create workspace storage in cluster", err)
			}
			return nil
		}
		w.logger.Errorf("failed to upsert workspace storage in cluster: %v", err)
		return newError("failed to upsert workspace storage in cluster", err)
	}

	// TODO: Use server side apply.
	desiredWorkspaceStorageCR.ResourceVersion = existingWorkspaceStorageCR.ResourceVersion
	if err := clusterClient.Update(ctx, desiredWorkspaceStorageCR); err != nil {
		w.logger.Errorf("failed to update workspace storage in cluster: %v", err)
		return newError("failed to update workspace storage in cluster", err)
	}
	return nil
}

func (w *workspaceStorageClusterService) DeleteWorkspaceStorageInCluster(ctx context.Context, workspaceStorage *models.WorkspaceStorage) *ClusterResourceError {
	cluster, err := w.clusterService.GetClusterForOrg(ctx, workspaceStorage.OrganisationID)
	if err != nil {
		w.logger.Errorf("failed to get cluster for org: %v", err)
		return newError("failed to get cluster for org", err)
	}

	workspaceUser, err := w.workspaceUserService.GetWorkspaceUser(ctx, workspaceStorage.UserID)
	if err != nil {
		w.logger.Errorf("failed to get workspace user: %v", err)
		return newError("failed to get workspace user", err)
	}

	workspaceStorageCR := w.desiredObjectInCluster(workspaceStorage, workspaceUser)

	clusterClient, clientGetErr := w.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		w.logger.Errorf("failed to get cluster client: %v", clientGetErr)
		return newError("failed to get cluster client", clientGetErr)
	}

	if err := clusterClient.Delete(ctx, workspaceStorageCR); err != nil {
		if k8sapierrors.IsNotFound(err) {
			w.logger.Warn(ctx, "workspace storage '%s' not found in cluster", workspaceStorage.ID)
			return nil
		}
		w.logger.Errorf("failed to delete workspace storage in cluster: %v", err)
		return newError("failed to delete workspace storage in cluster", err)
	}

	return nil
}

func (w *workspaceStorageClusterService) desiredObjectInCluster(workspaceStorage *models.WorkspaceStorage, wsUser *models.WorkspaceUser) *workspacev1alpha1.WorkspaceStorage {
	wsStorageCRLabels := workspaceStorage.Labels.ToMap()
	wsStorageCRLabels[models.WorkspaceStorageIDLabel] = workspaceStorage.ID
	wsStorageCRLabels[models.ObjectServerGeneration] = fmt.Sprintf("%d", workspaceStorage.Version)
	res := workspacev1alpha1.WorkspaceStorage{
		ObjectMeta: metav1.ObjectMeta{
			Name:        workspaceStorage.Name,
			Namespace:   workspaceStorage.Namespace,
			Labels:      wsStorageCRLabels,
			Annotations: workspaceStorage.Annotations.ToMap(),
		},
		Spec: workspacev1alpha1.WorkspaceStorageSpec{
			WorkspaceName:    workspaceStorage.WorkspaceName,
			UserPublicSSHKey: wsUser.SshPublicKey,
		},
	}

	resourceStorageSpecs := make(map[workspacev1alpha1.VolumeName]*workspacev1alpha1.WorkspaceVolumeSpec)
	for _, volume := range workspaceStorage.Volumes {
		currVolumeLabels := volume.Labels.ToMap()
		currVolumeLabels[models.WorkspaceStorageIDLabel] = workspaceStorage.ID

		currVolumeSpec := &workspacev1alpha1.WorkspaceVolumeSpec{
			Size:               volume.Size,
			StorageClass:       volume.StorageClass,
			NeedsSyncBeforeUse: volume.SyncBeforeUse,
			Labels:             currVolumeLabels,
			Annotations:        volume.Annotations.ToMap(),
		}
		if volume.VolumeSource != nil {
			switch {
			case volume.VolumeSource.LocalSource != nil:
				currVolumeSpec.Source = &workspacev1alpha1.VolumeSource{
					LocalDir: &workspacev1alpha1.LocalDirSource{
						Path:          volume.VolumeSource.LocalSource.Path,
						ContinousSync: volume.VolumeSource.LocalSource.Sync,
					},
				}

			case len(volume.VolumeSource.BuildSource) > 0:
				buildSrcs := make([]workspacev1alpha1.BuildArtifactSource, len(volume.VolumeSource.BuildSource))
				for i, buildSrc := range volume.VolumeSource.BuildSource {
					buildSrcs[i] = workspacev1alpha1.BuildArtifactSource{
						ResourceRef:     workspacev1alpha1.ResourceRef(buildSrc.ResourceName),
						SourcePath:      buildSrc.SourcePath,
						DestinationPath: buildSrc.DestinationPath,
					}
				}
				currVolumeSpec.Source = &workspacev1alpha1.VolumeSource{
					BuildArtifacts: buildSrcs,
				}
			}
		}
		resourceStorageSpecs[workspacev1alpha1.VolumeName(volume.Name)] = currVolumeSpec
	}
	res.Spec.ResourceStorageSpecs = resourceStorageSpecs

	return &res
}
