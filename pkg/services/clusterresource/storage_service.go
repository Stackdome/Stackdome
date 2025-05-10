package clusterresource

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

// StackStorageClusterResourceService is an interface for managing workspace storage resources in a cluster.
type StackStorageClusterResourceService interface {
	UpsertStorageInCluster(ctx context.Context, stackStorage *models.StackStorage) *ClusterResourceError
	DeleteStorageInCluster(ctx context.Context, stackStorage *models.StackStorage) *ClusterResourceError
}

// type WorkspaceStorageClusterResourceServiceSpec struct {
// 	ClusterService       DBClusterService
// 	WorkspaceUserService DBWorkspaceUserService
// 	ClusterManager       clustermanager.ClusterManager
// 	Logger               logger.Logger
// }

// type stackStorageClusterService struct {
// 	clusterService       DBClusterService
// 	workspaceUserService DBWorkspaceUserService
// 	clusterManager       clustermanager.ClusterManager
// 	logger               logger.Logger
// }

// func NewWorkspaceStorageClusterResourceService(spec WorkspaceStorageClusterResourceServiceSpec) StackStorageClusterResourceService {
// 	return &stackStorageClusterService{
// 		clusterService:       spec.ClusterService,
// 		workspaceUserService: spec.WorkspaceUserService,
// 		clusterManager:       spec.ClusterManager,
// 		logger:               spec.Logger,
// 	}
// }

// func (w *stackStorageClusterService) CreateStorageInCluster(ctx context.Context, storage *models.StackStorage) *ClusterResourceError {
// 	cluster, err := w.clusterService.GetClusterForOrg(ctx, storage.OrganisationID)
// 	if err != nil {
// 		w.logger.Errorf("failed to get cluster for org: %v", err)
// 		return newError("failed to get cluster for org", err)
// 	}

// 	storageCR := w.desiredObjectInCluster(storage)

// 	clusterClient, clientGetErr := w.clusterManager.GetClient(cluster.ID)
// 	if clientGetErr != nil {
// 		w.logger.Errorf("failed to get cluster client: %v", clientGetErr)
// 		return newError("failed to get cluster client", clientGetErr)
// 	}

// 	if err := clusterClient.Create(ctx, storageCR); err != nil {
// 		w.logger.Errorf("failed to create storage in cluster: %v", err)
// 		return newError("failed to create storage in cluster", err)
// 	}

// 	return nil
// }

// func (w *stackStorageClusterService) UpsertStorageInCluster(ctx context.Context, storage *models.StackStorage) *ClusterResourceError {
// 	cluster, err := w.clusterService.GetClusterForOrg(ctx, storage.OrganisationID)
// 	if err != nil {
// 		w.logger.Errorf("failed to get cluster for org: %v", err)
// 		return newError("failed to get cluster for org", err)
// 	}

// 	desiredStorageCR := w.desiredObjectInCluster(storage)

// 	clusterClient, clientGetErr := w.clusterManager.GetClient(cluster.ID)
// 	if clientGetErr != nil {
// 		w.logger.Errorf("failed to get cluster client: %v", clientGetErr)
// 		return newError("failed to get cluster client", clientGetErr)
// 	}

// 	existingStorageCR := &storagev1alpha1.Storage{}
// 	if err := clusterClient.Get(ctx, client.ObjectKeyFromObject(desiredStorageCR), existingStorageCR); err != nil {
// 		if k8sapierrors.IsNotFound(err) {
// 			if err := clusterClient.Create(ctx, desiredStorageCR); err != nil {
// 				w.logger.Errorf("failed to create  storage in cluster: %v", err)
// 				return newError("failed to create  storage in cluster", err)
// 			}
// 			return nil
// 		}
// 		w.logger.Errorf("failed to upsert storage in cluster: %v", err)
// 		return newError("failed to upsert  storage in cluster", err)
// 	}

// 	// TODO: Use server side apply.
// 	desiredStorageCR.ResourceVersion = existingStorageCR.ResourceVersion
// 	if err := clusterClient.Update(ctx, desiredStorageCR); err != nil {
// 		w.logger.Errorf("failed to update  storage in cluster: %v", err)
// 		return newError("failed to update  storage in cluster", err)
// 	}
// 	return nil
// }

