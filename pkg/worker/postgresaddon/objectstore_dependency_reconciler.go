package postgresaddon

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type objectStoreDependencyReconciler struct {
	objectStoreService objectStoreService
	secretService      secretService
	clusterManager     clustermanager.ClusterManager
	logger             logger.Logger
}

func newObjectStoreDependencyReconciler(spec PostgresAddonWorkerSpec) *objectStoreDependencyReconciler {
	return &objectStoreDependencyReconciler{
		objectStoreService: spec.ObjectStoreService,
		secretService:      spec.SecretService,
		clusterManager:     spec.ClusterManager,
		logger:             logger.NewLoggerWithPrefix(context.Background(), "postgres-addon-objectstore"),
	}
}

func (r *objectStoreDependencyReconciler) Name() string { return "objectstore-dependency" }

func (r *objectStoreDependencyReconciler) Reconcile(ctx context.Context, addon *models.PostgresAddon) (subReconcilerResult, error) {
	if addon.BackupConfig.ObjectStoreID == "" {
		return resultNil, nil
	}

	objectStore, err := r.objectStoreService.GetByID(ctx, addon.BackupConfig.ObjectStoreID)
	if err != nil {
		return resultNil, fmt.Errorf("failed to get object store '%s': %w", addon.BackupConfig.ObjectStoreID, err)
	}

	if _, cerr := r.clusterManager.GetClient(addon.ClusterID); cerr != nil {
		return resultNil, fmt.Errorf("failed to get cluster client: %w", cerr)
	}

	for _, deployed := range objectStore.Status.DeployedClusters {
		if deployed.ClusterID == addon.ClusterID && deployed.Namespace == addon.Namespace {
			return resultNil, nil
		}
	}

	r.logger.Infof("Creating ObjectStore CR '%s' in namespace '%s'", objectStore.Name, addon.Namespace)

	if err := r.createObjectStoreCR(ctx, objectStore, addon.Namespace); err != nil {
		return resultNil, fmt.Errorf("failed to create ObjectStore CR: %w", err)
	}

	objectStore.Status.DeployedClusters = append(objectStore.Status.DeployedClusters, models.DeployedClusterInfo{
		ClusterID: addon.ClusterID,
		Namespace: addon.Namespace,
	})
	if serr := r.objectStoreService.UpdateStatus(ctx, objectStore.ID, objectStore.Status); serr != nil {
		return resultNil, fmt.Errorf("failed to update object store status: %w", serr)
	}

	return resultNil, nil
}

func (r *objectStoreDependencyReconciler) createObjectStoreCR(_ context.Context, _ *models.ObjectStore, _ string) error {
	// TODO: Create barman-cloud ObjectStore CR using typed API when dependency is available
	return nil
}
