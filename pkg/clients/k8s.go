package clients

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/util/podutils"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

type MetricsStreamOptions interface {
	// PollInterval specifies how often to poll for metrics updates.
	PollInterval() time.Duration
	// MaxErrors returns the maximum number of errors to log before stopping the stream.
	MaxErrors() int
}

type PodLogStreamOptions interface {
	Follow() bool
	TailLines() int
	Since() string
	// MaxErrors returns the maximum number of errors to log before stopping the stream.
	MaxErrors() int
}

type KubernetesClient interface {
	AttachablePodFromService(ctx context.Context, svcKey types.NamespacedName) (*corev1.Pod, error)
	StreamPodLogs(ctx context.Context, pod *corev1.Pod, logOpts PodLogStreamOptions) (<-chan *LogStreamObject, error)
	BuildPodForJob(ctx context.Context, namespace string, jobName string) (*corev1.Pod, error)

	GetNamespaceMetrics(ctx context.Context, namespace string) (*MetricsStreamObject, error)
	StreamNamespaceMetrics(ctx context.Context, namespace string, metricsStreamOpts MetricsStreamOptions) (<-chan *MetricsStreamObject, error)
}

type kubernetesClient struct {
	clientSet               *kubernetes.Clientset
	controllerRuntimeClient client.Client
	logger                  logger.Logger
	restConfig              *rest.Config
}

type KubernetesClientSpec struct {
	RestConfig              *rest.Config
	ControllerRuntimeClient client.Client
	Logger                  logger.Logger
}

func NewKubernetesClient(spec KubernetesClientSpec) (KubernetesClient, error) {
	if spec.RestConfig == nil {
		return nil, fmt.Errorf("rest config cannot be nil")
	}
	if spec.ControllerRuntimeClient == nil {
		return nil, fmt.Errorf("controller runtime client cannot be nil")
	}

	if spec.Logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	clientset, err := kubernetes.NewForConfig(spec.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}
	return &kubernetesClient{
		clientSet:               clientset,
		controllerRuntimeClient: spec.ControllerRuntimeClient,
		logger:                  spec.Logger,
		restConfig:              spec.RestConfig,
	}, nil
}

func (k *kubernetesClient) AttachablePodFromService(ctx context.Context, svcKey types.NamespacedName) (*corev1.Pod, error) {
	svc, err := k.clientSet.CoreV1().Services(svcKey.Namespace).Get(ctx, svcKey.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	attachablePod, err := k.attachablePodForObject(svc, time.Second*10)
	if err != nil {
		return nil, err
	}
	return attachablePod, nil
}

func (k *kubernetesClient) StreamPodLogs(ctx context.Context, pod *corev1.Pod, logOpts PodLogStreamOptions) (<-chan *LogStreamObject, error) {
	logOptions := &corev1.PodLogOptions{}
	if logOpts.Since() != "" {
		sinceTime, err := time.ParseDuration(logOpts.Since())
		if err != nil {
			k.logger.Warn(ctx, "invalid since time format: %v. Ignoring set log since options", err)
		} else {
			logOptions.SinceSeconds = ptr.To(int64(sinceTime.Seconds()))
		}
	}
	logOptions.Follow = logOpts.Follow()
	logOptions.TailLines = ptr.To(int64(logOpts.TailLines()))

	req := k.clientSet.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOptions)
	stream, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs for pod %s/%s: %w", pod.Namespace, pod.Name, err)
	}
	reader := bufio.NewReader(stream)
	logChannel := make(chan *LogStreamObject)
	go func() {
		defer close(logChannel)
		defer func() { _ = stream.Close() }()
		numStreamErrors := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			line, err := reader.ReadString('\n')
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return
				}
				if err == io.EOF {
					k.logger.Debug(ctx, "log stream for pod %s/%s ended", pod.Namespace, pod.Name)
					return
				}
				if numStreamErrors >= logOpts.MaxErrors() {
					k.logger.Warn(ctx, "maximum number of stream errors reached (%d), stopping log stream", logOpts.MaxErrors())
					safeWriteToChannel(ctx, logChannel, &LogStreamObject{Error: fmt.Errorf("maximum number of stream errors reached")})
					return
				}
				k.logger.Error(ctx, "error reading log line: %v", err)
				safeWriteToChannel(ctx, logChannel, &LogStreamObject{Error: fmt.Errorf("error reading log line: %w", err)})
				numStreamErrors++
				continue
			}
			if line != "" {
				safeWriteToChannel(ctx, logChannel, &LogStreamObject{Data: line})
			}
		}
	}()
	return logChannel, nil
}

