package postgresaddon

import (
	"context"
	"fmt"
	"time"

	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	addonsv1alpha1 "stackdome.io/cluster-agent/api/addons/v1alpha1"
)

type deprovisionReconciler struct {
	postgresAddonService postgresAddonService
	referenceService     referenceService
	clusterManager       clustermanager.ClusterManager
	logger               logger.Logger
}

func newDeprovisionReconciler(spec PostgresAddonWorkerSpec) *deprovisionReconciler {
	return &deprovisionReconciler{
		postgresAddonService: spec.PostgresAddonService,
		referenceService:     spec.ReferenceService,
		clusterManager:       spec.ClusterManager,
		logger:               logger.NewLoggerWithPrefix(context.Background(), "postgres-addon-deprovision"),
	}
}

func (r *deprovisionReconciler) Name() string { return "deprovision" }

func (r *deprovisionReconciler) Reconcile(ctx context.Context, addon *models.PostgresAddon) (subReconcilerResult, error) {
	// TODO: Use addon.DeletionTimestamp != nil instead of state string check
	// (see docs/plans/postgres-addon-improvements.md #1).
	if addon.Status.State != models.PostgresAddonStateDeleting {
		return resultNil, nil
	}

	inUse, _, err := r.referenceService.IsReferentInUse(ctx, models.ReferentPostgresAddon, addon.ID)
	if err != nil {
		return resultNil, fmt.Errorf("failed to check addon usage: %w", err)
	}
	if inUse {
		r.logger.Infof("Postgres addon '%s' is still in use by stacks, requeueing", addon.Name)
		return resultRequeueAfter(30 * time.Second), nil
	}

	r.logger.Infof("Deprovisioning postgres addon '%s' in namespace '%s'", addon.Name, addon.Namespace)

	clusterClient, cerr := r.clusterManager.GetClient(addon.ClusterID)
	if cerr != nil {
		return resultNil, fmt.Errorf("failed to get cluster client: %w", cerr)
	}

	pgClusterCR := &addonsv1alpha1.PostgresCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      addon.Name,
			Namespace: addon.Namespace,
		},
	}
	if err := clusterClient.Delete(ctx, pgClusterCR, &client.DeleteOptions{
		PropagationPolicy: ptr.To(metav1.DeletePropagationForeground),
	}); client.IgnoreNotFound(err) != nil {
		return resultNil, fmt.Errorf("failed to delete PostgresCluster CR '%s': %w", addon.Name, err)
	}

	if serr := r.postgresAddonService.InternalDeleteFromDB(ctx, addon.ID); serr != nil {
		return resultNil, fmt.Errorf("failed to delete postgres addon from DB: %w", serr)
	}

	r.logger.Infof("Successfully deprovisioned postgres addon '%s'", addon.Name)
	return resultStop, nil
}
