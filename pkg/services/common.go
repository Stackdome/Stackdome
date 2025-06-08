package services

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/services/clusterresource"
	"github.com/ashishmax31/stackdome-api-server/pkg/worker/workermanager"
)

type ClusterResourceServiceInjectable interface {
	InjectClusterResourceServiceDeps(deps ClusterResourceServiceDeps)
}

// To be emdedded in services that require cluster resource service dependencies.
type ClusterResourceServiceDeps struct {
	ClusterStackService     clusterresource.ClusterStackService
	ClusterNamespaceService clusterresource.NamespaceClusterResourceService
	ClusterVolumeService    clusterresource.VolumeClusterResourceService
	ClusterLoggingService   clusterresource.ClusterLoggingService
	ClusterMetricsService   clusterresource.ClusterMetricsService
}

func (s *ClusterResourceServiceDeps) InjectClusterResourceServiceDeps(deps ClusterResourceServiceDeps) {
	s.ClusterStackService = deps.ClusterStackService
	s.ClusterNamespaceService = deps.ClusterNamespaceService
	s.ClusterVolumeService = deps.ClusterVolumeService
	s.ClusterLoggingService = deps.ClusterLoggingService
	s.ClusterMetricsService = deps.ClusterMetricsService
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
