package release

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/Stackdome/stackdome/pkg/worker/workermanager"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type provisioningPrerequisiteReconciler struct {
	runtimePolicy    runtimePolicy
	stackService     stackService
	volumeService    volumeService
	namespaceService namespaceService
	postgresAddons   postgresAddonService
	enqueuer         workermanager.BackgroundJobEnqueuer
	clusterManager   clustermanager.ClusterManager
}

func newProvisioningPrerequisiteReconciler(spec ReleaseWorkerSpec) *provisioningPrerequisiteReconciler {
	if spec.RuntimePolicy == nil {
		panic("release.newProvisioningPrerequisiteReconciler: RuntimePolicy is required")
	}
	return &provisioningPrerequisiteReconciler{
		runtimePolicy: spec.RuntimePolicy, stackService: spec.StackService,
		volumeService: spec.VolumeService, namespaceService: spec.NamespaceService,
		postgresAddons: spec.PostgresAddonService, enqueuer: spec.ReleaseWorkerEnqueuer,
		clusterManager: spec.ClusterManager,
	}
}

func (r *provisioningPrerequisiteReconciler) Name() string {
	return "provisioning-prerequisite"
}

func (r *provisioningPrerequisiteReconciler) Reconcile(ctx context.Context, release *models.StackRelease) (subReconcilerResult, error) {
	if r.runtimePolicy.DraftProvisioningMode() == services.ProvisioningModeEager {
		return resultNil, nil
	}

	snapshot := release.Snapshot.Stack
	if err := r.runtimePolicy.RequireActiveAllocation(ctx, snapshot.OrganisationID); err != nil {
		return resultNil, err
	}
	if snapshot.NamespaceID == "" || snapshot.Namespace == "" {
		return resultNil, fmt.Errorf("release %s has no persisted namespace", release.ID)
	}
	namespace, serr := r.namespaceService.Get(ctx, snapshot.NamespaceID)
	if serr != nil {
		return resultNil, serr
	}
	if namespace.Name != snapshot.Namespace {
		return resultNil, fmt.Errorf("release %s namespace identity does not match persisted namespace", release.ID)
	}
	if _, serr := r.stackService.InternalGetStack(ctx, release.StackID); serr != nil {
		return resultNil, serr
	}

	volumes, serr := r.volumeService.ListVolumesUsedByStack(ctx, release.StackID)
	if serr != nil {
		return resultNil, serr
	}
	addons, err := r.referencedPostgresAddons(ctx, release.Snapshot.Connections)
	if err != nil {
		return resultNil, err
	}

	if err := r.enqueuer.Enqueue(models.StackOperand{ID: release.StackID}); err != nil {
		return resultNil, fmt.Errorf("enqueue stack prerequisite: %w", err)
	}
	for _, volume := range volumes {
		if err := r.enqueuer.Enqueue(models.VolumeOperand{ID: volume.ID}); err != nil {
			return resultNil, fmt.Errorf("enqueue volume prerequisite %s: %w", volume.ID, err)
		}
	}
	for _, addon := range addons {
		if err := r.enqueuer.Enqueue(models.PostgresAddonOperand{ID: addon.ID}); err != nil {
			return resultNil, fmt.Errorf("enqueue postgres prerequisite %s: %w", addon.ID, err)
		}
	}

	clusterClient, err := r.clusterManager.GetClient(snapshot.ClusterID)
	if err != nil {
		return resultNil, fmt.Errorf("get cluster client for prerequisites: %w", err)
	}
	observedNamespace := &corev1.Namespace{}
	if err := clusterClient.Get(ctx, client.ObjectKey{Name: namespace.Name}, observedNamespace); err != nil {
		if k8sapierrors.IsNotFound(err) {
			return resultRequeue, nil
		}
		return resultNil, fmt.Errorf("observe namespace prerequisite: %w", err)
	}
	for _, label := range namespace.Labels {
		if observedNamespace.Labels[label.Key] != label.Value {
			return resultRequeue, nil
		}
	}
	if observedNamespace.Labels[models.CloudPolicyReadyLabelKey] != models.CloudPolicyReadyVersion {
		return resultRequeue, nil
	}
	for _, volume := range volumes {
		if volume.Status == nil || volume.Status.Phase != models.VolumePhaseReady {
			return resultRequeue, nil
		}
	}
	for _, addon := range addons {
		if addon.Status.State != models.PostgresAddonStateReady {
			return resultRequeue, nil
		}
	}
	return resultNil, nil
}

func (r *provisioningPrerequisiteReconciler) referencedPostgresAddons(ctx context.Context, connections models.StackConnections) ([]*models.PostgresAddon, error) {
	seen := make(map[string]struct{})
	addons := make([]*models.PostgresAddon, 0)
	for _, connection := range connections {
		if connection.From.Type != models.TopologyNodeTypePostgresAddon {
			continue
		}
		if connection.From.Id == "" {
			return nil, fmt.Errorf("postgres prerequisite connection has no addon id")
		}
		if _, exists := seen[connection.From.Id]; exists {
			continue
		}
		addon, serr := r.postgresAddons.InternalGetPostgresAddon(ctx, connection.From.Id)
		if serr != nil {
			return nil, serr
		}
		seen[connection.From.Id] = struct{}{}
		addons = append(addons, addon)
	}
	return addons, nil
}
