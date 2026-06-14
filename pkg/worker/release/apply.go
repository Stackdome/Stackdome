package release

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/worker"
	"k8s.io/apimachinery/pkg/api/equality"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

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
		k8sapierrors.IsTimeout(err) {
		return true
	}
	var netErr *net.OpError
	if stderrors.As(err, &netErr) {
		return true
	}
	return false
}

func (w *releaseWorker) applyAndConverge(ctx context.Context, release *models.StackRelease) (worker.Result, *errors.ServiceError) {
	if release.Manifest == nil {
		w.fail(ctx, release, "apply called with nil manifest")
		return worker.Result{}, nil
	}

	// Load live stack for cluster info and deletion check.
	stack, serr := w.stackService.InternalGetStack(ctx, release.StackID)
	if serr != nil {
		if serr.Is404() {
			w.fail(ctx, release, "stack not found")
			return worker.Result{}, nil
		}
		return worker.Result{}, serr
	}

	if stack.DeletionTimestamp != nil {
		w.fail(ctx, release, "stack is being deleted")
		return worker.Result{}, nil
	}

	clusterClient, cerr := w.clusterManager.GetClient(stack.ClusterID)
	if cerr != nil {
		return worker.Result{}, w.WorkerError.NewError("failed to get cluster client: %v", cerr)
	}

	// Preconditions: sync secrets.
	// Reconstruct the effective stack from snapshot for secret methods.
	effectiveStack := release.Snapshot.ToStack()

	if err := w.syncHubSecrets(ctx, clusterClient, effectiveStack); err != nil {
		if isTransientClusterError(err) {
			w.Logger().Warnf("release %s: transient error syncing hub secrets, requeueing: %v", release.ID, err)
			return worker.Result{RequeueAfter: convergencePollInterval}, nil
		}
		return worker.Result{}, w.WorkerError.NewError("failed to sync hub secrets: %v", err)
	}

	if err := w.syncPostgresCredentialSecrets(ctx, clusterClient, effectiveStack); err != nil {
		if isTransientClusterError(err) {
			w.Logger().Warnf("release %s: transient error syncing postgres secrets, requeueing: %v", release.ID, err)
			return worker.Result{RequeueAfter: convergencePollInterval}, nil
		}
		return worker.Result{}, w.WorkerError.NewError("failed to sync postgres credential secrets: %v", err)
	}

	// Check volume readiness.
	if ready, err := w.volumesReady(ctx, release.StackID); err != nil {
		return worker.Result{}, w.WorkerError.NewError("failed to check volume readiness: %v", err)
	} else if !ready {
		w.Logger().Infof("release %s: volumes not ready, requeueing", release.ID)
		return worker.Result{RequeueAfter: convergencePollInterval}, nil
	}

	// Apply Stack CR.
	stackCR, err := w.applyStackCR(ctx, clusterClient, release)
	if err != nil {
		if isTransient(err) {
			w.Logger().Warnf("release %s: transient error applying stack CR, requeueing: %v", release.ID, err)
			return worker.Result{RequeueAfter: convergencePollInterval}, nil
		}
		w.fail(ctx, release, fmt.Sprintf("failed to apply stack CR: %v", err))
		return worker.Result{}, nil
	}

	// Apply StackResource CRs.
	if err := w.applyStackResourceCRs(ctx, clusterClient, release, stackCR); err != nil {
		if isTransient(err) {
			w.Logger().Warnf("release %s: transient error applying resource CRs, requeueing: %v", release.ID, err)
			return worker.Result{RequeueAfter: convergencePollInterval}, nil
		}
		w.fail(ctx, release, fmt.Sprintf("failed to apply stack resource CRs: %v", err))
		return worker.Result{}, nil
	}

	// Prune orphaned StackResource CRs.
	if err := w.pruneStackResources(ctx, clusterClient, stack, release.Manifest.ResourceNames); err != nil {
		w.Logger().Errorf("release %s: prune error (non-fatal): %v", release.ID, err)
	}

	// Check convergence.
	return w.checkConverged(ctx, release, stack)
}

func (w *releaseWorker) applyStackCR(ctx context.Context, clusterClient client.Client, release *models.StackRelease) (*corev1alpha1.Stack, error) {
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

	specChanged := !equality.Semantic.DeepEqual(existing.Spec, desired.Spec)
	annotationsChanged := !equality.Semantic.DeepEqual(existing.Annotations, desired.Annotations)
	labelsChanged := !equality.Semantic.DeepEqual(existing.Labels, desired.Labels)
	if specChanged || annotationsChanged || labelsChanged {
		desired.ResourceVersion = existing.ResourceVersion
		if err := clusterClient.Update(ctx, desired); err != nil {
			return nil, wrapIfTransient(fmt.Errorf("failed to update stack CR: %w", err))
		}
	}
	// Use existing UID for owner references.
	desired.UID = existing.UID
	return desired, nil
}

func (w *releaseWorker) applyStackResourceCRs(ctx context.Context, clusterClient client.Client, release *models.StackRelease, stackCR *corev1alpha1.Stack) error {
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
				continue
			}
			return wrapIfTransient(fmt.Errorf("failed to get StackResource CR '%s': %w", name, err))
		}

		specChanged := !equality.Semantic.DeepEqual(existing.Spec, desired.Spec)
		annotationsChanged := !equality.Semantic.DeepEqual(existing.Annotations, desired.Annotations)
		labelsChanged := !equality.Semantic.DeepEqual(existing.Labels, desired.Labels)
		ownerRefsChanged := !equality.Semantic.DeepEqual(existing.OwnerReferences, desired.OwnerReferences)
		if specChanged || annotationsChanged || labelsChanged || ownerRefsChanged {
			desired.ResourceVersion = existing.ResourceVersion
			if err := clusterClient.Update(ctx, desired); err != nil {
				return wrapIfTransient(fmt.Errorf("failed to update StackResource CR '%s': %w", name, err))
			}
		}
	}
	return nil
}

func (w *releaseWorker) pruneStackResources(ctx context.Context, clusterClient client.Client, stack *models.Stack, activeNames []string) error {
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
			w.Logger().Infof("pruning orphaned StackResource CR '%s'", srList.Items[i].Name)
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

func (w *releaseWorker) volumesReady(ctx context.Context, stackID string) (bool, error) {
	volumes, serr := w.volumeService.ListVolumesUsedByStack(ctx, stackID)
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
