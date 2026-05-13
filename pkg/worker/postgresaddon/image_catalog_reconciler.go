package postgresaddon

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/builders"
	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type imageCatalogReconciler struct {
	clusterManager clustermanager.ClusterManager
	logger         logger.Logger
}

func newImageCatalogReconciler(spec PostgresAddonWorkerSpec) *imageCatalogReconciler {
	return &imageCatalogReconciler{
		clusterManager: spec.ClusterManager,
		logger:         logger.NewLoggerWithPrefix(context.Background(), "postgres-addon-image-catalog"),
	}
}

func (r *imageCatalogReconciler) Name() string { return "image-catalog" }

func (r *imageCatalogReconciler) Reconcile(ctx context.Context, addon *models.PostgresAddon) (subReconcilerResult, error) {
	clusterClient, err := r.clusterManager.GetClient(addon.ClusterID)
	if err != nil {
		return resultNil, fmt.Errorf("failed to get cluster client: %w", err)
	}

	existing := &cnpgv1.ImageCatalog{}
	key := client.ObjectKey{Name: builders.DefaultImageCatalogName, Namespace: addon.Namespace}
	if err := clusterClient.Get(ctx, key, existing); err != nil {
		if k8sapierrors.IsNotFound(err) {
			r.logger.Infof("Creating ImageCatalog '%s' in namespace '%s'", builders.DefaultImageCatalogName, addon.Namespace)
			catalog := &cnpgv1.ImageCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name:      builders.DefaultImageCatalogName,
					Namespace: addon.Namespace,
				},
				Spec: cnpgv1.ImageCatalogSpec{
					Images: defaultPostgresImages(),
				},
			}
			return resultNil, clusterClient.Create(ctx, catalog)
		}
		return resultNil, fmt.Errorf("failed to check ImageCatalog: %w", err)
	}

	return resultNil, nil
}

func defaultPostgresImages() []cnpgv1.CatalogImage {
	return []cnpgv1.CatalogImage{
		{Major: 13, Image: "ghcr.io/cloudnative-pg/postgresql:13"},
		{Major: 14, Image: "ghcr.io/cloudnative-pg/postgresql:14"},
		{Major: 15, Image: "ghcr.io/cloudnative-pg/postgresql:15"},
		{Major: 16, Image: "ghcr.io/cloudnative-pg/postgresql:16"},
		{Major: 17, Image: "ghcr.io/cloudnative-pg/postgresql:17"},
	}
}
