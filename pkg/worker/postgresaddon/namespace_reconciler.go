package postgresaddon

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/worker"
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

func newNamespaceReconciler(spec PostgresAddonWorkerSpec) *namespaceReconciler {
	return &namespaceReconciler{
		clusterManager:   spec.ClusterManager,
		namespaceService: spec.NamespaceService,
		logger:           logger.NewLoggerWithPrefix(context.Background(), "postgres-addon-namespace"),
	}
}

func (r *namespaceReconciler) Name() string { return "namespace" }

func (r *namespaceReconciler) Reconcile(ctx context.Context, addon *models.PostgresAddon, authorizeMutation worker.MutationAuthorizer) (subReconcilerResult, error) {
	clusterClient, err := r.clusterManager.GetClient(addon.ClusterID)
	if err != nil {
		return resultNil, fmt.Errorf("failed to get cluster client: %w", err)
	}

	namespace, serr := r.namespaceService.Get(ctx, addon.NamespaceID)
	if serr != nil {
		return resultNil, fmt.Errorf("failed to get namespace: %w", serr)
	}

	existingNamespace := &corev1.Namespace{}
	if err := clusterClient.Get(ctx, client.ObjectKey{Name: namespace.Name}, existingNamespace); err != nil {
		if k8sapierrors.IsNotFound(err) {
			r.logger.Info(ctx, "Creating namespace '%s' in cluster", namespace.Name)
			if err := authorizeMutation(ctx); err != nil {
				return resultStop, err
			}
			return resultNil, clusterClient.Create(ctx, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:        namespace.Name,
					Labels:      namespace.Labels.ToMap(),
					Annotations: namespace.Annotations.ToMap(),
				},
			})
		}
		return resultNil, fmt.Errorf("failed to get namespace '%s': %w", namespace.Name, err)
	}

	mergedLabels, labelsChanged := mergeDesiredNamespaceMetadata(existingNamespace.Labels, namespace.Labels.ToMap())
	mergedAnnotations, annotationsChanged := mergeDesiredNamespaceMetadata(existingNamespace.Annotations, namespace.Annotations.ToMap())
	if !labelsChanged && !annotationsChanged {
		return resultNil, nil
	}

	existingNamespace.Labels = mergedLabels
	existingNamespace.Annotations = mergedAnnotations
	r.logger.Info(ctx, "Repairing metadata on namespace '%s'", namespace.Name)
	if err := authorizeMutation(ctx); err != nil {
		return resultStop, err
	}
	return resultNil, clusterClient.Update(ctx, existingNamespace)
}

func mergeDesiredNamespaceMetadata(existing, desired map[string]string) (map[string]string, bool) {
	merged := labels.Merge(existing, desired)
	return merged, !labels.Equals(existing, merged)
}
