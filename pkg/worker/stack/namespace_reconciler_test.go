package stack

import (
	"context"
	"testing"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/worker"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNamespaceReconcilerRepairsExistingNamespaceMetadata(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	clusterManager := mocks.NewMockClusterManager(ctrl)
	namespaceService := NewMocknamespaceService(ctrl)
	scheme := runtime.NewScheme()
	g.Expect(corev1.AddToScheme(scheme)).To(gomega.Succeed())
	ctx := context.Background()
	namespace := &models.Namespace{
		ID:             "namespace-1",
		Name:           "api-namespace",
		OrganisationID: "organisation-1",
		Labels: models.Labels{
			{Key: models.ManagedByLabelKey, Value: models.ManagedByLabelValue},
			{Key: models.CloudTenantLabelKey, Value: models.CloudTenantLabelValue},
			{Key: models.OrganizationIDLabelKey, Value: "organisation-1"},
			{Key: models.NamespaceRoleLabelKey, Value: models.NamespaceRoleStack},
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
	stack := &models.Stack{ID: "stack-1", ClusterID: "cluster-1", NamespaceID: namespace.ID}
	clusterManager.EXPECT().GetClient(stack.ClusterID).Return(clusterClient, nil)
	namespaceService.EXPECT().Get(ctx, namespace.ID).Return(namespace, nil)
	reconciler := NewNamespaceReconciler(NamespaceReconcilerSpec{
		ClusterManager:   clusterManager,
		NamespaceService: namespaceService,
		Logger:           logger.NewLoggerWithPrefix(ctx, "test"),
	})

	_, err := reconciler.Reconcile(ctx, stack, worker.AllowMutation)

	g.Expect(err).NotTo(gomega.HaveOccurred())
	updated := &corev1.Namespace{}
	g.Expect(clusterClient.Get(ctx, client.ObjectKey{Name: namespace.Name}, updated)).To(gomega.Succeed())
	g.Expect(updated.Labels).To(gomega.HaveKeyWithValue("unrelated", "preserved"))
	for key, value := range namespace.Labels.ToMap() {
		g.Expect(updated.Labels).To(gomega.HaveKeyWithValue(key, value))
	}
	g.Expect(updated.Annotations).To(gomega.HaveKeyWithValue("stackdome.io/test", "expected"))
}
