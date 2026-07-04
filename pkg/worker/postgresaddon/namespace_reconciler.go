package postgresaddon

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (r *namespaceReconciler) Reconcile(ctx context.Context, addon *models.PostgresAddon) (subReconcilerResult, error) {
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
			r.logger.Infof("Creating namespace '%s' in cluster", namespace.Name)
			return resultNil, clusterClient.Create(ctx, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{Name: namespace.Name},
			})
		}
		return resultNil, fmt.Errorf("failed to get namespace '%s': %w", namespace.Name, err)
	}

	return resultNil, nil
}
