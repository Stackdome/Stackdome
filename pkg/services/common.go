package services

import "github.com/ashishmax31/stackdome-api-server/pkg/services/clusterresource"

type ClusterResourceServiceInjectable interface {
	InjectClusterResourceServiceDeps(deps ClusterResourceServiceDeps)
}

// To be emdedded in services that require cluster resource service dependencies.
type ClusterResourceServiceDeps struct {
	ClusterStackService     clusterresource.ClusterStackService
	ClusterNamespaceService clusterresource.NamespaceClusterResourceService
	ClusterVolumeService    clusterresource.VolumeClusterResourceService
	ClusterLoggingService   clusterresource.ClusterLoggingService
}

func (s *ClusterResourceServiceDeps) InjectClusterResourceServiceDeps(deps ClusterResourceServiceDeps) {
	s.ClusterStackService = deps.ClusterStackService
	s.ClusterNamespaceService = deps.ClusterNamespaceService
	s.ClusterVolumeService = deps.ClusterVolumeService
	s.ClusterLoggingService = deps.ClusterLoggingService
}
