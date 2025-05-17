package services

import "github.com/ashishmax31/stackdome-api-server/pkg/services/clusterresource"

type ClusterResourceServiceInjectable interface {
	InjectClusterResourceServiceDeps(deps ClusterResourceServiceDeps)
}

type ClusterResourceServiceDeps struct {
	ClusterStackService     clusterresource.ClusterStackService
	NamespaceClusterService clusterresource.NamespaceClusterResourceService
	VolumeClusterService    clusterresource.VolumeClusterResourceService
}
