package clusterresource

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/clients"
	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/interfaces"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/samber/lo"
	"golang.org/x/time/rate"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	DefaultLogStreamBufferSize     = 1000
	DefaultStreamTimeoutDuration   = 20 * time.Minute
	DefaultLogStreamRateLimit      = 50 * time.Millisecond
	DefaultLogStreamRateLimitBurst = 200
)

type ResourceError struct {
	ResourceName string
	Error        error
}

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

	client, cerr := s.clusterManager.GetClient(cluster.ID)
	if cerr != nil {
		return nil, cerr
	}

	k8sclient, cerr := clients.NewKubernetesClient(clients.KubernetesClientSpec{
		RestConfig:              restConfig,
		ControllerRuntimeClient: client,
		Logger:                  s.logger,
	})
	if cerr != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", cerr)
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
		s.logger.Warnf("resource pod resolution failures: %v", errors)
	}

	if len(resourcePodMap) == 0 {
		return nil, fmt.Errorf("no attachable pods found for ready resources")
	}

	return &LogStreamer{
		resourcePodMap: resourcePodMap,
		k8sclient:      k8sclient,
		logOptions:     options,
		Logger:         s.logger,
		streamConfig: LogStreamConfig{
			LogStreamBufferSize:   DefaultLogStreamBufferSize,
			StreamTimeoutDuration: DefaultStreamTimeoutDuration,
			rateLimiter:           rate.NewLimiter(rate.Every(DefaultLogStreamRateLimit), DefaultLogStreamRateLimitBurst),
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
			s.logger.Errorf("failed to get pod for resource %s: %v", resource.Name, err)
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
	logOptions     LoggingParams
	streamConfig   LogStreamConfig
	Logger         logger.Logger
}

type LogStreamConfig struct {
	LogStreamBufferSize   int
	StreamTimeoutDuration time.Duration
	rateLimiter           *rate.Limiter
}

func (s *LogStreamer) Stream(ctx context.Context) (<-chan interfaces.StreamObject, error) {
	streamerChan := make(chan interfaces.StreamObject, s.streamConfig.LogStreamBufferSize)
	ctxWithTimeout, cancel := context.WithTimeout(ctx, s.streamConfig.StreamTimeoutDuration)

	podResourceStreamMap := make(map[string]<-chan *clients.LogLine)
	for resourceName, pod := range s.resourcePodMap {
		podLogChan, err := s.k8sclient.StreamPodLogs(ctxWithTimeout, pod, s.logOptions)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to stream logs for resource '%s': %v", resourceName, err)
		}
		podResourceStreamMap[resourceName] = podLogChan
	}

	var wg sync.WaitGroup
	for resourceName, podLogChan := range podResourceStreamMap {
		wg.Add(1)
		go func(resourceName string, podLogChan <-chan *clients.LogLine) {
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
		s.Logger.Infof("log stream channel is full, applying rate limiting..")
	}

	if err := s.streamConfig.rateLimiter.Wait(ctx); err != nil {
		s.Logger.Errorf("rate limiter wait failed: %v", err)
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
		s.Logger.Warnf("log stream channel is full, dropping message")
	}
}
