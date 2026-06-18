package release

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net"

	"github.com/ashishmax31/stackdome-api-server/pkg/builders"
	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
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

	ready, err := r.volumesReady(ctx, release.StackID)
	if err != nil {
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

func (r *applyReconciler) volumesReady(ctx context.Context, stackID string) (bool, error) {
	volumes, serr := r.volumeService.ListVolumesUsedByStack(ctx, stackID)
	if serr != nil {
		return false, serr
	}
	for _, v := range volumes {
		if v.Status == nil || v.Status.Phase != "Ready" {
			return false, nil
		}
	}
	return true, nil
}
