package services

import (
	"context"
	stderrors "errors"
	"fmt"
	"net/url"
	"strconv"

	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/interfaces"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services/clusterresource"
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

//go:generate mockgen -destination=image_build_service_mock.go -package=services github.com/Stackdome/stackdome/pkg/services ImageBuildService
//go:generate mockgen -destination=cluster_logging_service_mock.go -package=services github.com/Stackdome/stackdome/pkg/services/clusterresource ClusterLoggingService

type LoggingService interface {
	StreamLogsForStackResource(ctx context.Context, orgID string, stackID string, stackResourceName string, options *LoggingParams) (interfaces.ServerSideStreamable, error)
	StreamLogsForStack(ctx context.Context, orgID string, stackID string, options *LoggingParams) (interfaces.ServerSideStreamable, error)
	StreamLogsForBuild(ctx context.Context, orgID string, buildID string, options *LoggingParams) (interfaces.ServerSideStreamable, *errors.ServiceError)
	ClusterResourceServiceInjectable
}

type loggingService struct {
	clusterService       ClusterService
	stackResourceService StackResourceService
	imageBuildService    ImageBuildService
	logger               logger.Logger
	ClusterResourceServiceDeps
}

type LoggingServiceSpec struct {
	ClusterService       ClusterService
	StackResourceService StackResourceService
	ImageBuildService    ImageBuildService
	Logger               logger.Logger
}

func NewLoggingService(spec LoggingServiceSpec) LoggingService {
	return &loggingService{
		clusterService:       spec.ClusterService,
		stackResourceService: spec.StackResourceService,
		imageBuildService:    spec.ImageBuildService,
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

func (s *loggingService) StreamLogsForBuild(ctx context.Context, orgID string, buildID string, options *LoggingParams) (interfaces.ServerSideStreamable, *errors.ServiceError) {
	build, err := s.imageBuildService.GetByID(ctx, buildID)
	if err != nil {
		return nil, err
	}

	if build.Status == nil || build.Status.BuildSourceRevision == "" {
		return nil, errors.Conflict("build %s has not started yet", buildID)
	}

	jobName := buildsv1alpha1.BuildJobName(build.StackResourceName, build.Status.BuildSourceRevision)

	streamer, cerr := s.ClusterLoggingService.GetLogsForBuildPod(ctx, orgID, build.Namespace, jobName, options)
	if cerr != nil {
		if stderrors.Is(cerr, clusterresource.ErrBuildPodNotFound) {
			return nil, errors.NotFound("no logs available for build %s: %s", buildID, cerr.Error())
		}
		return nil, errors.GeneralError("failed to get logs for build %s: %s", buildID, cerr.Error())
	}
	return streamer, nil
}
