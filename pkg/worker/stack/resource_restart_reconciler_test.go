package stack

import (
	"context"
	"testing"
	"time"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

func TestResourceRestartReconcilerAppliesPersistedRequest(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	clusterManager := mocks.NewMockClusterManager(ctrl)
	scheme := runtime.NewScheme()
	g.Expect(corev1alpha1.AddToScheme(scheme)).To(gomega.Succeed())

	requestedAt := time.Date(2026, time.August, 10, 12, 0, 0, 0, time.UTC)
	resource := &models.StackResource{
		ID: "resource-1", Name: "api", Namespace: "stack-1",
		LifecycleConfig: &models.LifecycleConfig{RestartRequestTime: &requestedAt},
	}
	clusterResource := &corev1alpha1.StackResource{}
	clusterResource.Name = resource.Name
	clusterResource.Namespace = resource.Namespace
	clusterClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(clusterResource).Build()
	clusterManager.EXPECT().GetClient("cluster-1").Return(clusterClient, nil)

	reconciler := NewResourceRestartReconciler(ResourceRestartReconcilerSpec{
		ClusterManager: clusterManager,
		Logger:         logger.NewLoggerWithPrefix(context.Background(), "test"),
	})
	_, err := reconciler.Reconcile(context.Background(), &models.Stack{
		ID: "stack-1", ClusterID: "cluster-1", StackResources: []*models.StackResource{resource},
	})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	updated := &corev1alpha1.StackResource{}
	g.Expect(clusterClient.Get(context.Background(), client.ObjectKey{
		Name: resource.Name, Namespace: resource.Namespace,
	}, updated)).To(gomega.Succeed())
	g.Expect(updated.Spec.RestartRequest).NotTo(gomega.BeNil())
	g.Expect(updated.Spec.RestartRequest.Time).To(gomega.BeTemporally("==", requestedAt))
}
