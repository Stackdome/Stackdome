package release

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/builders"
	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/Stackdome/stackdome/pkg/worker/workermanager"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	storagev1alpha1 "stackdome.io/cluster-agent/api/storage/v1alpha1"
)

type provisioningPrerequisiteReconciler struct {
	runtimePolicy    runtimePolicy
	stackService     stackService
	volumeService    volumeService
	namespaceService namespaceService
	postgresAddons   postgresAddonService
	enqueuer         workermanager.BackgroundJobEnqueuer
	clusterManager   clustermanager.ClusterManager
	crBuilder        builders.ClusterResourceBuilder
}

func newProvisioningPrerequisiteReconciler(spec ReleaseWorkerSpec) *provisioningPrerequisiteReconciler {
	return &provisioningPrerequisiteReconciler{
		runtimePolicy: spec.RuntimePolicy, stackService: spec.StackService,
		volumeService: spec.VolumeService, namespaceService: spec.NamespaceService,
		postgresAddons: spec.PostgresAddonService, enqueuer: spec.ReleaseWorkerEnqueuer,
		clusterManager: spec.ClusterManager,
		crBuilder:      spec.CRBuilder,
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
	if release.StackID != snapshot.ID {
		return resultNil, fmt.Errorf("release %s stack identity does not match its snapshot", release.ID)
	}
	if err := r.runtimePolicy.RequireComputeAccess(ctx, snapshot.OrganisationID); err != nil {
		return resultNil, err
	}
	expectedPolicyVersion := r.runtimePolicy.IsolationPolicyVersion()
	if expectedPolicyVersion == "" {
		return resultNil, fmt.Errorf("cloud isolation policy version is not configured")
	}
	if snapshot.NamespaceID == "" || snapshot.Namespace == "" {
		return resultNil, fmt.Errorf("release %s has no persisted namespace", release.ID)
	}
	namespace, serr := r.namespaceService.Get(ctx, snapshot.NamespaceID)
	if serr != nil {
		return resultNil, serr
	}
	if namespace.ID != snapshot.NamespaceID || namespace.Name != snapshot.Namespace || namespace.OrganisationID != snapshot.OrganisationID {
		return resultNil, fmt.Errorf("release %s namespace identity does not match persisted namespace", release.ID)
	}
	stack, serr := r.stackService.InternalGetStack(ctx, release.StackID)
	if serr != nil {
		return resultNil, serr
	}
	if stack.ID != snapshot.ID || stack.OrganisationID != snapshot.OrganisationID || stack.ClusterID != snapshot.ClusterID ||
		stack.NamespaceID != snapshot.NamespaceID || stack.Namespace != snapshot.Namespace || stack.DeletionTimestamp != nil {
		return resultNil, fmt.Errorf("release %s stack identity does not match its snapshot", release.ID)
	}

	volumes := make([]*models.Volume, 0, len(release.Snapshot.Volumes))
	for _, snapshotVolume := range release.Snapshot.Volumes {
		if snapshotVolume == nil || snapshotVolume.ID == "" {
			return resultNil, fmt.Errorf("release %s contains a volume without an id", release.ID)
		}
		volume, serr := r.volumeService.InternalGet(ctx, snapshotVolume.ID)
		if serr != nil {
			return resultNil, serr
		}
		if volume.ID != snapshotVolume.ID || volume.OrganisationID != snapshotVolume.OrganisationID ||
			volume.ProjectID != snapshotVolume.ProjectID || volume.UserID != snapshotVolume.UserID ||
			volume.Name != snapshotVolume.Name || volume.NamespaceID != snapshotVolume.NamespaceID ||
			volume.Namespace != snapshotVolume.Namespace {
			return resultNil, fmt.Errorf("release %s volume %s identity does not match its snapshot", release.ID, snapshotVolume.ID)
		}
		volumes = append(volumes, snapshotVolume)
	}
	addons, err := r.referencedPostgresAddons(ctx, release.Snapshot.Connections)
	if err != nil {
		return resultNil, err
	}

	if err := r.enqueuer.Enqueue(models.StackOperand{ID: release.StackID, ReleaseID: release.ID}); err != nil {
		return resultNil, fmt.Errorf("enqueue stack prerequisite: %w", err)
	}
	for _, volume := range volumes {
		if err := r.enqueuer.Enqueue(models.VolumeOperand{ID: volume.ID, ReleaseID: release.ID}); err != nil {
			return resultNil, fmt.Errorf("enqueue volume prerequisite %s: %w", volume.ID, err)
		}
	}
	for _, addon := range addons {
		if err := r.enqueuer.Enqueue(models.PostgresAddonOperand{ID: addon.ID, ReleaseID: release.ID}); err != nil {
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
	if observedNamespace.Labels[models.CloudPolicyReadyLabelKey] != expectedPolicyVersion {
		return resultRequeue, nil
	}
	for _, volume := range volumes {
		desired, buildErr := r.crBuilder.BuildVolumeCR(ctx, volume)
		if buildErr != nil {
			return resultNil, fmt.Errorf("build volume prerequisite %s: %w", volume.ID, buildErr)
		}
		setPrerequisiteVolumeReleaseID(desired, release.ID)
		observed := &storagev1alpha1.Volume{}
		if err := clusterClient.Get(ctx, client.ObjectKey{Name: desired.Name, Namespace: desired.Namespace}, observed); err != nil {
			if k8sapierrors.IsNotFound(err) {
				return resultRequeue, nil
			}
			return resultNil, fmt.Errorf("observe volume prerequisite %s: %w", volume.ID, err)
		}
		if !volumeReadyForRelease(observed, desired, release.ID) {
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

func setPrerequisiteVolumeReleaseID(volume *storagev1alpha1.Volume, releaseID string) {
	if volume.Annotations == nil {
		volume.Annotations = make(map[string]string)
	}
	volume.Annotations[models.VolumeReleaseIDAnnotation] = releaseID
}

func volumeReadyForRelease(observed, desired *storagev1alpha1.Volume, releaseID string) bool {
	if observed.Annotations[models.VolumeReleaseIDAnnotation] != releaseID ||
		!containsDesiredMetadata(observed.Labels, desired.Labels) ||
		!containsDesiredMetadata(observed.Annotations, desired.Annotations) ||
		!apiequality.Semantic.DeepEqual(observed.Spec, desired.Spec) ||
		observed.Generation <= 0 || observed.Status.ObservedGeneration != observed.Generation ||
		observed.Status.Phase != storagev1alpha1.VolumePhaseReady {
		return false
	}
	if desired.Spec.Source == nil {
		return true
	}
	if desired.Spec.Source.RemoteDir != nil {
		return observed.Status.LastRemoteSyncHash == desired.Spec.Source.RemoteDir.CurrentDirectoryHash
	}
	if desired.Spec.Source.GitRepo != nil {
		return observed.Status.LastSyncedGitReference == desired.Spec.Source.GitRepo.Revision.GetGitRevisionString()
	}
	return true
}

func containsDesiredMetadata(observed, desired map[string]string) bool {
	for key, value := range desired {
		if observed[key] != value {
			return false
		}
	}
	return true
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