// func (w *stackStorageClusterService) DeleteStorageInCluster(ctx context.Context, storage *models.StackStorage) *ClusterResourceError {
// 	cluster, err := w.clusterService.GetClusterForOrg(ctx, storage.OrganisationID)
// 	if err != nil {
// 		w.logger.Errorf("failed to get cluster for org: %v", err)
// 		return newError("failed to get cluster for org", err)
// 	}

// 	storageCR := w.desiredObjectInCluster(storage)

// 	clusterClient, clientGetErr := w.clusterManager.GetClient(cluster.ID)
// 	if clientGetErr != nil {
// 		w.logger.Errorf("failed to get cluster client: %v", clientGetErr)
// 		return newError("failed to get cluster client", clientGetErr)
// 	}

// 	if err := clusterClient.Delete(ctx, storageCR); err != nil {
// 		if k8sapierrors.IsNotFound(err) {
// 			w.logger.Warn(ctx, "storage '%s' not found in cluster", storage.ID)
// 			return nil
// 		}
// 		w.logger.Errorf("failed to delete  storage in cluster: %v", err)
// 		return newError("failed to delete  storage in cluster", err)
// 	}

// 	return nil
// }

// func (w *stackStorageClusterService) desiredObjectInCluster(storage *models.StackStorage) *storagev1alpha1.Storage {
// 	storageCRLabels := storage.Labels.ToMap()
// 	storageCRLabels[models.StackStorageIDLabel] = storage.ID
// 	storageCRLabels[models.ObjectServerGeneration] = fmt.Sprintf("%d", storage.Version)
// 	res := storagev1alpha1.Storage{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:        storage.Name,
// 			Namespace:   storage.Namespace,
// 			Labels:      storageCRLabels,
// 			Annotations: storage.Annotations.ToMap(),
// 		},
// 		Spec: storagev1alpha1.StorageSpec{
// 			ProvisionedFor:   storage.WorkspaceName,
// 			UserPublicSSHKey: storage.SSHConfig.PublicKey,
// 		},
// 	}

// 	storageSpecs := make(map[storagev1alpha1.VolumeName]*storagev1alpha1.VolumeSpec)
// 	for _, volume := range storage.Volumes {
// 		currVolumeLabels := volume.Labels.ToMap()
// 		currVolumeLabels[models.StackStorageIDLabel] = storage.ID

// 		currVolumeSpec := &storagev1alpha1.VolumeSpec{
// 			Size:               volume.Size,
// 			StorageClass:       volume.StorageClass,
// 			NeedsSyncBeforeUse: volume.SyncBeforeUse,
// 			AccessMode:         corev1.PersistentVolumeAccessMode(volume.AccessMode),
// 			Labels:             currVolumeLabels,
// 			Annotations:        volume.Annotations.ToMap(),
// 		}
// 		if volume.VolumeSource != nil {
// 			switch {
// 			case volume.VolumeSource.LocalSource != nil:
// 				currVolumeSpec.Source = &storagev1alpha1.VolumeSource{
// 					LocalDir: &storagev1alpha1.LocalDirSource{
// 						Path:          volume.VolumeSource.LocalSource.Path,
// 						ContinousSync: volume.VolumeSource.LocalSource.Sync,
// 					},
// 				}

// 			case len(volume.VolumeSource.BuildSource) > 0:
// 				buildSrcs := make([]storagev1alpha1.BuildArtifactSource, len(volume.VolumeSource.BuildSource))
// 				for i, buildSrc := range volume.VolumeSource.BuildSource {
// 					buildSrcs[i] = storagev1alpha1.BuildArtifactSource{
// 						BuildSource: storagev1alpha1.StackResourceReference{
// 							Name: buildSrc.ResourceName,
// 						},
// 						SourcePath:      buildSrc.SourcePath,
// 						DestinationPath: buildSrc.DestinationPath,
// 					}
// 				}
// 				currVolumeSpec.Source = &storagev1alpha1.VolumeSource{
// 					BuildArtifacts: buildSrcs,
// 				}
// 			}
// 		}
// 		storageSpecs[storagev1alpha1.VolumeName(volume.Name)] = currVolumeSpec
// 	}
// 	res.Spec.VolumeSpecs = storageSpecs

// 	return &res
// }
