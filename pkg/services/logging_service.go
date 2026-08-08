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
	StreamLogsForStackResource(ctx context.Context, orgID string, stackID string, stackResourceName string, options *LoggingParams) (interfaces.ServerSideStreamable, *errors.ServiceError)
	StreamLogsForStack(ctx context.Context, orgID string, stackID string, options *LoggingParams) (interfaces.ServerSideStreamable, *errors.ServiceError)
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

// Logs are best-effort: we stream them whenever the workload is known, whatever
// state the resource is in. The durable diagnostics live in the status API.
func (s *loggingService) StreamLogsForStackResource(ctx context.Context, orgID string, stackID string, stackResourceID string, options *LoggingParams) (interfaces.ServerSideStreamable, *errors.ServiceError) {
	stackResource, err := s.stackResourceService.GetByStackIDAndResourceName(ctx, stackID, stackResourceID)
	if err != nil {
		return nil, err
	}

	logStreamer, cerr := s.ClusterLoggingService.GetLogsForResources(ctx, orgID, []*models.StackResource{stackResource}, options)
	if cerr != nil {
		return nil, logStreamError(stackResource.Name, cerr)
	}

	return logStreamer, nil
}

func (s *loggingService) StreamLogsForStack(ctx context.Context, orgID string, stackID string, options *LoggingParams) (interfaces.ServerSideStreamable, *errors.ServiceError) {
	stackResources, err := s.stackResourceService.GetByStackID(ctx, stackID)
	if err != nil {
		return nil, err
	}

	if len(stackResources) == 0 {
		return nil, errors.NotFound("no resources found for stack %s", stackID)
	}

	logStreamer, cerr := s.ClusterLoggingService.GetLogsForResources(ctx, orgID, stackResources, options)
	if cerr != nil {
		return nil, logStreamError(stackID, cerr)
	}

	return logStreamer, nil
}

// A missing pod or workload is a 404; anything else is a real failure.
func logStreamError(target string, err error) *errors.ServiceError {
	if stderrors.Is(err, clusterresource.ErrNoWorkload) {
		return errors.NotFound("no logs available for %s: %s", target, err.Error())
	}
	return errors.GeneralError("failed to get logs for %s: %s", target, err.Error())
}

func (s *loggingService) StreamLogsForBuild(ctx context.Context, orgID string, buildID string, options *LoggingParams) (interfaces.ServerSideStreamable, *errors.ServiceError) {
	build, err := s.imageBuildService.GetByID(ctx, buildID)
	if err != nil {
		return nil, err
	}

	if !build.Status.IsConditionTrue(string(buildsv1alpha1.BuildJobCreated)) {
		return nil, errors.Conflict("build job for %s has not been created yet", buildID)
	}

	// build.ID holds the ImageBuild CR name; see createImageBuildInDB.
	streamer, cerr := s.ClusterLoggingService.GetLogsForBuildPod(ctx, orgID, build.Namespace, build.ID, options)
	if cerr != nil {
		switch {
		case stderrors.Is(cerr, clusterresource.ErrBuildPodNotFound):
			return nil, errors.NotFound("no logs available for build %s: %s", buildID, cerr.Error())
		case stderrors.Is(cerr, clusterresource.ErrBuildPodNotReady):
			return nil, errors.Conflict("build pod for %s is starting; logs are not available yet", buildID)
		default:
			return nil, errors.GeneralError("failed to get logs for build %s: %s", buildID, cerr.Error())
		}
	}
	return streamer, nil
}
