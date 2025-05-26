package clusterresource

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/clients"
	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/interfaces"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/samber/lo"
	"golang.org/x/time/rate"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const (
	MetricsServerPollingInterval = 15 * time.Second

	StackResourceTargetType = "stack_resource"
	StackTargetType         = "stack"
)

type ClusterMetricsService interface {
	GetMetricsForResource(ctx context.Context, orgID string, stackResource *models.StackResource) (*models.ResourceMetrics, error)
	GetMetricsForStack(ctx context.Context, orgID string, stack *models.Stack) (*models.ResourceMetrics, error)
	StreamMetricsForResource(ctx context.Context, orgID string, stackResource *models.StackResource) (interfaces.ServerSideStreamable, error)
	StreamMetricsForStack(ctx context.Context, orgID string, stack *models.Stack) (interfaces.ServerSideStreamable, error)
}

type clusterMetricsService struct {
	clusterService DBClusterService
	clusterManager clustermanager.ClusterManager
	logger         logger.Logger
}

type ClusterMetricsServiceSpec struct {
	ClusterService DBClusterService
	ClusterManager clustermanager.ClusterManager
	Logger         logger.Logger
}

type clusterMetricsStreamer struct {
	target    metricsStreamTarget
	k8sclient clients.KubernetesClient
	logger    logger.Logger
	config    metricsStreamConfig
}

type metricsStreamTarget struct {
	// Either a stack resource or a stack
	stackResource *models.StackResource
	stack         *models.Stack
}

func (t *metricsStreamTarget) Type() string {
	if t.stackResource != nil {
		return StackResourceTargetType
	} else if t.stack != nil {
		return StackTargetType
	}
	return ""
}

type metricsStreamConfig struct {
	streamBufferSize      int
	streamTimeoutDuration time.Duration
	rateLimiter           *rate.Limiter
	maxLogStreamErrors    int
	pollInterval          time.Duration
}

func (m *metricsStreamConfig) PollInterval() time.Duration {
	return m.pollInterval
}
func (m *metricsStreamConfig) MaxErrors() int {
	return m.maxLogStreamErrors
}

type StreamedMetrics struct {
	data []byte
	err  error
}

func (s *StreamedMetrics) Data() string {
	return string(s.data)
}

func (s *StreamedMetrics) Error() error {
	return s.err
}

func NewClusterMetricsService(spec ClusterMetricsServiceSpec) ClusterMetricsService {
	return &clusterMetricsService{
		clusterService: spec.ClusterService,
		clusterManager: spec.ClusterManager,
		logger:         spec.Logger,
	}
}

func (s *clusterMetricsService) GetMetricsForResource(ctx context.Context, orgID string, stackResource *models.StackResource) (*models.ResourceMetrics, error) {
	cluster, err := s.clusterService.GetClusterForOrg(ctx, orgID)
	if err != nil {
		return nil, newError("failed to get cluster for organisation", err)
	}

	clusterClient, cErr := s.clusterManager.GetClient(cluster.ID)
	if cErr != nil {
		s.logger.Errorf("failed to get cluster client: %v", cErr)
		return nil, newError("failed to get cluster client", cErr)
	}

	restConfig, cErr := s.clusterManager.GetRestConfig(cluster.ID)
	if cErr != nil {
		s.logger.Errorf("failed to get cluster rest config: %v", cErr)
		return nil, newError("failed to get cluster rest config", cErr)
	}

	k8sClient, cerr := clients.NewKubernetesClient(clients.KubernetesClientSpec{
		RestConfig:              restConfig,
		ControllerRuntimeClient: clusterClient,
		Logger:                  s.logger,
	})
	if cerr != nil {
		return nil, newError("failed to create Kubernetes client", cerr)
	}

	metricsObj, cerr := k8sClient.GetNamespaceMetrics(ctx, stackResource.Namespace)
	if cerr != nil {
		return nil, cerr
	}

	podMetricsForResource := lo.Filter(metricsObj.NamespaceMetrics, func(podMetrics v1beta1.PodMetrics, _ int) bool {
		resourceNameInPod, ok := podMetrics.ObjectMeta.Labels["resource"]
		return ok && resourceNameInPod == stackResource.Name
	})

	// TODO: Handle replicas? What if a stack resource has multiple replica pods?
	metrics := accumulatePodMetrics(podMetricsForResource)

	for nodeName, capacity := range metricsObj.NodeCapacityMap {
		metrics.NodeCapacities = append(metrics.NodeCapacities, &models.NodeCapacity{
			NodeName: nodeName,
			CPU:      capacity.Cpu(),
			Memory:   capacity.Memory(),
			Storage:  capacity.Storage(),
		})
	}

	return metrics, nil
}

