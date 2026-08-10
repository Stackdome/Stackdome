package postgresaddon

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Postgres addon namespace reconciliation", func() {
	It("copies stored labels and annotations to a newly created namespace", func() {
		ctrl := gomock.NewController(GinkgoT())
		clusterManager := mocks.NewMockClusterManager(ctrl)
		namespaceService := NewMocknamespaceService(ctrl)
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		clusterClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		ctx := context.Background()
		namespace := &models.Namespace{
			ID:             "namespace-1",
			Name:           "stackdome-addons-postgres-database",
			OrganisationID: "organisation-1",
			Labels: models.Labels{
				{Key: models.ManagedByLabelKey, Value: models.ManagedByLabelValue},
				{Key: models.CloudTenantLabelKey, Value: models.CloudTenantLabelValue},
				{Key: models.OrganizationIDLabelKey, Value: "organisation-1"},
				{Key: models.NamespaceRoleLabelKey, Value: models.NamespaceRoleAddon},
			},
			Annotations: models.Annotations{{Key: "stackdome.io/test", Value: "annotation"}},
		}
		addon := &models.PostgresAddon{
			ClusterID:   "cluster-1",
			NamespaceID: namespace.ID,
		}
		clusterManager.EXPECT().GetClient(addon.ClusterID).Return(clusterClient, nil)
		namespaceService.EXPECT().Get(ctx, namespace.ID).Return(namespace, nil)
		reconciler := &namespaceReconciler{
			clusterManager:   clusterManager,
			namespaceService: namespaceService,
			logger:           logger.NewLoggerWithPrefix(ctx, "test"),
		}

		_, err := reconciler.Reconcile(ctx, addon)

		Expect(err).NotTo(HaveOccurred())
		created := &corev1.Namespace{}
		Expect(clusterClient.Get(ctx, client.ObjectKey{Name: namespace.Name}, created)).To(Succeed())
		Expect(created.Labels).To(Equal(namespace.Labels.ToMap()))
		Expect(created.Annotations).To(Equal(namespace.Annotations.ToMap()))
	})

	It("repairs stored labels and annotations on an existing namespace", func() {
		ctrl := gomock.NewController(GinkgoT())
		clusterManager := mocks.NewMockClusterManager(ctrl)
		namespaceService := NewMocknamespaceService(ctrl)
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		ctx := context.Background()
		namespace := &models.Namespace{
			ID:             "namespace-1",
			Name:           "stackdome-addons-postgres-database",
			OrganisationID: "organisation-1",
			Labels: models.Labels{
				{Key: models.ManagedByLabelKey, Value: models.ManagedByLabelValue},
				{Key: models.CloudTenantLabelKey, Value: models.CloudTenantLabelValue},
				{Key: models.OrganizationIDLabelKey, Value: "organisation-1"},
				{Key: models.NamespaceRoleLabelKey, Value: models.NamespaceRoleAddon},
			},
			Annotations: models.Annotations{{Key: "stackdome.io/test", Value: "expected"}},
		}
		existing := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:        namespace.Name,
				Labels:      map[string]string{"unrelated": "preserved"},
				Annotations: map[string]string{"stackdome.io/test": "stale"},
			},
		}
		clusterClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existing).Build()
		addon := &models.PostgresAddon{ClusterID: "cluster-1", NamespaceID: namespace.ID}
		clusterManager.EXPECT().GetClient(addon.ClusterID).Return(clusterClient, nil)
		namespaceService.EXPECT().Get(ctx, namespace.ID).Return(namespace, nil)
		reconciler := &namespaceReconciler{
			clusterManager:   clusterManager,
			namespaceService: namespaceService,
			logger:           logger.NewLoggerWithPrefix(ctx, "test"),
		}

		_, err := reconciler.Reconcile(ctx, addon)

		Expect(err).NotTo(HaveOccurred())
		updated := &corev1.Namespace{}
		Expect(clusterClient.Get(ctx, client.ObjectKey{Name: namespace.Name}, updated)).To(Succeed())
		Expect(updated.Labels).To(HaveKeyWithValue("unrelated", "preserved"))
		for key, value := range namespace.Labels.ToMap() {
			Expect(updated.Labels).To(HaveKeyWithValue(key, value))
		}
		Expect(updated.Annotations).To(HaveKeyWithValue("stackdome.io/test", "expected"))
	})
})
