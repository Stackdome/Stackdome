package clusterresource

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"sync"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/interfaces"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/util/podutils"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterLoggingService interface {
	// GetLogsForResources retrieves logs for the specified resources in the cluster.
	GetLogsForResources(ctx context.Context, orgID string, resources []*models.StackResource, options LoggingParams) (interfaces.ServerSideStreamable, error)
}

type LoggingParams interface {
	// Follow indicates whether to follow the log stream.
	Follow() bool
	// TailLines specifies the number of lines to tail from the end of the log.
	TailLines() int
	// Since specifies the time from which to start streaming logs.
	Since() string
}

type loggingService struct {
	clusterService DBClusterService
	clusterManager clustermanager.ClusterManager
	logger         logger.Logger
}

type LoggingServiceSpec struct {
	ClusterService DBClusterService
	ClusterManager clustermanager.ClusterManager
	Logger         logger.Logger
}

type LogStreamer struct {
	clientSet      *kubernetes.Clientset
	resourcePodMap map[string]*corev1.Pod
	logOptions     *corev1.PodLogOptions
}

type StreamedLog struct {
	data string
	err  error
}

func (s *StreamedLog) Data() string {
	return s.data
}

func (s *StreamedLog) Error() error {
	return s.err
}

func NewLoggingService(spec LoggingServiceSpec) ClusterLoggingService {
	return &loggingService{
		clusterService: spec.ClusterService,
		clusterManager: spec.ClusterManager,
		logger:         spec.Logger,
	}
}

func (s *loggingService) GetLogsForResources(ctx context.Context, orgID string, resources []*models.StackResource, options LoggingParams) (interfaces.ServerSideStreamable, error) {
	cluster, err := s.clusterService.GetClusterForOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}

	restConfig, cerr := s.clusterManager.GetRestConfig(cluster.ID)
	if cerr != nil {
		return nil, cerr
	}

	clientset, kerr := kubernetes.NewForConfig(restConfig)
	if kerr != nil {
		return nil, kerr
	}

	client, cerr := s.clusterManager.GetClient(cluster.ID)
	if cerr != nil {
		return nil, cerr
	}

	targetResourcePodMap := make(map[string]*corev1.Pod)
	unReadyResources := []string{}
	for _, resource := range resources {
		if resource.Status.State == models.StackResourcePhaseReady && resource.Status.InternalServiceName != nil {
			pod, err := s.attachablePodFromService(ctx, clientset, client, *resource.Status.InternalServiceName, resource.Namespace)
			if err != nil {
				return nil, fmt.Errorf("failed to get attachable pod for resource %s: %v", resource.Name, err)
			}
			if pod != nil {
				targetResourcePodMap[resource.Name] = pod
			} else {
				unReadyResources = append(unReadyResources, resource.Name)
			}
		} else {
			unReadyResources = append(unReadyResources, resource.Name)
		}
	}
	if len(targetResourcePodMap) == 0 {
		return nil, fmt.Errorf("resources are not ready: %v", unReadyResources)
	}

	podLogOptions := &corev1.PodLogOptions{}
	if options.Follow() {
		podLogOptions.Follow = true
	}
	if options.TailLines() != 0 {
		podLogOptions.TailLines = ptr.To(int64(options.TailLines()))
	}
	if options.Since() != "" {
		since, err := time.ParseDuration(options.Since())
		if err == nil {
			podLogOptions.SinceSeconds = ptr.To(int64(since.Seconds()))
		} else {
			s.logger.Warnf("failed to parse 'since' duration: %v. Ignoring passed since options", err)
		}
	}

	return &LogStreamer{
		resourcePodMap: targetResourcePodMap,
		logOptions:     podLogOptions,
		clientSet:      clientset,
	}, nil
}

func (s *LogStreamer) Stream(ctx context.Context) (<-chan interfaces.StreamObject, error) {
	streamerChan := make(chan interfaces.StreamObject)
	var wg sync.WaitGroup
	for resourceName, pod := range s.resourcePodMap {
		req := s.clientSet.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, s.logOptions)
		wg.Add(1)
		go func(resourceName string, req *rest.Request) {
			defer wg.Done()
			podLog, err := req.Stream(ctx)
			if err != nil {
				if err == context.Canceled {
					obj := &StreamedLog{
						err: context.Canceled,
					}
					safeWriteToChannel(ctx, streamerChan, obj)
				}
				return
			}
			reader := bufio.NewReader(podLog)
			defer podLog.Close()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				line, err := reader.ReadString('\n')
				if err == io.EOF {
					return
				}
				if err == context.Canceled {
					return
				}

				if err != nil {
					obj := &StreamedLog{
						err: fmt.Errorf("error reading log line: %v", err),
					}
					safeWriteToChannel(ctx, streamerChan, obj)
					return
				}
				var obj *StreamedLog
				var jsonLine string
				if err := json.Unmarshal([]byte(line), &jsonLine); err != nil {
					obj = &StreamedLog{
						data: fmt.Sprintf("[%s] %s", resourceName, line),
						err:  nil,
					}
				} else {
					obj = &StreamedLog{
						data: fmt.Sprintf("[%s] %s", resourceName, jsonLine),
						err:  nil,
					}
				}
				safeWriteToChannel(ctx, streamerChan, obj)
			}
		}(resourceName, req)
	}

	go func() {
		wg.Wait()
		close(streamerChan)
	}()
	return streamerChan, nil
}

func safeWriteToChannel(ctx context.Context, ch chan<- interfaces.StreamObject, obj interfaces.StreamObject) {
	select {
	case <-ctx.Done():
		return
	case ch <- obj:
	}
}

func (k *loggingService) attachablePodFromService(ctx context.Context, clientset *kubernetes.Clientset, ctrlrmtimeClient client.Client, serviceName string, serviceNamespace string) (*corev1.Pod, error) {
	svc := &corev1.Service{}
	if err := ctrlrmtimeClient.Get(
		ctx, types.NamespacedName{
			Name:      serviceName,
			Namespace: serviceNamespace,
		},
		svc,
	); err != nil {
		return nil, err
	}
	attachablePod, err := attachablePodForObject(clientset, svc, time.Second*10)
	if err != nil {
		return nil, err
	}
	return attachablePod, nil

}

func attachablePodForObject(client *kubernetes.Clientset, object runtime.Object, timeout time.Duration) (*corev1.Pod, error) {
	namespace, selector, err := polymorphichelpers.SelectorsForObject(object)
	if err != nil {
		return nil, fmt.Errorf("cannot attach to %T: %v", object, err)
	}
	sortBy := func(pods []*corev1.Pod) sort.Interface { return sort.Reverse(podutils.ActivePods(pods)) }
	pod, _, err := polymorphichelpers.GetFirstPod(client.CoreV1(), namespace, selector.String(), timeout, sortBy)
	return pod, err
}
