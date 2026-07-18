package clusterresource

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Stackdome/stackdome/pkg/clients"
	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/interfaces"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/samber/lo"
	"golang.org/x/time/rate"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ResourceError struct {
	ResourceName string
	Error        error
}

type ClusterLoggingService interface {
	// GetLogsForResources retrieves logs for the specified resources in the cluster.
	GetLogsForResources(ctx context.Context, orgID string, resources []*models.StackResource, options LoggingParams) (interfaces.ServerSideStreamable, error)
	// GetLogsForBuildPod streams logs from the pod of the given build Job.
	GetLogsForBuildPod(ctx context.Context, orgID string, namespace string, jobName string, options LoggingParams) (interfaces.ServerSideStreamable, error)
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

	ctrlRuntimeclient, cerr := s.clusterManager.GetClient(cluster.ID)
	if cerr != nil {
		return nil, cerr
	}

	k8sclient, cerr := clients.NewKubernetesClient(clients.KubernetesClientSpec{
		RestConfig:              restConfig,
		ControllerRuntimeClient: ctrlRuntimeclient,
		Logger:                  logger.NewLoggerWithPrefix(ctx, "kubernetes-client"),
	})
	if cerr != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", cerr)
	}

	readyResources := lo.Filter(resources, func(resource *models.StackResource, _ int) bool {
		return resource.Status.State == models.StackResourcePhaseReady && resource.Status.InternalServiceName != nil
	})

	unreadyResources, _ := lo.Difference(resources, readyResources)
	if len(unreadyResources) == len(resources) {
		unreadyResourceNames := lo.Map(unreadyResources, func(r *models.StackResource, _ int) string { return r.Name })
		return nil, fmt.Errorf("no resources are ready for logging: %v", unreadyResourceNames)
	}

	resourcePodMap, errors := s.resolveResourcePods(ctx, k8sclient, readyResources)
	if len(errors) > 0 {
		s.logger.Warn(ctx, "resource pod resolution failures: %v", errors)
	}

	if len(resourcePodMap) == 0 {
		return nil, fmt.Errorf("no attachable pods found for ready resources")
	}

	return &LogStreamer{
		resourcePodMap: resourcePodMap,
		k8sclient:      k8sclient,
		Logger:         s.logger,
		streamConfig: LogStreamConfig{
			logStreamBufferSize:   DefaultLogStreamBufferSize,
			streamTimeoutDuration: DefaultStreamTimeoutDuration,
			rateLimiter:           rate.NewLimiter(rate.Every(DefaultLogStreamRateLimit), DefaultLogStreamRateLimitBurst),
			logOptions:            options,
			maxLogStreamErrors:    MaxLogStreamErrors,
		},
	}, nil
}

func (s *loggingService) GetLogsForBuildPod(ctx context.Context, orgID string, namespace string, jobName string, options LoggingParams) (interfaces.ServerSideStreamable, error) {
	cluster, err := s.clusterService.GetClusterForOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}

	restConfig, cerr := s.clusterManager.GetRestConfig(cluster.ID)
	if cerr != nil {
		return nil, cerr
	}

	ctrlRuntimeclient, cerr := s.clusterManager.GetClient(cluster.ID)
	if cerr != nil {
		return nil, cerr
	}

	k8sclient, cerr := clients.NewKubernetesClient(clients.KubernetesClientSpec{
		RestConfig:              restConfig,
		ControllerRuntimeClient: ctrlRuntimeclient,
		Logger:                  logger.NewLoggerWithPrefix(ctx, "kubernetes-client"),
	})
	if cerr != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", cerr)
	}

	pod, perr := k8sclient.BuildPodForJob(ctx, namespace, jobName)
	if perr != nil {
		return nil, fmt.Errorf("failed to find build pod for job %s: %w", jobName, perr)
	}
	if pod == nil {
		return nil, fmt.Errorf("no build pod found for job %s: the build has not started yet or its logs have been pruned", jobName)
	}

	return &LogStreamer{
		resourcePodMap: map[string]*corev1.Pod{jobName: pod},
		k8sclient:      k8sclient,
		Logger:         s.logger,
		streamConfig: LogStreamConfig{
			logStreamBufferSize:   DefaultLogStreamBufferSize,
			streamTimeoutDuration: DefaultStreamTimeoutDuration,
			rateLimiter:           rate.NewLimiter(rate.Every(DefaultLogStreamRateLimit), DefaultLogStreamRateLimitBurst),
			logOptions:            options,
			maxLogStreamErrors:    MaxLogStreamErrors,
		},
	}, nil
}

