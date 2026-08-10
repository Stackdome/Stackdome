package release

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services/clusterresource"
	"github.com/Stackdome/stackdome/pkg/stores"
)

type registryPrerequisiteReconciler struct {
	imageRegistryStore    stores.ClusterImageRegistryStore
	imageRegistryResource clusterresource.ClusterImageRegistryService
}

func newRegistryPrerequisiteReconciler(spec ReleaseWorkerSpec) *registryPrerequisiteReconciler {
	return &registryPrerequisiteReconciler{
		imageRegistryStore:    spec.ImageRegistryStore,
		imageRegistryResource: spec.ImageRegistryResource,
	}
}

func (r *registryPrerequisiteReconciler) Name() string { return "registry-prerequisite" }

func (r *registryPrerequisiteReconciler) Reconcile(ctx context.Context, release *models.StackRelease) (subReconcilerResult, error) {
	if !snapshotUsesInClusterRegistry(release.Snapshot) {
		return resultNil, nil
	}

	registry, serr := r.imageRegistryStore.GetForOrgAndCluster(
		ctx,
		release.Snapshot.Stack.OrganisationID,
		release.Snapshot.Stack.ClusterID,
	)
	if serr != nil {
		return resultNil, fmt.Errorf("load in-cluster registry prerequisite: %w", serr)
	}
	if resourceErr := r.imageRegistryResource.EnsureImageRegistryInCluster(ctx, registry); resourceErr != nil {
		return resultNil, fmt.Errorf("ensure in-cluster registry prerequisite: %w", resourceErr)
	}

	return resultNil, nil
}

func snapshotUsesInClusterRegistry(snapshot models.StackSnapshot) bool {
	for _, resource := range snapshot.Resources {
		if resource != nil && resource.BuildConfig != nil && resource.BuildConfig.BuildImageRepository.UseInClusterRegistry {
			return true
		}
	}
	return false
}
