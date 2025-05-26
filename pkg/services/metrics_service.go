package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/interfaces"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type MetricsService interface {
	GetMetricsForStackResource(ctx context.Context, orgID string, stackID string, stackResourceName string) (*models.ResourceMetrics, error)
	GetMetricsForStack(ctx context.Context, orgID string, stackID string) (*models.ResourceMetrics, error)
	StreamMetricsForStackResource(ctx context.Context, orgID string, stackID string, stackResourceName string) (interfaces.ServerSideStreamable, error)
	StreamMetricsForStack(ctx context.Context, orgID string, stackID string) (interfaces.ServerSideStreamable, error)
	ClusterResourceServiceInjectable
}

type metricsService struct {
	clusterService       ClusterService
	stackResourceService StackResourceService
	stackService         StackService
	logger               logger.Logger
	ClusterResourceServiceDeps
}

type MetricsServiceSpec struct {
	ClusterService       ClusterService
	StackResourceService StackResourceService
	StackService         StackService
	Logger               logger.Logger
}

func NewMetricsService(spec MetricsServiceSpec) MetricsService {
	return &metricsService{
		stackService:         spec.StackService,
		clusterService:       spec.ClusterService,
		stackResourceService: spec.StackResourceService,
		logger:               spec.Logger,
	}
}

func (s *metricsService) GetMetricsForStackResource(ctx context.Context, orgID string, stackID string, stackResourceName string) (*models.ResourceMetrics, error) {
	resource, err := s.stackResourceService.GetByStackIDAndResourceName(ctx, stackID, stackResourceName)
	if err != nil {
		s.logger.Errorf("failed to get stack resource: %v", err)
		return nil, err
	}
	if resource.Status != nil && resource.Status.State != models.StackResourcePhaseReady {
		return nil, errors.UnprocessableEntity("resource %s is not ready for metrics", resource.Name)
	}

	res, cerr := s.ClusterResourceServiceDeps.ClusterMetricsService.GetMetricsForResource(ctx, orgID, resource)
	if cerr != nil {
		return nil, cerr
	}
	return res, nil
}

func (s *metricsService) GetMetricsForStack(ctx context.Context, orgID string, stackID string) (*models.ResourceMetrics, error) {
	stack, err := s.stackService.GetStack(ctx, stackID)
	if err != nil {
		s.logger.Errorf("failed to get stack: %v", err)
		return nil, err
	}

	if stack.Status != nil && stack.Status.State != models.StackReady {
		return nil, errors.UnprocessableEntity("stack %s is not ready for metrics", stackID)
	}

	res, cerr := s.ClusterResourceServiceDeps.ClusterMetricsService.GetMetricsForStack(ctx, orgID, stack)
	if cerr != nil {
		s.logger.Errorf("failed to get metrics for stack: %v", cerr)
		return nil, cerr
	}
	return res, nil
}

func (s *metricsService) StreamMetricsForStackResource(ctx context.Context, orgID string, stackID string, stackResourceName string) (interfaces.ServerSideStreamable, error) {
	resource, err := s.stackResourceService.GetByStackIDAndResourceName(ctx, stackID, stackResourceName)
	if err != nil {
		s.logger.Errorf("failed to get stack resource: %v", err)
		return nil, err
	}
	if resource.Status != nil && resource.Status.State != models.StackResourcePhaseReady {
		return nil, errors.UnprocessableEntity("resource %s is not ready for metrics", resource.Name)
	}
	streamer, cerr := s.ClusterResourceServiceDeps.ClusterMetricsService.StreamMetricsForResource(ctx, orgID, resource)
	if cerr != nil {
		s.logger.Errorf("failed to stream metrics for resource: %v", cerr)
		return nil, cerr
	}
	return streamer, nil
}
func (s *metricsService) StreamMetricsForStack(ctx context.Context, orgID string, stackID string) (interfaces.ServerSideStreamable, error) {
	stack, err := s.stackService.GetStack(ctx, stackID)
	if err != nil {
		s.logger.Errorf("failed to get stack: %v", err)
		return nil, err
	}

	if stack.Status != nil && stack.Status.State != models.StackReady {
		return nil, errors.UnprocessableEntity("stack %s is not ready for metrics", stackID)
	}
	streamer, cerr := s.ClusterResourceServiceDeps.ClusterMetricsService.StreamMetricsForStack(ctx, orgID, stack)
	if cerr != nil {
		s.logger.Errorf("failed to stream metrics for stack: %v", cerr)
		return nil, cerr
	}
	return streamer, nil
}
