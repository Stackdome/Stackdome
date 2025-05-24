package services

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/ashishmax31/stackdome-api-server/pkg/interfaces"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type LoggingParams struct {
	follow    bool
	tailLines int
	since     string
}

func (p *LoggingParams) Follow() bool {
	return p.follow
}
func (p *LoggingParams) TailLines() int {
	return p.tailLines
}
func (p *LoggingParams) Since() string {
	return p.since
}

func NewLoggingParams(queryValues url.Values) (*LoggingParams, error) {
	l := &LoggingParams{}
	tailLines := queryValues.Get("tail")
	if tailLines != "" {
		tailNum, err := strconv.Atoi(tailLines)
		if err != nil {
			return nil, fmt.Errorf("invalid tail value: %w", err)
		}
		if tailNum < 0 {
			tailNum = 1
		}
		if tailNum > 500 {
			tailNum = 500
		}
		l.tailLines = tailNum
	} else {
		l.tailLines = 200 // Default tail lines if not specified
	}

	l.follow = queryValues.Get("follow") == "true"

	if queryValues.Get("since") != "" {
		l.since = queryValues.Get("since")
	}
	return l, nil
}

type LoggingService interface {
	StreamLogsForStackResource(ctx context.Context, orgID string, stackID string, stackResourceName string, options *LoggingParams) (interfaces.ServerSideStreamable, error)
	StreamLogsForStack(ctx context.Context, orgID string, stackID string, options *LoggingParams) (interfaces.ServerSideStreamable, error)
	ClusterResourceServiceInjectable
}

type loggingService struct {
	clusterService       ClusterService
	stackResourceService StackResourceService
	logger               logger.Logger
	ClusterResourceServiceDeps
}

type LoggingServiceSpec struct {
	ClusterService       ClusterService
	StackResourceService StackResourceService
	Logger               logger.Logger
}

func NewLoggingService(spec LoggingServiceSpec) LoggingService {
	return &loggingService{
		clusterService:       spec.ClusterService,
		stackResourceService: spec.StackResourceService,
		logger:               spec.Logger,
	}
}

func (s *loggingService) StreamLogsForStackResource(ctx context.Context, orgID string, stackID string, stackResourceID string, options *LoggingParams) (interfaces.ServerSideStreamable, error) {
	stackResource, err := s.stackResourceService.GetByStackIDAndResourceName(ctx, stackID, stackResourceID)
	if err != nil {
		return nil, err
	}

	if stackResource.Status.State != models.StackResourcePhaseReady || stackResource.Status.InternalServiceName == nil {
		return nil, fmt.Errorf("resource %s is not ready for logging", stackResource.Name)
	}

	logStreamer, serr := s.ClusterLoggingService.GetLogsForResources(ctx, orgID, []*models.StackResource{stackResource}, options)
	if serr != nil {
		return nil, serr
	}

	return logStreamer, nil
}

func (s *loggingService) StreamLogsForStack(ctx context.Context, orgID string, stackID string, options *LoggingParams) (interfaces.ServerSideStreamable, error) {
	stackResources, err := s.stackResourceService.GetByStackID(ctx, stackID)
	if err != nil {
		return nil, err
	}

	if len(stackResources) == 0 {
		return nil, fmt.Errorf("no resources found for stack %s", stackID)
	}

	filteredResources := make([]*models.StackResource, 0, len(stackResources))
	for _, resource := range stackResources {
		if resource.Status.State == models.StackResourcePhaseReady && resource.Status.InternalServiceName != nil {
			filteredResources = append(filteredResources, resource)
		}
	}
	if len(filteredResources) == 0 {
		return nil, fmt.Errorf("no ready resources found for stack %s", stackID)
	}

	logStreamer, serr := s.ClusterLoggingService.GetLogsForResources(ctx, orgID, filteredResources, options)
	if serr != nil {
		return nil, serr
	}

	return logStreamer, nil
}
