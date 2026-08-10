package clusterresource

import (
	"context"
	"fmt"
	"time"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/worker/workermanager"
)

const (
	DefaultLogStreamBufferSize     = 1000
	DefaultStreamTimeoutDuration   = 20 * time.Minute
	DefaultLogStreamRateLimit      = 200 * time.Millisecond
	DefaultLogStreamRateLimitBurst = 100
	MaxLogStreamErrors             = 50
)

type ClusterResourceError struct {
	Message string
	Err     error
}

func (e *ClusterResourceError) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
}

func newError(message string, err error) *ClusterResourceError {
	return &ClusterResourceError{
		Message: message,
		Err:     err,
	}
}

func WrapErrAsServiceError(err error) *errors.ServiceError {
	if err == nil {
		return nil
	}
	return errors.GeneralError("%s", err.Error())
}

// ClusterResourceServiceInjectable is implemented (by embedding
// ClusterResourceServiceDeps) by services that need cluster resource service
// dependencies injected. It lives in this leaf package — not pkg/services —
// so that mocks of services interfaces embedding it (pkg/mocks) never import
// pkg/services, which would close an import cycle through the in-package
// tests of pkg/services and its dependents.
type ClusterResourceServiceInjectable interface {
	InjectClusterResourceServiceDeps(deps ClusterResourceServiceDeps)
}

type BackgroundJobEnqueuerDep struct {
	BackgroundJobEnqueuer workermanager.BackgroundJobEnqueuer
}

type BackgroundJobEnqueuerInjectable interface {
	InjectBackgroundJobEnqueuer(dep BackgroundJobEnqueuerDep)
}

func (s *BackgroundJobEnqueuerDep) InjectBackgroundJobEnqueuer(dep BackgroundJobEnqueuerDep) {
	s.BackgroundJobEnqueuer = dep.BackgroundJobEnqueuer
}

// ClusterResourceServiceDeps is embedded in services that require cluster
// resource service dependencies.
type ClusterResourceServiceDeps struct {
	ClusterNamespaceService NamespaceClusterResourceService
	ClusterVolumeService    VolumeClusterResourceService
	ClusterLoggingService   ClusterLoggingService
	ClusterMetricsService   ClusterMetricsService
}

func (s *ClusterResourceServiceDeps) InjectClusterResourceServiceDeps(deps ClusterResourceServiceDeps) {
	s.ClusterNamespaceService = deps.ClusterNamespaceService
	s.ClusterVolumeService = deps.ClusterVolumeService
	s.ClusterLoggingService = deps.ClusterLoggingService
	s.ClusterMetricsService = deps.ClusterMetricsService
}

type DBClusterService interface {
	GetClusterForOrg(ctx context.Context, orgID string) (*models.Cluster, *errors.ServiceError)
}

type DBSecretService interface {
	InternalGetByID(ctx context.Context, secretID string) (*models.Secret, *errors.ServiceError)
}

type DBOrganisationService interface {
	Get(ctx context.Context, ID string) (*models.Organisation, *errors.ServiceError)
}

type DBStackResourceService interface {
	GetByStackResourceID(ctx context.Context, stackResourceID string) (*models.StackResource, *errors.ServiceError)
	GetByStackID(ctx context.Context, stackID string) ([]*models.StackResource, *errors.ServiceError)
}
