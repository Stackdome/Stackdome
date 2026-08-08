package clusterresource

import (
	"context"
	stderrors "errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"golang.org/x/time/rate"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"

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
		orgID     = "org-1"
		ns        = "stack-ns"
		buildName = "web-abc1234"
	)
	var (
		ctrl     *gomock.Controller
		mockDB   *MockDBClusterService
		mockCM   *MockClusterManager
		mockK8s  *MockKubernetesClient
		svc      ClusterLoggingService
		ctx      context.Context
		opts     LoggingParams
		jobName  string
		crClient client.Client
	)

	buildCR := func() *buildsv1alpha1.ImageBuild {
		return &buildsv1alpha1.ImageBuild{
			ObjectMeta: metav1.ObjectMeta{Name: buildName, Namespace: ns},
			Spec:       buildsv1alpha1.ImageBuildSpec{ResourceName: "web"},
		}
	}

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		Expect(buildsv1alpha1.AddToScheme(scheme)).To(Succeed())
		crClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(buildCR()).Build()
		jobName = buildCR().BuildJobName()

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
		mockCM.EXPECT().GetClient("c1").DoAndReturn(func(string) (client.Client, error) { return crClient, nil })
	})

	AfterEach(func() { ctrl.Finish() })

	It("returns ErrBuildPodNotFound when the ImageBuild CR is gone", func() {
		crClient = fake.NewClientBuilder().WithScheme(crClient.Scheme()).Build()
		_, err := svc.GetLogsForBuildPod(ctx, orgID, ns, buildName, opts)
		Expect(stderrors.Is(err, ErrBuildPodNotFound)).To(BeTrue())
	})

	It("propagates a non-NotFound error from the CR lookup", func() {
		boom := stderrors.New("apiserver unavailable")
		crClient = fake.NewClientBuilder().WithScheme(crClient.Scheme()).WithObjects(buildCR()).
			WithInterceptorFuncs(interceptor.Funcs{
				Get: func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
					return boom
				},
			}).Build()
		_, err := svc.GetLogsForBuildPod(ctx, orgID, ns, buildName, opts)
		Expect(stderrors.Is(err, boom)).To(BeTrue())
		Expect(stderrors.Is(err, ErrBuildPodNotFound)).To(BeFalse())
	})

	It("returns ErrBuildPodNotFound when no pod matches the job", func() {
		mockK8s.EXPECT().BuildPodForJob(ctx, ns, jobName).Return(nil, nil)
		_, err := svc.GetLogsForBuildPod(ctx, orgID, ns, buildName, opts)
		Expect(err).To(HaveOccurred())
		Expect(stderrors.Is(err, ErrBuildPodNotFound)).To(BeTrue())
	})

	It("propagates the client error", func() {
		boom := stderrors.New("boom")
		mockK8s.EXPECT().BuildPodForJob(ctx, ns, jobName).Return(nil, boom)
		_, err := svc.GetLogsForBuildPod(ctx, orgID, ns, buildName, opts)
		Expect(err).To(HaveOccurred())
		Expect(stderrors.Is(err, boom)).To(BeTrue())
	})

	It("returns ErrBuildPodNotReady when the pod is still pending", func() {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: jobName + "-xyz", Namespace: ns},
			Status:     corev1.PodStatus{Phase: corev1.PodPending},
		}
		mockK8s.EXPECT().BuildPodForJob(ctx, ns, jobName).Return(pod, nil)
		_, err := svc.GetLogsForBuildPod(ctx, orgID, ns, buildName, opts)
		Expect(stderrors.Is(err, ErrBuildPodNotReady)).To(BeTrue())
	})

	It("returns a streamer for the resolved build pod", func() {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: jobName + "-xyz", Namespace: ns},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		}
		mockK8s.EXPECT().BuildPodForJob(ctx, ns, jobName).Return(pod, nil)
		streamer, err := svc.GetLogsForBuildPod(ctx, orgID, ns, buildName, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(streamer).NotTo(BeNil())
	})
})