func (k *kubernetesClient) GetNamespaceMetrics(ctx context.Context, namespace string) (*MetricsStreamObject, error) {
	metricsClient, err := metrics.NewForConfig(k.restConfig)
	if err != nil {
		return nil, err
	}

	nodeCapacityMap, err := k.getNodeCapacityMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node capacity map: %w", err)
	}

	podMetricsList, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	//  add assigned node as an annotation to each pod metric
	for i := range podMetricsList.Items {
		podMetrics := &podMetricsList.Items[i]
		pod, err := k.clientSet.CoreV1().Pods(namespace).Get(ctx, podMetrics.Name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get pod %s/%s: %w", namespace, podMetrics.Name, err)
		}
		if pod.Spec.NodeName != "" {
			if podMetrics.Annotations == nil {
				podMetrics.Annotations = make(map[string]string)
			}
			podMetrics.Annotations[models.AssignedNodeAnnotation] = pod.Spec.NodeName
		}
	}

	return &MetricsStreamObject{
		NamespaceMetrics: podMetricsList.Items,
		NodeCapacityMap:  nodeCapacityMap,
	}, nil
}

func (k *kubernetesClient) StreamNamespaceMetrics(ctx context.Context, namespace string, metricsStreamOpts MetricsStreamOptions) (<-chan *MetricsStreamObject, error) {
	metricsClient, err := metrics.NewForConfig(k.restConfig)
	if err != nil {
		return nil, err
	}

	// Hit the api first to ensure the metrics server is available
	_, err = metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		if k8sapierrors.IsNotFound(err) {
			return nil, fmt.Errorf("metrics server not installed: %w", err)
		}
		if k8sapierrors.IsServiceUnavailable(err) {
			return nil, fmt.Errorf("metrics server is unavailable: %w", err)
		}
		return nil, err
	}

	streamChannel := make(chan *MetricsStreamObject)
	k.logger.Info(ctx, "starting metrics stream for namespace %s with poll interval %s", namespace, metricsStreamOpts.PollInterval().String())
	pollingTicker := time.NewTicker(metricsStreamOpts.PollInterval())

	ticker := make(chan struct{})
	firstTickChan := make(chan struct{}, 1)
	go func() {
		defer close(ticker)
		defer close(firstTickChan)
		defer pollingTicker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-pollingTicker.C:
				ticker <- struct{}{}
			case <-firstTickChan:
				ticker <- struct{}{}
			}
		}
	}()

	firstTickChan <- struct{}{} // Trigger the first tick immediately
	go func() {
		defer close(streamChannel)
		numStreamErrors := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker:
				podMetricsList, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
				if err != nil {
					if numStreamErrors >= metricsStreamOpts.MaxErrors() {
						k.logger.Warn(ctx, "maximum number of stream errors reached (%d), stopping metrics stream", metricsStreamOpts.MaxErrors())
						safeWriteToChannel(ctx, streamChannel, &MetricsStreamObject{Error: fmt.Errorf("maximum number of stream errors reached")})
						return
					}
					k.logger.Error(ctx, "error fetching pod metrics: %v", err)
					safeWriteToChannel(ctx, streamChannel, &MetricsStreamObject{Error: fmt.Errorf("error fetching pod metrics: %w", err)})
					numStreamErrors++
					continue
				}
				safeWriteToChannel(ctx, streamChannel, &MetricsStreamObject{
					NamespaceMetrics: podMetricsList.Items,
				})
			}
		}
	}()
	return streamChannel, nil
}

func (k *kubernetesClient) BuildPodForJob(ctx context.Context, namespace string, jobName string) (*corev1.Pod, error) {
	podList, err := k.clientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", batchv1.JobNameLabel, jobName),
	})
	if err != nil {
		return nil, err
	}
	return mostRecentPod(podList.Items), nil
}

// mostRecentPod returns the pod with the latest creation timestamp, or nil when
// the slice is empty. A build Job may produce several pods across retries
// (BackoffLimit: 3); the newest is the active/last attempt.
func mostRecentPod(pods []corev1.Pod) *corev1.Pod {
	if len(pods) == 0 {
		return nil
	}
	latest := &pods[0]
	for i := range pods {
		if pods[i].CreationTimestamp.Time.After(latest.CreationTimestamp.Time) {
			latest = &pods[i]
		}
	}
	return latest
}

func (k *kubernetesClient) attachablePodForObject(object runtime.Object, timeout time.Duration) (*corev1.Pod, error) {
	namespace, selector, err := polymorphichelpers.SelectorsForObject(object)
	if err != nil {
		return nil, fmt.Errorf("cannot attach to %T: %w", object, err)
	}
	sortBy := func(pods []*corev1.Pod) sort.Interface { return sort.Reverse(podutils.ActivePods(pods)) }
	pod, _, err := polymorphichelpers.GetFirstPod(k.clientSet.CoreV1(), namespace, selector.String(), timeout, sortBy)
	return pod, err
}

func (k *kubernetesClient) getNodeCapacityMap(ctx context.Context) (map[string]corev1.ResourceList, error) {
	nodes, err := k.clientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodeCapacityMap := make(map[string]corev1.ResourceList)
	for _, node := range nodes.Items {
		nodeCapacityMap[node.Name] = node.Status.Capacity
	}
	return nodeCapacityMap, nil
}

func safeWriteToChannel[T any](ctx context.Context, ch chan<- *T, obj *T) {
	select {
	case <-ctx.Done():
		return
	case ch <- obj:
	}
}
