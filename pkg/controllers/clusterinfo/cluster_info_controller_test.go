package clusterinfo

import (
	"context"
	"testing"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

func TestClusterInfoController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ClusterInfo Controller Suite")
}

var _ = Describe("clusterInfoReconciler", func() {
	It("persists the storage classes reported by the agent", func() {
		scheme := runtime.NewScheme()
		Expect(corev1alpha1.AddToScheme(scheme)).To(Succeed())

		cr := &corev1alpha1.ClusterInfo{
			ObjectMeta: metav1.ObjectMeta{Name: corev1alpha1.ClusterInfoSingletonName},
			Status: corev1alpha1.ClusterInfoStatus{
				Phase:             corev1alpha1.ClusterInfoPhaseReady,
				KubernetesVersion: "v1.31.2",
				Nodes: []corev1alpha1.NodeInfo{
					{Name: "node-1", Ready: true, AllocatableCPU: "3800m", AllocatableMemory: "7168Mi"},
				},
				StorageClasses: []corev1alpha1.StorageClassInfo{
					{Name: "local-path", Provisioner: "rancher.io/local-path", IsDefault: true},
				},
			},
		}

		ctrlMock := gomock.NewController(GinkgoT())
		defer ctrlMock.Finish()
		clusterService := mocks.NewMockClusterService(ctrlMock)
		clusterService.EXPECT().
			InternalUpdateClusterInfo(gomock.Any(), "cluster-1", gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string, info *models.ClusterInfo) *errors.ServiceError {
				Expect(info.DefaultStorageClass()).To(Equal("local-path"))
				Expect(info.KubernetesVersion).To(Equal("v1.31.2"))
				Expect(info.Nodes[0].AllocatableCPU.MilliValue()).To(Equal(int64(3800)))
				return nil
			})

		r := &clusterInfoReconciler{
			Client:         fake.NewClientBuilder().WithScheme(scheme).WithObjects(cr).Build(),
			ClusterID:      "cluster-1",
			ClusterService: clusterService,
			Log:            logger.NewLoggerWithPrefix(context.Background(), "clusterinfo-test"),
		}

		_, err := r.Reconcile(context.Background(), ctrl.Request{
			NamespacedName: types.NamespacedName{Name: corev1alpha1.ClusterInfoSingletonName},
		})
		Expect(err).ToNot(HaveOccurred())
	})


	It("returns an empty result without error when the CR is not found", func() {
		scheme := runtime.NewScheme()
		Expect(corev1alpha1.AddToScheme(scheme)).To(Succeed())

		r := &clusterInfoReconciler{
			Client:         fake.NewClientBuilder().WithScheme(scheme).Build(),
			ClusterID:      "cluster-1",
			ClusterService: mocks.NewMockClusterService(gomock.NewController(GinkgoT())),
			Log:            logger.NewLoggerWithPrefix(context.Background(), "clusterinfo-test"),
		}

		result, err := r.Reconcile(context.Background(), ctrl.Request{
			NamespacedName: types.NamespacedName{Name: corev1alpha1.ClusterInfoSingletonName},
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{}))
	})
})

var _ = Describe("mapClusterInfoPhase", func() {
	It("maps a recognised agent phase", func() {
		Expect(mapClusterInfoPhase(corev1alpha1.ClusterInfoPhaseReady)).To(Equal(models.ClusterInfoPhaseReady))
	})

	It("maps an unrecognised agent phase to unknown", func() {
		Expect(mapClusterInfoPhase(corev1alpha1.ClusterInfoPhase("SomethingNew"))).To(Equal(models.ClusterInfoPhaseUnknown))
	})
})

var _ = Describe("clusterInfoReconciler.parseQuantity", func() {
	var r *clusterInfoReconciler

	BeforeEach(func() {
		r = &clusterInfoReconciler{
			Log: logger.NewLoggerWithPrefix(context.Background(), "clusterinfo-test"),
		}
	})

	It("returns nil for an empty value", func() {
		Expect(r.parseQuantity(context.Background(), "node-1", "")).To(BeNil())
	})

	It("returns nil for an unparseable value without failing", func() {
		Expect(r.parseQuantity(context.Background(), "node-1", "not-a-quantity")).To(BeNil())
	})
})
