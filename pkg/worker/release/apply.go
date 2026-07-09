package release

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/Stackdome/stackdome/pkg/builders"
	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"k8s.io/apimachinery/pkg/api/equality"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

var errReleaseSuperseded = stderrors.New("release superseded by cluster annotation")

type transientApplyError struct {
	err error
}

func (e *transientApplyError) Error() string { return e.err.Error() }
func (e *transientApplyError) Unwrap() error { return e.err }

func wrapIfTransient(err error) error {
	if isTransientClusterError(err) {
		return &transientApplyError{err: err}
	}
	return err
}

func isTransient(err error) bool {
	var t *transientApplyError
	return stderrors.As(err, &t)
}

func isTransientClusterError(err error) bool {
	if k8sapierrors.IsConflict(err) ||
		k8sapierrors.IsServerTimeout(err) ||
		k8sapierrors.IsServiceUnavailable(err) ||
		k8sapierrors.IsTooManyRequests(err) ||
		k8sapierrors.IsInternalError(err) ||
		k8sapierrors.IsTimeout(err) ||
		k8sapierrors.IsNotFound(err) {
		return true
	}
	var netErr *net.OpError
	if stderrors.As(err, &netErr) {
		return true
	}
	return false
}

type applyReconciler struct {
	releaseService       releaseService
	stackService         stackService
	clusterManager       clustermanager.ClusterManager
	secretBuilder        builders.SecretBuilder
	secretService        secretService
	credentialResolver   credentials.Resolver
	postgresAddonService postgresAddonService
	volumeService        volumeService
	logger               logger.Logger
}

func newApplyReconciler(spec ReleaseWorkerSpec) *applyReconciler {
	return &applyReconciler{
		releaseService:       spec.ReleaseService,
		stackService:         spec.StackService,
		clusterManager:       spec.ClusterManager,
		secretBuilder:        spec.SecretBuilder,
		secretService:        spec.SecretService,
		credentialResolver:   spec.CredentialResolver,
		postgresAddonService: spec.PostgresAddonService,
		volumeService:        spec.VolumeService,
		logger:               logger.NewLoggerWithPrefix(context.Background(), "release-apply"),
	}
}

func (r *applyReconciler) Name() string { return "apply" }

func (r *applyReconciler) Reconcile(ctx context.Context, release *models.StackRelease) (subReconcilerResult, error) {
	if release.Manifest == nil {
		return resultNil, nil
	}

	stack, serr := r.stackService.InternalGetStack(ctx, release.StackID)
	if serr != nil {
		if serr.Is404() {
			failRelease(ctx, r.releaseService, r.logger, release, "stack not found")
			return resultStop, nil
		}
		return resultNil, fmt.Errorf("failed to get stack: %w", serr)
	}

	if stack.DeletionTimestamp != nil {
		failRelease(ctx, r.releaseService, r.logger, release, "stack is being deleted")
		return resultStop, nil
	}

	clusterClient, cerr := r.clusterManager.GetClient(stack.ClusterID)
	if cerr != nil {
		return resultNil, fmt.Errorf("failed to get cluster client: %w", cerr)
	}

	effectiveStack := release.Snapshot.ToStack()

	if err := r.syncHubSecrets(ctx, clusterClient, effectiveStack); err != nil {
		if isTransientClusterError(err) {
			r.logger.Warnf("release %s: transient error syncing hub secrets, requeueing: %v", release.ID, err)
			return resultRequeueAfter(convergencePollInterval), nil
		}
		return resultNil, fmt.Errorf("failed to sync hub secrets: %w", err)
	}

	if err := r.syncPostgresCredentialSecrets(ctx, clusterClient, effectiveStack); err != nil {
		if isTransientClusterError(err) {
			r.logger.Warnf("release %s: transient error syncing postgres secrets, requeueing: %v", release.ID, err)
			return resultRequeueAfter(convergencePollInterval), nil
		}
		return resultNil, fmt.Errorf("failed to sync postgres credential secrets: %w", err)
	}

	if err := r.syncGenericSecrets(ctx, clusterClient, effectiveStack); err != nil {
		if isTransientClusterError(err) {
			r.logger.Warnf("release %s: transient error syncing generic secrets, requeueing: %v", release.ID, err)
			return resultRequeueAfter(convergencePollInterval), nil
		}
		return resultNil, fmt.Errorf("failed to sync generic secrets: %w", err)
	}

	ready, err := r.volumesReady(ctx, release)
	if err != nil {
		var missing *missingVolumesError
		if stderrors.As(err, &missing) {
			failRelease(ctx, r.releaseService, r.logger, release, err.Error())
			return resultStop, nil
		}
		return resultNil, fmt.Errorf("failed to check volume readiness: %w", err)
	}
	if !ready {
		r.logger.Infof("release %s: volumes not ready, requeueing", release.ID)
		return resultRequeueAfter(convergencePollInterval), nil
	}

	stackCR, err := r.applyStackCR(ctx, clusterClient, release)
	if err != nil {
		if stderrors.Is(err, errReleaseSuperseded) {
			return resultStop, nil
		}
		if isTransient(err) {
			r.logger.Warnf("release %s: transient error applying stack CR, requeueing: %v", release.ID, err)
			return resultRequeueAfter(convergencePollInterval), nil
		}
		failRelease(ctx, r.releaseService, r.logger, release, fmt.Sprintf("failed to apply stack CR: %v", err))
		return resultStop, nil
	}
	if err := r.applyStackResourceCRs(ctx, clusterClient, release, stackCR); err != nil {
		if isTransient(err) {
			r.logger.Warnf("release %s: transient error applying resource CRs, requeueing: %v", release.ID, err)
			return resultRequeueAfter(convergencePollInterval), nil
		}
		failRelease(ctx, r.releaseService, r.logger, release, fmt.Sprintf("failed to apply stack resource CRs: %v", err))
		return resultStop, nil
	}

	if err := r.pruneStackResources(ctx, clusterClient, stack, release.Manifest.ResourceNames); err != nil {
		r.logger.Errorf("release %s: prune error (non-fatal): %v", release.ID, err)
	}

	return resultNil, nil
}

