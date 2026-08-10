package clusterresource

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	registryv1alpha1 "stackdome.io/cluster-agent/api/registry/v1alpha1"
)

var _ = Describe("Cluster image registry resources", func() {
	var (
		ctx            context.Context
		clusterManager *MockClusterManager
		clusterClient  client.Client
		service        ClusterImageRegistryService
		registry       *models.ClusterImageRegistry
	)

	BeforeEach(func() {
		ctx = context.Background()
		clusterManager = NewMockClusterManager(gomock.NewController(GinkgoT()))
		scheme := runtime.NewScheme()
		Expect(registryv1alpha1.AddToScheme(scheme)).To(Succeed())
		clusterClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		service = NewClusterImageRegistryService(ClusterImageRegistryServiceSpec{
			ClusterManager: clusterManager,
			Logger:         logger.NewLogger(),
		})
		registry = &models.ClusterImageRegistry{
			ID:                  "registry-1",
			ClusterID:           "cluster-1",
			OrganisationID:      "org-1",
			Name:                "registry-1",
			BackendStorageSize:  "10Gi",
			BackendStorageClass: "fast",
		}
	})

	expectClient := func() {
		clusterManager.EXPECT().GetClient(registry.ClusterID).Return(clusterClient, nil)
	}

	It("creates a missing registry and updates managed spec drift", func() {
		expectClient()
		Expect(service.EnsureImageRegistryInCluster(ctx, registry)).To(BeNil())

		existing := &registryv1alpha1.ClusterRegistry{}
		Expect(clusterClient.Get(ctx, client.ObjectKey{Name: registry.Name}, existing)).To(Succeed())
		Expect(existing.Labels[models.ImageRegistryIDLabel]).To(Equal(registry.ID))
		Expect(existing.Spec.Storage.Size).To(Equal("10Gi"))

		registry.BackendStorageSize = "20Gi"
		expectClient()
		Expect(service.EnsureImageRegistryInCluster(ctx, registry)).To(BeNil())
		Expect(clusterClient.Get(ctx, client.ObjectKey{Name: registry.Name}, existing)).To(Succeed())
		Expect(existing.Spec.Storage.Size).To(Equal("20Gi"))
	})

	It("refuses to manage a same-name registry with a different ID", func() {
		foreign := desiredImageRegistryObject(registry)
		foreign.Labels[models.ImageRegistryIDLabel] = "registry-2"
		Expect(clusterClient.Create(ctx, foreign)).To(Succeed())

		expectClient()
		Expect(service.EnsureImageRegistryInCluster(ctx, registry)).ToNot(BeNil())
	})

	It("deletes a matching registry and reports absence on the following pass", func() {
		Expect(clusterClient.Create(ctx, desiredImageRegistryObject(registry))).To(Succeed())

		expectClient()
		absent, resourceErr := service.DeleteImageRegistryInCluster(ctx, registry)
		Expect(resourceErr).To(BeNil())
		Expect(absent).To(BeFalse())

		expectClient()
		absent, resourceErr = service.DeleteImageRegistryInCluster(ctx, registry)
		Expect(resourceErr).To(BeNil())
		Expect(absent).To(BeTrue())
	})
})
