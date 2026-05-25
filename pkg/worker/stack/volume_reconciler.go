package stack

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/builders"
	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
	storagev1alpha1 "stackdome.io/cluster-agent/api/storage/v1alpha1"
)

type volumeReconciler struct {
	clusterManager  clustermanager.ClusterManager
	volumeService   volumeService
	volumeCRbuilder builders.ClusterResourceBuilder
}

type VolumeReconcilerSpec struct {
	ClusterManager  clustermanager.ClusterManager
	VolumeService   volumeService
	VolumeCrBuilder builders.ClusterResourceBuilder
}

func NewVolumeReconciler(spec VolumeReconcilerSpec) *volumeReconciler {
	return &volumeReconciler{
		clusterManager:  spec.ClusterManager,
		volumeService:   spec.VolumeService,
		volumeCRbuilder: spec.VolumeCrBuilder,
	}
}
func (r *volumeReconciler) Reconcile(ctx context.Context, stack *models.Stack) (subReconcilerResult, error) {
	var volumes []*models.Volume
	if len(stack.Volumes) > 0 {
		volumes = stack.Volumes
	} else {
		loaded, serr := r.volumeService.ListVolumesUsedByStack(ctx, stack.ID)
		if serr != nil {
			return resultNil, fmt.Errorf("failed to list volumes for stack '%s': %w", stack.ID, serr)
		}
		volumes = loaded
	}

	clusterClient, cerr := r.clusterManager.GetClient(stack.ClusterID)
	if cerr != nil {
		return resultNil, fmt.Errorf("failed to get cluster client for cluster %s: %w", stack.ClusterID, cerr)
	}

	for _, volume := range volumes {
		volumeCR, err := r.volumeCRbuilder.BuildVolumeCR(ctx, volume)
		if err != nil {
			return resultNil, fmt.Errorf("failed to build volume CR for volume '%s': %w", volume.ID, err)
		}
		existingVolumeCR := &storagev1alpha1.Volume{}
		if err := clusterClient.Get(ctx, client.ObjectKey{Name: volumeCR.Name, Namespace: volumeCR.Namespace}, existingVolumeCR); err != nil {
			if k8sapierrors.IsNotFound(err) {
				if err := clusterClient.Create(ctx, volumeCR); err != nil {
					return resultNil, fmt.Errorf("failed to create volume CR for volume '%s': %w", volume.ID, err)
				}
				continue
			}
			return resultNil, fmt.Errorf("failed to get volume CR for volume '%s': %w", volume.ID, err)
		}

		if volume.VolumeSource == nil {
			continue
		}
		// update git revision
		if volume.VolumeSource.GitRepoSource != nil {
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
			}

			if existingVolumeCR.Spec.Source.GitRepo.Revision != updatedRevision {
				existingVolumeCR.Spec.Source.GitRepo.Revision = updatedRevision
				if err := clusterClient.Update(ctx, existingVolumeCR); err != nil {
					return resultNil, fmt.Errorf("failed to update volume CR for volume '%s': %w", volume.ID, err)
				}
			}
			continue
		}

		// update source revision
		if volume.VolumeSource.RemoteDirSource != nil {
			if existingVolumeCR.Spec.Source.RemoteDir.CurrentDirectoryHash != volume.VolumeSource.RemoteDirSource.CurrentDirectoryHash {
				existingVolumeCR.Spec.Source.RemoteDir.CurrentDirectoryHash = volume.VolumeSource.RemoteDirSource.CurrentDirectoryHash
				if err := clusterClient.Update(ctx, existingVolumeCR); err != nil {
					return resultNil, fmt.Errorf("failed to update volume CR for volume '%s': %w", volume.ID, err)
				}
			}
		}
	}

	return resultNil, nil
}

func (r *volumeReconciler) Name() string {
	return "volume-reconciler"
}
