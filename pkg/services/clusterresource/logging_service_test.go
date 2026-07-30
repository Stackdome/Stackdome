package clusterresource

import (
	"context"
	stderrors "errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/Stackdome/stackdome/pkg/clients"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
)

// stubLogParams satisfies the LoggingParams interface for tests.
type stubLogParams struct{}

func (stubLogParams) Follow() bool   { return false }
func (stubLogParams) TailLines() int { return 100 }
func (stubLogParams) Since() string  { return "" }

var _ = Describe("ClusterLoggingService.GetLogsForBuildPod", func() {
	const (
		orgID   = "org-1"
		ns      = "stack-ns"
		jobName = "web-abc1234-build"
	)
	var (
		ctrl    *gomock.Controller
		mockDB  *MockDBClusterService
		mockCM  *MockClusterManager
		mockK8s *MockKubernetesClient
		svc     ClusterLoggingService
		ctx     context.Context
		opts    LoggingParams
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockDB = NewMockDBClusterService(ctrl)
		mockCM = NewMockClusterManager(ctrl)
		mockK8s = NewMockKubernetesClient(ctrl)
		svc = NewLoggingService(LoggingServiceSpec{
			ClusterService: mockDB,
			ClusterManager: mockCM,
			Logger:         logger.NewLogger(),
			NewK8sClient: func(clients.KubernetesClientSpec) (clients.KubernetesClient, error) {
				return mockK8s, nil
			},
		})
		ctx = context.Background()
		opts = stubLogParams{}

		mockDB.EXPECT().GetClusterForOrg(ctx, orgID).Return(&models.Cluster{ID: "c1"}, nil)
		mockCM.EXPECT().GetRestConfig("c1").Return(&rest.Config{}, nil)
		mockCM.EXPECT().GetClient("c1").Return(nil, nil)
	})

	AfterEach(func() { ctrl.Finish() })

	It("returns ErrBuildPodNotFound when no pod matches the job", func() {
		mockK8s.EXPECT().BuildPodForJob(ctx, ns, jobName).Return(nil, nil)
		_, err := svc.GetLogsForBuildPod(ctx, orgID, ns, jobName, opts)
		Expect(err).To(HaveOccurred())
		Expect(stderrors.Is(err, ErrBuildPodNotFound)).To(BeTrue())
	})

	It("propagates the client error", func() {
		boom := stderrors.New("boom")
		mockK8s.EXPECT().BuildPodForJob(ctx, ns, jobName).Return(nil, boom)
		_, err := svc.GetLogsForBuildPod(ctx, orgID, ns, jobName, opts)
		Expect(err).To(HaveOccurred())
		Expect(stderrors.Is(err, boom)).To(BeTrue())
	})

	It("returns ErrBuildPodNotReady when the pod is still pending", func() {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: jobName + "-xyz", Namespace: ns},
			Status:     corev1.PodStatus{Phase: corev1.PodPending},
		}
		mockK8s.EXPECT().BuildPodForJob(ctx, ns, jobName).Return(pod, nil)
		_, err := svc.GetLogsForBuildPod(ctx, orgID, ns, jobName, opts)
		Expect(stderrors.Is(err, ErrBuildPodNotReady)).To(BeTrue())
	})

	It("returns a streamer for the resolved build pod", func() {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: jobName + "-xyz", Namespace: ns},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		}
		mockK8s.EXPECT().BuildPodForJob(ctx, ns, jobName).Return(pod, nil)
		streamer, err := svc.GetLogsForBuildPod(ctx, orgID, ns, jobName, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(streamer).NotTo(BeNil())
	})
})
