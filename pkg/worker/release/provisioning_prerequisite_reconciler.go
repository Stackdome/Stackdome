package release

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/builders"
	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/Stackdome/stackdome/pkg/services/clusterresource"
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
	imageRegistries  services.ImageRegistryService
	registryResource clusterresource.ClusterImageRegistryService
	enqueuer         workermanager.BackgroundJobEnqueuer
	clusterManager   clustermanager.ClusterManager
	crBuilder        builders.ClusterResourceBuilder
}

func newProvisioningPrerequisiteReconciler(spec ReleaseWorkerSpec) *provisioningPrerequisiteReconciler {
	return &provisioningPrerequisiteReconciler{
		runtimePolicy: spec.RuntimePolicy, stackService: spec.StackService,
		volumeService: spec.VolumeService, namespaceService: spec.NamespaceService,
		postgresAddons: spec.PostgresAddonService, enqueuer: spec.ReleaseWorkerEnqueuer,
		imageRegistries: spec.ImageRegistryService, registryResource: spec.ImageRegistryResource,
		clusterManager: spec.ClusterManager,
		crBuilder:      spec.CRBuilder,
	}
}

func (r *provisioningPrerequisiteReconciler) Name() string {
	return "provisioning-prerequisite"
}

func (r *provisioningPrerequisiteReconciler) Reconcile(ctx context.Context, release *models.StackRelease) (subReconcilerResult, error) {
	snapshot := release.Snapshot.Stack
	if release.StackID != snapshot.ID {
		return resultNil, fmt.Errorf("release %s stack identity does not match its snapshot", release.ID)
	}
	expectedPolicyVersion := r.runtimePolicy.IsolationPolicyVersion()
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
	if err := r.ensureImageRegistry(ctx, &release.Snapshot); err != nil {
		return resultNil, err
	}

	volumeByID := make(map[string]*models.Volume, len(release.Snapshot.Volumes))
	volumeByName := make(map[string]*models.Volume, len(release.Snapshot.Volumes))
	for _, snapshotVolume := range release.Snapshot.Volumes {
		if snapshotVolume == nil || snapshotVolume.ID == "" {
			return resultNil, fmt.Errorf("release %s contains a volume without an id", release.ID)
		}
		volumeByID[snapshotVolume.ID] = snapshotVolume
		volumeByName[snapshotVolume.Name] = snapshotVolume
	}
	volumeRefs := referencedVolumeRefs(&release.Snapshot)
	volumes := make([]*models.Volume, 0, len(volumeRefs))
	for _, ref := range volumeRefs {
		snapshotVolume := volumeByID[ref.id]
		if snapshotVolume == nil {
			snapshotVolume = volumeByName[ref.name]
		}
		if snapshotVolume == nil {
			return resultNil, fmt.Errorf("release %s references an unknown volume", release.ID)
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
	if expectedPolicyVersion != "" && observedNamespace.Labels[models.CloudPolicyReadyLabelKey] != expectedPolicyVersion {
		return resultRequeue, nil
	}
	for _, volume := range volumes {
		desired, buildErr := r.crBuilder.BuildVolumeCR(ctx, volume)
		if buildErr != nil {
			return resultNil, fmt.Errorf("build volume prerequisite %s: %w", volume.ID, buildErr)
		}
		observed := &storagev1alpha1.Volume{}
		if err := clusterClient.Get(ctx, client.ObjectKey{Name: desired.Name, Namespace: desired.Namespace}, observed); err != nil {
			if k8sapierrors.IsNotFound(err) {
				return resultRequeue, nil
			}
			return resultNil, fmt.Errorf("observe volume prerequisite %s: %w", volume.ID, err)
		}
		if !volumeReadyForRelease(observed, desired) {
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

func (r *provisioningPrerequisiteReconciler) ensureImageRegistry(ctx context.Context, snapshot *models.StackSnapshot) error {
	needsRegistry := false
	for _, stackResource := range snapshot.Resources {
		if stackResource != nil && stackResource.BuildConfig != nil && stackResource.BuildConfig.BuildImageRepository.UseInClusterRegistry {
			needsRegistry = true
			break
		}
	}
	if !needsRegistry {
		return nil
	}
	registry, serr := r.imageRegistries.InternalGetForOrgAndCluster(ctx, snapshot.Stack.OrganisationID, snapshot.Stack.ClusterID)
	if serr != nil {
		return serr
	}
	for _, stackResource := range snapshot.Resources {
		if stackResource != nil && stackResource.BuildConfig != nil && stackResource.BuildConfig.BuildImageRepository.UseInClusterRegistry {
			stackResource.BuildConfig.BuildImageRepository.ClusterRegistryName = registry.Name
		}
	}
	if resourceErr := r.registryResource.EnsureImageRegistryInCluster(ctx, registry); resourceErr != nil {
		return fmt.Errorf("ensure image registry %s: %w", registry.ID, resourceErr)
	}
	return nil
}

func volumeReadyForRelease(observed, desired *storagev1alpha1.Volume) bool {
	if !containsDesiredMetadata(observed.Labels, desired.Labels) ||
		!containsDesiredMetadata(observed.Annotations, desired.Annotations) ||
		!apiequality.Semantic.DeepEqual(observed.Spec, desired.Spec) ||
		observed.Generation <= 0 || observed.Status.ObservedGeneration != observed.Generation ||
		observed.Status.Phase != storagev1alpha1.VolumePhaseReady {
		return false
	}
	if desired.Spec.Source == nil {
		return true
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
