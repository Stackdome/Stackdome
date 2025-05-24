package clients

import (
	"bufio"
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/util/podutils"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type LogLine struct {
	Data  string `json:"data"`
	Error error  `json:"error"`
}

type PodLogOptions interface {
	// Follow indicates whether to follow the log stream.
	Follow() bool
	// TailLines specifies the number of lines to tail from the end of the log.
	TailLines() int
	// Since specifies the time from which to start streaming logs.
	Since() string
}

type KubernetesClient interface {
	// AttachablePodFromService retrieves an attachable pod for a given service.
	AttachablePodFromService(ctx context.Context, svcKey types.NamespacedName) (*corev1.Pod, error)
	StreamPodLogs(ctx context.Context, pod *corev1.Pod, logOpts PodLogOptions) (<-chan *LogLine, error)
}

type kubernetesClient struct {
	clientSet               *kubernetes.Clientset
	controllerRuntimeClient client.Client
	logger                  logger.Logger
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

func (k *kubernetesClient) StreamPodLogs(ctx context.Context, pod *corev1.Pod, logOpts PodLogOptions) (<-chan *LogLine, error) {
	logOptions := &corev1.PodLogOptions{}
	if logOpts.Since() != "" {
		sinceTime, err := time.ParseDuration(logOpts.Since())
		if err != nil {
			k.logger.Warnf("invalid since time format: %v. Ignoring set log since options", err)
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
	logChannel := make(chan *LogLine)
	go func() {
		defer close(logChannel)
		defer stream.Close()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					return
				}
				k.logger.Errorf("error reading log line: %v", err)
				safeWriteToChannel(ctx, logChannel, &LogLine{Error: fmt.Errorf("error reading log line: %v", err)})
			}
			if line != "" {
				safeWriteToChannel(ctx, logChannel, &LogLine{Data: line})
			}
		}
	}()
	return logChannel, nil
}

func (k *kubernetesClient) attachablePodForObject(object runtime.Object, timeout time.Duration) (*corev1.Pod, error) {
	namespace, selector, err := polymorphichelpers.SelectorsForObject(object)
	if err != nil {
		return nil, fmt.Errorf("cannot attach to %T: %v", object, err)
	}
	sortBy := func(pods []*corev1.Pod) sort.Interface { return sort.Reverse(podutils.ActivePods(pods)) }
	pod, _, err := polymorphichelpers.GetFirstPod(k.clientSet.CoreV1(), namespace, selector.String(), timeout, sortBy)
	return pod, err
}

func safeWriteToChannel(ctx context.Context, ch chan<- *LogLine, obj *LogLine) {
	select {
	case <-ctx.Done():
		return
	case ch <- obj:
	}
}
