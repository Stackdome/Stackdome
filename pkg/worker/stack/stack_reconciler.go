package stack

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/builders"
	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"k8s.io/apimachinery/pkg/api/equality"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

type stackReconciler struct {
	clusterManager clustermanager.ClusterManager
	stackService   stackService
	stackCRbuilder builders.ClusterResourceBuilder
}

type StackReconcilerSpec struct {
	ClusterManager clustermanager.ClusterManager
	StackService   stackService
	StackCRBuilder builders.ClusterResourceBuilder
}

func NewStackReconciler(spec StackReconcilerSpec) *stackReconciler {
	return &stackReconciler{
		clusterManager: spec.ClusterManager,
		stackService:   spec.StackService,
		stackCRbuilder: spec.StackCRBuilder,
	}
}
func (r *stackReconciler) Reconcile(ctx context.Context, stack *models.Stack) (subReconcilerResult, error) {
	clusterClient, cerr := r.clusterManager.GetClient(stack.ClusterID)
	if cerr != nil {
		return resultNil, cerr
	}

	desiredStackCR, err := r.stackCRbuilder.BuildStackCR(stack)
	if err != nil {
		return resultNil, err
	}

	// Ensure the stack CR has the revision annotation set
	if desiredStackCR.Annotations == nil {
		desiredStackCR.Annotations = make(map[string]string)
		desiredStackCR.Annotations[corev1alpha1.StackdomeServerObjectRevisionAnnotationKey] = stack.CrRevision
	}

	existingStackCR := corev1alpha1.Stack{}
	if err := clusterClient.Get(ctx, client.ObjectKey{Name: desiredStackCR.Name, Namespace: desiredStackCR.Namespace}, &existingStackCR); err != nil {
		if k8sapierrors.IsNotFound(err) {
			return resultNil, clusterClient.Create(ctx, desiredStackCR)
		}
		return resultNil, err
	}
	if !equality.Semantic.DeepEqual(existingStackCR.Spec, desiredStackCR.Spec) {
		desiredStackCR.ResourceVersion = existingStackCR.ResourceVersion
		return resultNil, clusterClient.Update(ctx, desiredStackCR)
	}
	return resultNil, nil
}

func (r *stackReconciler) Name() string {
	return "stack-reconciler"
}
