package services

import (
	"github.com/Stackdome/stackdome/pkg/services/clusterresource"
	"github.com/Stackdome/stackdome/pkg/worker/workermanager"
)

// Aliases: the deps type and injectable interface are defined in the
// clusterresource leaf package so that generated mocks of services
// interfaces embedding ClusterResourceServiceInjectable (pkg/mocks) import
// that leaf instead of pkg/services — importing pkg/services from pkg/mocks
// would close an import cycle through the in-package tests of pkg/services
// and its dependents.
type ClusterResourceServiceInjectable = clusterresource.ClusterResourceServiceInjectable

type ClusterResourceServiceDeps = clusterresource.ClusterResourceServiceDeps

type BackgroundJobEnqueuerDep struct {
	BackgroundJobEnqueuer workermanager.BackgroundJobEnqueuer
}

type BackgroundJobEnqueuerInjectable interface {
	InjectBackgroundJobEnqueuer(dep BackgroundJobEnqueuerDep)
}

func (s *BackgroundJobEnqueuerDep) InjectBackgroundJobEnqueuer(dep BackgroundJobEnqueuerDep) {
	s.BackgroundJobEnqueuer = dep.BackgroundJobEnqueuer
}