func (r *applyReconciler) supersededByClusterCR(ctx context.Context, existing *corev1alpha1.Stack, release *models.StackRelease) (bool, error) {
	appliedReleaseID := existing.GetAnnotations()[corev1alpha1.ReleaseIDAnnotation]
	if appliedReleaseID == "" {
		return false, nil
	}

	appliedRelease, serr := r.releaseService.InternalGet(ctx, appliedReleaseID)
	if serr != nil {
		if serr.Is404() {
			return false, nil
		}
		return false, fmt.Errorf("failed to get applied release: %w", serr)
	}

	if appliedRelease.Sequence <= release.Sequence {
		return false, nil
	}

	reason := fmt.Sprintf("superseded by release #%d already applied to cluster", appliedRelease.Sequence)
	if _, err := r.releaseService.MarkSuperseded(ctx, release.ID, reason); err != nil {
		return false, fmt.Errorf("failed to mark release superseded: %w", err)
	}
	r.logger.Infof("release %s: %s", release.ID, reason)
	return true, nil
}

func (r *applyReconciler) applyStackCR(ctx context.Context, clusterClient client.Client, release *models.StackRelease) (*corev1alpha1.Stack, error) {
	desired := &corev1alpha1.Stack{}
	if err := json.Unmarshal(release.Manifest.StackCR, desired); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stack CR: %w", err)
	}

	if desired.Annotations == nil {
		desired.Annotations = make(map[string]string)
	}
	desired.Annotations[corev1alpha1.RevisionAnnotation] = release.ManifestRevision
	desired.Annotations[corev1alpha1.ReleaseIDAnnotation] = release.ID

	existing := &corev1alpha1.Stack{}
	if err := clusterClient.Get(ctx, client.ObjectKey{Name: desired.Name, Namespace: desired.Namespace}, existing); err != nil {
		if k8sapierrors.IsNotFound(err) {
			if err := clusterClient.Create(ctx, desired); err != nil {
				return nil, wrapIfTransient(fmt.Errorf("failed to create stack CR: %w", err))
			}
			return desired, nil
		}
		return nil, wrapIfTransient(fmt.Errorf("failed to get stack CR: %w", err))
	}

	if superseded, err := r.supersededByClusterCR(ctx, existing, release); err != nil {
		return nil, err
	} else if superseded {
		return nil, errReleaseSuperseded
	}

	if equality.Semantic.DeepDerivative(desired.Spec, existing.Spec) &&
		equality.Semantic.DeepDerivative(desired.Annotations, existing.Annotations) &&
		equality.Semantic.DeepDerivative(desired.Labels, existing.Labels) {
		desired.UID = existing.UID
		return desired, nil
	}

	r.logger.Infof("stack CR '%s': updating", desired.Name)
	desired.ResourceVersion = existing.ResourceVersion
	if err := clusterClient.Update(ctx, desired); err != nil {
		return nil, wrapIfTransient(fmt.Errorf("failed to update stack CR: %w", err))
	}
	desired.UID = existing.UID
	return desired, nil
}