func (s *loggingService) resolveResourcePods(ctx context.Context, k8sclient clients.KubernetesClient, resources []*models.StackResource) (map[string]*corev1.Pod, []ResourceError) {
	resourcePodMap := make(map[string]*corev1.Pod)
	var errors []ResourceError

	for _, resource := range resources {
		pod, err := k8sclient.AttachablePodFromService(ctx, types.NamespacedName{
			Name:      *resource.Status.InternalServiceName,
			Namespace: resource.Namespace,
		})
		if err != nil {
			errors = append(errors, ResourceError{
				ResourceName: resource.Name,
				Error:        err,
			})
			s.logger.Error(ctx, "failed to get pod for resource %s: %v", resource.Name, err)
			continue
		}
		if pod != nil {
			resourcePodMap[resource.Name] = pod
		} else {
			errors = append(errors, ResourceError{
				ResourceName: resource.Name,
				Error:        fmt.Errorf("no attachable pod found"),
			})
		}
	}
	return resourcePodMap, errors
}

type LogStreamer struct {
	k8sclient      clients.KubernetesClient
	resourcePodMap map[string]*corev1.Pod
	streamConfig   LogStreamConfig
	Logger         logger.Logger
}

type LogStreamConfig struct {
	logStreamBufferSize   int
	streamTimeoutDuration time.Duration
	rateLimiter           *rate.Limiter
	logOptions            LoggingParams
	maxLogStreamErrors    int
}

var _ clients.PodLogStreamOptions = &LogStreamConfig{}

func (s *LogStreamConfig) Follow() bool {
	return s.logOptions.Follow()
}
func (s *LogStreamConfig) TailLines() int {
	return s.logOptions.TailLines()
}
func (s *LogStreamConfig) Since() string {
	return s.logOptions.Since()
}
func (s *LogStreamConfig) MaxErrors() int {
	return s.maxLogStreamErrors
}

func (s *LogStreamer) Stream(ctx context.Context) (<-chan interfaces.StreamObject, error) {
	streamerChan := make(chan interfaces.StreamObject, s.streamConfig.logStreamBufferSize)
	// Clients will have to reconnect after this duration.
	ctxWithTimeout, cancel := context.WithTimeout(ctx, s.streamConfig.streamTimeoutDuration)

	podResourceStreamMap := make(map[string]<-chan *clients.LogStreamObject)
	for resourceName, pod := range s.resourcePodMap {
		podLogChan, err := s.k8sclient.StreamPodLogs(ctxWithTimeout, pod, &s.streamConfig)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to stream logs for resource '%s': %w", resourceName, err)
		}
		podResourceStreamMap[resourceName] = podLogChan
	}

	var wg sync.WaitGroup
	for resourceName, podLogChan := range podResourceStreamMap {
		wg.Add(1)
		go func(resourceName string, podLogChan <-chan *clients.LogStreamObject) {
			defer wg.Done()
			for {
				select {
				case <-ctxWithTimeout.Done():
					return
				case logLine, ok := <-podLogChan:
					if !ok {
						return
					}
					if logLine.Error != nil {
						s.safeWriteToChannel(ctxWithTimeout, streamerChan, &StreamedLog{err: resourceNamePrefixedError(resourceName, logLine.Error)})
						continue
					}
					s.safeWriteToChannel(ctxWithTimeout, streamerChan, &StreamedLog{data: resourceNamePrefixedLogLine(resourceName, logLine.Data)})
				}
			}
		}(resourceName, podLogChan)
	}
	go func() {
		wg.Wait()
		cancel()
		close(streamerChan)
	}()
	return streamerChan, nil
}

func resourceNamePrefixedLogLine(prefix string, line string) string {
	return fmt.Sprintf("[%s]: %s", prefix, line)
}

func resourceNamePrefixedError(prefix string, err error) error {
	return fmt.Errorf("[%s]: %w", prefix, err)
}

func (s *LogStreamer) safeWriteToChannel(ctx context.Context, ch chan<- interfaces.StreamObject, obj interfaces.StreamObject) {
	select {
	case <-ctx.Done():
		return
	case ch <- obj:
		return
	default:
		// Channel full
		s.Logger.Info(ctx, "log stream channel is full, applying rate limiting..")
	}

	if err := s.streamConfig.rateLimiter.Wait(ctx); err != nil {
		s.Logger.Error(ctx, "rate limiter wait failed: %v", err)
		return
	}
	// Try again after waiting for the rate limiter.
	select {
	case <-ctx.Done():
		return
	case ch <- obj:
		return
	default:
		// Channel is stil full, drop the message
		s.Logger.Warn(ctx, "log stream channel is full, dropping message")
	}
}