var _ = Describe("ClusterLoggingService.GetLogsForResources", func() {
	const (
		orgID = "org-1"
		ns    = "stack-ns"
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

	resource := func(state models.StackResourceState, serviceName *string) *models.StackResource {
		return &models.StackResource{
			Name:      "web",
			Namespace: ns,
			Status: &models.StackResourceStatus{
				State:               state,
				InternalServiceName: serviceName,
			},
		}
	}

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

	It("serves logs for a crash-looping resource that is not Ready", func() {
		crashing := resource(models.StackResourcePhaseFailed, ptr.To("web-svc"))
		mockK8s.EXPECT().
			AttachablePodFromService(ctx, types.NamespacedName{Name: "web-svc", Namespace: ns}).
			Return(&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "web-1", Namespace: ns}}, nil)

		streamer, err := svc.GetLogsForResources(ctx, orgID, []*models.StackResource{crashing}, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(streamer).NotTo(BeNil())
	})

	It("returns ErrNoWorkload when the resource has no internal service yet", func() {
		pending := resource(models.StackResourcePhasePending, nil)
		_, err := svc.GetLogsForResources(ctx, orgID, []*models.StackResource{pending}, opts)
		Expect(stderrors.Is(err, ErrNoWorkload)).To(BeTrue())
	})

	It("returns ErrNoWorkload when the pod is gone", func() {
		crashing := resource(models.StackResourcePhaseFailed, ptr.To("web-svc"))
		mockK8s.EXPECT().AttachablePodFromService(ctx, gomock.Any()).Return(nil, nil)

		_, err := svc.GetLogsForResources(ctx, orgID, []*models.StackResource{crashing}, opts)
		Expect(stderrors.Is(err, ErrNoWorkload)).To(BeTrue())
	})
})

var _ = Describe("LogStreamer previous-container fallback", func() {
	var (
		ctrl     *gomock.Controller
		mockK8s  *MockKubernetesClient
		streamer *LogStreamer
		pod      *corev1.Pod
		ctx      context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockK8s = NewMockKubernetesClient(ctrl)
		pod = &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "web-1", Namespace: "stack-ns"}}
		streamer = &LogStreamer{
			k8sclient:      mockK8s,
			resourcePodMap: map[string]*corev1.Pod{"web": pod},
			Logger:         logger.NewLogger(),
			streamConfig: LogStreamConfig{
				logStreamBufferSize:   DefaultLogStreamBufferSize,
				streamTimeoutDuration: DefaultStreamTimeoutDuration,
				rateLimiter:           rate.NewLimiter(rate.Every(DefaultLogStreamRateLimit), DefaultLogStreamRateLimitBurst),
				logOptions:            stubLogParams{},
				maxLogStreamErrors:    MaxLogStreamErrors,
			},
		}
		ctx = context.Background()
	})

	AfterEach(func() { ctrl.Finish() })

	It("retries with the previous container when the current one has no logs", func() {
		restarting := stderrors.New("container is waiting to start")
		mockK8s.EXPECT().
			StreamPodLogs(gomock.Any(), pod, gomock.Cond(func(o clients.PodLogStreamOptions) bool { return !o.Previous() })).
			Return(nil, restarting)
		mockK8s.EXPECT().
			StreamPodLogs(gomock.Any(), pod, gomock.Cond(func(o clients.PodLogStreamOptions) bool { return o.Previous() })).
			Return(make(chan *clients.LogStreamObject), nil)

		out, err := streamer.Stream(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).NotTo(BeNil())
	})

	It("reports the original error when the previous container has no logs either", func() {
		boom := stderrors.New("container is waiting to start")
		mockK8s.EXPECT().StreamPodLogs(gomock.Any(), pod, gomock.Any()).Return(nil, boom).Times(2)

		_, err := streamer.Stream(ctx)
		Expect(stderrors.Is(err, boom)).To(BeTrue())
	})
})