func (r *applyReconciler) applyStackResourceCRs(ctx context.Context, clusterClient client.Client, release *models.StackRelease, stackCR *corev1alpha1.Stack) error {
	for _, name := range release.Manifest.ResourceNames {
		crBytes, ok := release.Manifest.StackResourceCRs[name]
		if !ok {
			return fmt.Errorf("manifest missing CR for resource '%s'", name)
		}

		desired := &corev1alpha1.StackResource{}
		if err := json.Unmarshal(crBytes, desired); err != nil {
			return fmt.Errorf("failed to unmarshal CR for resource '%s': %w", name, err)
		}

		if desired.Annotations == nil {
			desired.Annotations = make(map[string]string)
		}
		desired.Annotations[corev1alpha1.RevisionAnnotation] = release.Manifest.ResourceRevisions[name]
		desired.Annotations[corev1alpha1.ReleaseIDAnnotation] = release.ID

		desired.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion: corev1alpha1.GroupVersion.String(),
				Kind:       "Stack",
				Name:       stackCR.Name,
				UID:        stackCR.UID,
			},
		}

		existing := &corev1alpha1.StackResource{}
		if err := clusterClient.Get(ctx, client.ObjectKey{Name: desired.Name, Namespace: desired.Namespace}, existing); err != nil {
			if k8sapierrors.IsNotFound(err) {
				if err := clusterClient.Create(ctx, desired); err != nil {
					return wrapIfTransient(fmt.Errorf("failed to create StackResource CR '%s': %w", name, err))
				}
				r.logger.Infof("resource '%s': created", name)
				continue
			}
			return wrapIfTransient(fmt.Errorf("failed to get StackResource CR '%s': %w", name, err))
		}

		if equality.Semantic.DeepDerivative(desired.Spec, existing.Spec) &&
			equality.Semantic.DeepDerivative(desired.Annotations, existing.Annotations) &&
			equality.Semantic.DeepDerivative(desired.Labels, existing.Labels) {
			continue
		}

		r.logger.Infof("resource '%s': updating", name)
		desired.ResourceVersion = existing.ResourceVersion
		if err := clusterClient.Update(ctx, desired); err != nil {
			return wrapIfTransient(fmt.Errorf("failed to update StackResource CR '%s': %w", name, err))
		}
	}
	return nil
}