func (s *clusterMetricsService) GetMetricsForStack(ctx context.Context, orgID string, stack *models.Stack) (*models.ResourceMetrics, error) {
	cluster, err := s.clusterService.GetClusterForOrg(ctx, orgID)
	if err != nil {
		return nil, newError("failed to get cluster for organisation", err)
	}

	clusterClient, cErr := s.clusterManager.GetClient(cluster.ID)
	if cErr != nil {
		s.logger.Errorf("failed to get cluster client: %v", cErr)
		return nil, newError("failed to get cluster client", cErr)
	}

	restConfig, cErr := s.clusterManager.GetRestConfig(cluster.ID)
	if cErr != nil {
		s.logger.Errorf("failed to get cluster rest config: %v", cErr)
		return nil, newError("failed to get cluster rest config", cErr)
	}

	k8sClient, cerr := clients.NewKubernetesClient(clients.KubernetesClientSpec{
		RestConfig:              restConfig,
		ControllerRuntimeClient: clusterClient,
		Logger:                  s.logger,
	})
	if cerr != nil {
		return nil, newError("failed to create Kubernetes client", cerr)
	}

	metricsObj, cerr := k8sClient.GetNamespaceMetrics(ctx, stack.Namespace)
	if cerr != nil {
		return nil, cerr
	}
	// TODO: Handle replicas? What if a stack resource has multiple replica pods?
	metrics := accumulatePodMetrics(metricsObj.NamespaceMetrics)
	for nodeName, capacity := range metricsObj.NodeCapacityMap {
		metrics.NodeCapacities = append(metrics.NodeCapacities, &models.NodeCapacity{
			NodeName: nodeName,
			CPU:      capacity.Cpu(),
			Memory:   capacity.Memory(),
			Storage:  capacity.Storage(),
		})
	}
	return metrics, nil
}

func (s *clusterMetricsService) StreamMetricsForResource(ctx context.Context, orgID string, stackResource *models.StackResource) (interfaces.ServerSideStreamable, error) {
	cluster, err := s.clusterService.GetClusterForOrg(ctx, orgID)
	if err != nil {
		return nil, newError("failed to get cluster for organisation", err)
	}

	clusterClient, cErr := s.clusterManager.GetClient(cluster.ID)
	if cErr != nil {
		s.logger.Errorf("failed to get cluster client: %v", cErr)
		return nil, newError("failed to get cluster client", cErr)
	}

	restConfig, cErr := s.clusterManager.GetRestConfig(cluster.ID)
	if cErr != nil {
		s.logger.Errorf("failed to get cluster rest config: %v", cErr)
		return nil, newError("failed to get cluster rest config", cErr)
	}

	k8sClient, cerr := clients.NewKubernetesClient(clients.KubernetesClientSpec{
		RestConfig:              restConfig,
		ControllerRuntimeClient: clusterClient,
		Logger:                  s.logger,
	})
	if cerr != nil {
		return nil, newError("failed to create Kubernetes client", cerr)
	}
	return &clusterMetricsStreamer{
		target: metricsStreamTarget{
			stackResource: stackResource,
		},
		k8sclient: k8sClient,
		logger:    s.logger,
		config: metricsStreamConfig{
			streamBufferSize:      DefaultLogStreamBufferSize,
			streamTimeoutDuration: DefaultStreamTimeoutDuration,
			rateLimiter:           rate.NewLimiter(rate.Every(DefaultLogStreamRateLimit), DefaultLogStreamRateLimitBurst),
			maxLogStreamErrors:    MaxLogStreamErrors,
			pollInterval:          MetricsServerPollingInterval,
		},
	}, nil
}

func (s *clusterMetricsService) StreamMetricsForStack(ctx context.Context, orgID string, stack *models.Stack) (interfaces.ServerSideStreamable, error) {
	cluster, err := s.clusterService.GetClusterForOrg(ctx, orgID)
	if err != nil {
		return nil, newError("failed to get cluster for organisation", err)
	}

	clusterClient, cErr := s.clusterManager.GetClient(cluster.ID)
	if cErr != nil {
		s.logger.Errorf("failed to get cluster client: %v", cErr)
		return nil, newError("failed to get cluster client", cErr)
	}

	restConfig, cErr := s.clusterManager.GetRestConfig(cluster.ID)
	if cErr != nil {
		s.logger.Errorf("failed to get cluster rest config: %v", cErr)
		return nil, newError("failed to get cluster rest config", cErr)
	}

	k8sClient, cerr := clients.NewKubernetesClient(clients.KubernetesClientSpec{
		RestConfig:              restConfig,
		ControllerRuntimeClient: clusterClient,
		Logger:                  s.logger,
	})
	if cerr != nil {
		return nil, newError("failed to create Kubernetes client", cerr)
	}
	return &clusterMetricsStreamer{
		target: metricsStreamTarget{
			stack: stack,
		},
		k8sclient: k8sClient,
		logger:    s.logger,
		config: metricsStreamConfig{
			streamBufferSize:      DefaultLogStreamBufferSize,
			streamTimeoutDuration: DefaultStreamTimeoutDuration,
			rateLimiter:           rate.NewLimiter(rate.Every(DefaultLogStreamRateLimit), DefaultLogStreamRateLimitBurst),
			maxLogStreamErrors:    MaxLogStreamErrors,
			pollInterval:          MetricsServerPollingInterval,
		},
	}, nil
}

