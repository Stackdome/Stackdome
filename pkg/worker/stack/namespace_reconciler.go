package stack

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type namespaceReconciler struct {
	clusterManager   clustermanager.ClusterManager
	namespaceService namespaceService
	logger           logger.Logger
}

type NamespaceReconcilerSpec struct {
	ClusterManager   clustermanager.ClusterManager
	NamespaceService namespaceService
	Logger           logger.Logger
}

func NewNamespaceReconciler(spec NamespaceReconcilerSpec) *namespaceReconciler {
	return &namespaceReconciler{
		clusterManager:   spec.ClusterManager,
		namespaceService: spec.NamespaceService,
		logger:           spec.Logger,
	}
}

func (r *namespaceReconciler) Reconcile(ctx context.Context, stack *models.Stack) (subReconcilerResult, error) {
	log := r.logger.WithFields(map[string]interface{}{
		logger.FieldStackID:   stack.ID,
		logger.FieldClusterID: stack.ClusterID,
	})

	clusterClient, cerr := r.clusterManager.GetClient(stack.ClusterID)
	if cerr != nil {
		return resultNil, fmt.Errorf("failed to get cluster client for cluster %s: %w", stack.ClusterID, cerr)
	}

	namespace, err := r.namespaceService.Get(ctx, stack.NamespaceID)
	if err != nil {
		return resultNil, err
	}

	desiredNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        namespace.Name,
			Labels:      labelsToMap(namespace.Labels),
			Annotations: annotationsToMap(namespace.Annotations),
		},
	}

	existingNamespace := &corev1.Namespace{}
	if err := clusterClient.Get(ctx, client.ObjectKey{Name: namespace.Name}, existingNamespace); err != nil {
		if k8sapierrors.IsNotFound(err) {
			log.Info(ctx, "creating namespace %s in cluster", namespace.Name)
			return resultNil, clusterClient.Create(ctx, desiredNamespace)
		}
		return resultNil, fmt.Errorf("failed to get namespace %s: %w", namespace.Name, err)
	}

	mergedLabels, labelsChanged := mergeDesiredMetadata(existingNamespace.Labels, desiredNamespace.Labels)
	mergedAnnotations, annotationsChanged := mergeDesiredMetadata(existingNamespace.Annotations, desiredNamespace.Annotations)
	if !labelsChanged && !annotationsChanged {
		log.Debug(ctx, "namespace %s already has the desired metadata", namespace.Name)
		return resultNil, nil
	}

	existingNamespace.Labels = mergedLabels
	existingNamespace.Annotations = mergedAnnotations
	log.Info(ctx, "repairing metadata on namespace %s", namespace.Name)
	return resultNil, clusterClient.Update(ctx, existingNamespace)
}

func (r *namespaceReconciler) Name() string {
	return "namespace-reconciler"
}

func mergeDesiredMetadata(existing, desired map[string]string) (map[string]string, bool) {
	merged := labels.Merge(existing, desired)
	return merged, !labels.Equals(existing, merged)
}

func labelsToMap(labels models.Labels) map[string]string {
	m := make(map[string]string)
	for _, l := range labels {
		m[l.Key] = l.Value
	}
	return m
}

func annotationsToMap(annotations models.Annotations) map[string]string {
	m := make(map[string]string)
	for _, a := range annotations {
		m[a.Key] = a.Value
	}
	return m
}
