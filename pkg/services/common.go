package services

import (
	"github.com/Stackdome/stackdome/pkg/services/clusterresource"
	"github.com/Stackdome/stackdome/pkg/worker/workermanager"
)

type ClusterResourceServiceInjectable interface {
	InjectClusterResourceServiceDeps(deps ClusterResourceServiceDeps)
}

// To be emdedded in services that require cluster resource service dependencies.
type ClusterResourceServiceDeps struct {
	ClusterNamespaceService clusterresource.NamespaceClusterResourceService
	ClusterVolumeService    clusterresource.VolumeClusterResourceService
	ClusterLoggingService   clusterresource.ClusterLoggingService
	ClusterMetricsService   clusterresource.ClusterMetricsService
}

func (s *ClusterResourceServiceDeps) InjectClusterResourceServiceDeps(deps ClusterResourceServiceDeps) {
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