func (s *clusterMetricsStreamer) Stream(ctx context.Context) (<-chan interfaces.StreamObject, error) {
	switch s.target.Type() {
	case StackResourceTargetType:
		return s.streamStackResourceMetrics(ctx)
	case StackTargetType:
		return s.streamStackMetrics(ctx)
	default:
		return nil, fmt.Errorf("unknown metrics stream target type: %s", s.target.Type())
	}
}

func (s *clusterMetricsStreamer) streamStackResourceMetrics(ctx context.Context) (<-chan interfaces.StreamObject, error) {
	if s.target.stackResource == nil {
		return nil, fmt.Errorf("metrics stream target is not a stack resource")
	}
	namespaceMetricsChan, err := s.k8sclient.StreamNamespaceMetrics(ctx, s.target.stackResource.Namespace, &s.config)
	if err != nil {
		return nil, newError("failed to stream metrics for resource", err)
	}
	streamerChan := make(chan interfaces.StreamObject, s.config.streamBufferSize)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case metricsObj, ok := <-namespaceMetricsChan:
				if !ok {
					return
				}
				if metricsObj.Error != nil {
					s.safeWriteToChannel(ctx, streamerChan, &StreamedMetrics{err: metricsObj.Error})
					continue
				}
				filteredMetrics := lo.Filter(metricsObj.NamespaceMetrics, func(podMetrics v1beta1.PodMetrics, _ int) bool {
					resourceNameInPod, ok := podMetrics.ObjectMeta.Labels["resource"]
					return ok && resourceNameInPod == s.target.stackResource.Name
				})
				metrics := accumulatePodMetrics(filteredMetrics)
				// Convert to api schema.
				apiObj := presenters.PresentResourceMetrics(metrics)
				data, err := json.Marshal(apiObj)
				if err != nil {
					s.logger.Errorf("failed to marshal metrics object: %v", err)
					continue
				}
				s.safeWriteToChannel(ctx, streamerChan, &StreamedMetrics{data: data})
			}
		}
	}()

	go func() {
		wg.Wait()
		close(streamerChan)
	}()
	return streamerChan, nil
}

func (s *clusterMetricsStreamer) streamStackMetrics(ctx context.Context) (<-chan interfaces.StreamObject, error) {
	if s.target.stack == nil {
		return nil, fmt.Errorf("metrics stream target is not a stack")
	}
	namespaceMetricsChan, err := s.k8sclient.StreamNamespaceMetrics(ctx, s.target.stack.Namespace, &s.config)
	if err != nil {
		return nil, newError("failed to stream metrics for stack", err)
	}
	streamerChan := make(chan interfaces.StreamObject, s.config.streamBufferSize)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case metricsObj, ok := <-namespaceMetricsChan:
				if !ok {
					return
				}
				if metricsObj.Error != nil {
					s.safeWriteToChannel(ctx, streamerChan, &StreamedMetrics{err: metricsObj.Error})
					continue
				}
				metrics := accumulatePodMetrics(metricsObj.NamespaceMetrics)
				// Convert to api schema.
				apiObj := presenters.PresentResourceMetrics(metrics)
				data, err := json.Marshal(apiObj)
				if err != nil {
					s.logger.Errorf("failed to marshal metrics object: %v", err)
					continue
				}
				s.safeWriteToChannel(ctx, streamerChan, &StreamedMetrics{data: data})
			}
		}
	}()

	go func() {
		wg.Wait()
		close(streamerChan)
	}()
	return streamerChan, nil
}

func (s *clusterMetricsStreamer) safeWriteToChannel(ctx context.Context, ch chan<- interfaces.StreamObject, obj interfaces.StreamObject) {

	select {
	case <-ctx.Done():
		return
	case ch <- obj:
		return
	default:
		// Channel full
		s.logger.Infof("log stream channel is full, applying rate limiting..")
	}
	if err := s.config.rateLimiter.Wait(ctx); err != nil {
		s.logger.Errorf("rate limiter wait failed: %v", err)
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
		s.logger.Warnf("log stream channel is full, dropping message")
	}
}

func accumulatePodMetrics(podMetricsList []v1beta1.PodMetrics) *models.ResourceMetrics {
	metrics := &models.ResourceMetrics{
		AssignedNodes:  make([]string, 0),
		NodeCapacities: make([]*models.NodeCapacity, 0),
	}
	for _, podMetrics := range podMetricsList {
		for _, container := range podMetrics.Containers {
			metrics.CPUUsage.Add(container.Usage[corev1.ResourceCPU])
			metrics.MemoryUsage.Add(container.Usage[corev1.ResourceMemory])
			metrics.TimeStamp = podMetrics.Timestamp.Time.UTC()
		}
		assignedNode, ok := podMetrics.Annotations[models.AssignedNodeAnnotation]
		if ok && assignedNode != "" {
			if !lo.Contains(metrics.AssignedNodes, assignedNode) {
				metrics.AssignedNodes = append(metrics.AssignedNodes, assignedNode)
			}
		}
	}
	return metrics
}