func (r *applyReconciler) pruneStackResources(ctx context.Context, clusterClient client.Client, stack *models.Stack, activeNames []string) error {
	activeSet := make(map[string]struct{}, len(activeNames))
	for _, n := range activeNames {
		activeSet[n] = struct{}{}
	}

	srList := &corev1alpha1.StackResourceList{}
	if err := clusterClient.List(ctx, srList, client.InNamespace(stack.Namespace), client.MatchingLabels{
		corev1alpha1.LabelStackID: stack.ID,
	}); err != nil {
		return fmt.Errorf("failed to list StackResource CRs: %w", err)
	}

	for i := range srList.Items {
		resourceName := srList.Items[i].Labels[corev1alpha1.LabelResourceName]
		if _, keep := activeSet[resourceName]; !keep {
			r.logger.Infof("pruning orphaned StackResource CR '%s'", srList.Items[i].Name)
			if err := clusterClient.Delete(ctx, &srList.Items[i], client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil {
				if k8sapierrors.IsNotFound(err) {
					continue
				}
				return fmt.Errorf("failed to delete orphaned StackResource CR '%s': %w", srList.Items[i].Name, err)
			}
		}
	}
	return nil
}

// missingVolumesError signals that the release snapshot references a volume
// that no longer exists in the stack's live volume set. Applying stack/resource
// CRs against a missing volume would leave the cluster agent stalled trying to
// mount a PVC that will never appear, so the release is failed outright rather
// than requeued.
type missingVolumesError struct {
	refs []string
}

func (e *missingVolumesError) Error() string {
	return fmt.Sprintf("release references volume(s) that no longer exist: %s", strings.Join(e.refs, ", "))
}

// volumesReady gates release application on the readiness of only the volumes
// the release snapshot actually references — not every live volume on the
// stack. An unreferenced volume (e.g. one belonging to a resource dropped from
// this release, or a broken volume nobody mounts) must never block a release.
// Status is still read live from the DB (that's the correct, current cluster
// state); it's the SET of volumes being waited on that must come from the
// immutable snapshot, so a volume created after this release was cut can't
// gate it either.
func (r *applyReconciler) volumesReady(ctx context.Context, release *models.StackRelease) (bool, error) {
	byID, byName := referencedVolumeRefs(&release.Snapshot)
	if len(byID) == 0 && len(byName) == 0 {
		return true, nil
	}

	volumes, serr := r.volumeService.ListVolumesUsedByStack(ctx, release.StackID)
	if serr != nil {
		return false, serr
	}

	foundByID := make(map[string]struct{}, len(byID))
	foundByName := make(map[string]struct{}, len(byName))
	ready := true
	for _, v := range volumes {
		_, wantID := byID[v.ID]
		_, wantName := byName[v.Name]
		if !wantID && !wantName {
			continue
		}
		if wantID {
			foundByID[v.ID] = struct{}{}
		}
		if wantName {
			foundByName[v.Name] = struct{}{}
		}
		if v.Status == nil || v.Status.Phase != models.VolumePhaseReady {
			ready = false
		}
	}

	var missing []string
	for id := range byID {
		if _, ok := foundByID[id]; !ok {
			missing = append(missing, id)
		}
	}
	for name := range byName {
		if _, ok := foundByName[name]; !ok {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return false, &missingVolumesError{refs: missing}
	}

	return ready, nil
}

// referencedVolumeRefs computes the set of volumes referenced by a release
// snapshot: each resource's direct VolumeMounts and build-context volume
// source, plus volume_mount connections.
//
// Connections need separate handling because the stackdeploy Resolver
// (pkg/stackdeploy/volume.go) only materializes volume_mount connections into
// resource.VolumeMounts at render time, on a throwaway Stack rebuilt from the
// snapshot (release.Snapshot.ToStack()) — that mutation never gets written
// back into the persisted snapshot. So resource.VolumeMounts, as stored in
// release.Snapshot.Resources, does NOT reflect connection-driven mounts; the
// connections themselves must be walked directly to recover them.
//
// build_artifact_source connections are deliberately excluded: they point
// FROM a stack resource TO a volume (the resource's build output is copied
// into that volume by the volume controller), so the volume is never
// referenced by the Stack/StackResource CRs this reconciler applies — it's
// managed entirely by the separate volume worker/controller.
func referencedVolumeRefs(snapshot *models.StackSnapshot) (byID, byName map[string]struct{}) {
	byID = make(map[string]struct{})
	byName = make(map[string]struct{})

	addRef := func(id, name string) {
		switch {
		case id != "":
			byID[id] = struct{}{}
		case name != "":
			byName[name] = struct{}{}
		}
	}

	for _, resource := range snapshot.Resources {
		if resource == nil {
			continue
		}
		for _, vm := range resource.VolumeMounts {
			if vm == nil {
				continue
			}
			addRef(vm.SourceVolumeID, vm.SourceVolumeName)
		}
		if resource.BuildConfig != nil && resource.BuildConfig.SourceContext.Volume != nil {
			vol := resource.BuildConfig.SourceContext.Volume
			addRef(vol.SourceVolumeID, vol.SourceVolumeName)
		}
	}

	for _, conn := range snapshot.Connections {
		if conn.Kind != models.ConnectionKindVolumeMount {
			continue
		}
		addRef(conn.From.Id, conn.From.Name)
	}

	return byID, byName
}
