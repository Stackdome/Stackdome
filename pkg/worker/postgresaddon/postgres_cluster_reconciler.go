package postgresaddon

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/builders"
	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"k8s.io/apimachinery/pkg/api/equality"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	addonsv1alpha1 "stackdome.io/cluster-agent/api/addons/v1alpha1"
)

type postgresClusterReconciler struct {
	clusterManager       clustermanager.ClusterManager
	postgresAddonService postgresAddonService
	objectStoreService   objectStoreService
	crBuilder            builders.PostgresClusterBuilder
	logger               logger.Logger
}

func newPostgresClusterReconciler(spec PostgresAddonWorkerSpec) *postgresClusterReconciler {
	return &postgresClusterReconciler{
		clusterManager:       spec.ClusterManager,
		postgresAddonService: spec.PostgresAddonService,
		objectStoreService:   spec.ObjectStoreService,
		crBuilder:            spec.CRBuilder,
		logger:               logger.NewLoggerWithPrefix(context.Background(), "postgres-addon-cluster"),
	}
}

func (r *postgresClusterReconciler) Name() string { return "postgres-cluster" }

func (r *postgresClusterReconciler) Reconcile(ctx context.Context, addon *models.PostgresAddon) (subReconcilerResult, error) {
	clusterClient, err := r.clusterManager.GetClient(addon.ClusterID)
	if err != nil {
		return resultNil, fmt.Errorf("failed to get cluster client: %w", err)
	}

	buildCtx, err := r.buildContext(ctx, addon)
	if err != nil {
		return resultNil, fmt.Errorf("failed to build context: %w", err)
	}

	desiredCR, err := r.crBuilder.BuildPostgresClusterCR(addon, buildCtx)
	if err != nil {
		return resultNil, fmt.Errorf("failed to build PostgresCluster CR: %w", err)
	}

	existingCR := &addonsv1alpha1.PostgresCluster{}
	if err := clusterClient.Get(ctx, client.ObjectKeyFromObject(desiredCR), existingCR); err != nil {
		if k8sapierrors.IsNotFound(err) {
			r.logger.Info(ctx, "Creating PostgresCluster CR '%s' in namespace '%s'", desiredCR.Name, desiredCR.Namespace)
			return resultNil, clusterClient.Create(ctx, desiredCR)
		}
		return resultNil, fmt.Errorf("failed to get PostgresCluster CR: %w", err)
	}

	if !equality.Semantic.DeepEqual(existingCR.Spec, desiredCR.Spec) {
		r.logger.Info(ctx, "Updating PostgresCluster CR '%s'", desiredCR.Name)
		desiredCR.ResourceVersion = existingCR.ResourceVersion
		return resultNil, clusterClient.Update(ctx, desiredCR)
	}

	return resultNil, nil
}

func (r *postgresClusterReconciler) buildContext(ctx context.Context, addon *models.PostgresAddon) (builders.PostgresClusterBuildContext, error) {
	buildCtx := builders.PostgresClusterBuildContext{}

	if addon.BackupConfig.ObjectStoreID != "" {
		objectStore, err := r.objectStoreService.InternalGetByID(ctx, addon.BackupConfig.ObjectStoreID)
		if err != nil {
			return buildCtx, fmt.Errorf("failed to get object store: %w", err)
		}
		buildCtx.BackupObjectStoreName = objectStore.Name
	}

	if addon.Initialization.RestoreFromObjectStore != nil {
		if addon.Initialization.RestoreFromObjectStore.ObjectStoreID != "" {
			restoreObjectStore, rerr := r.objectStoreService.InternalGetByID(ctx, addon.Initialization.RestoreFromObjectStore.ObjectStoreID)
			if rerr != nil {
				return buildCtx, fmt.Errorf("failed to get restore object store: %w", rerr)
			}
			buildCtx.RestoreObjectStoreName = restoreObjectStore.Name
		}

		// SourceClusterName is the name of the PostgresCluster CR that originally
		// archived backups to the object store. Since CR names match addon names,
		// we resolve this by looking up the source addon.
		sourceAddon, serr := r.postgresAddonService.InternalGetPostgresAddon(ctx, addon.Initialization.RestoreFromObjectStore.SourcePostgresAddonID)
		if serr != nil {
			return buildCtx, fmt.Errorf("failed to get source postgres addon '%s': %w", addon.Initialization.RestoreFromObjectStore.SourcePostgresAddonID, serr)
		}
		buildCtx.RecoverySourceClusterName = sourceAddon.Name
	}

	return buildCtx, nil
}
