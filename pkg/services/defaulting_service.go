package services

import (
	"github.com/Stackdome/stackdome/pkg/models"
)

type DefaultingService[T any] interface {
	PopulateDefaultValues(resource T) (T, error)
}

type stackDefaultingService struct{}

func NewStackDefaultingService() DefaultingService[*models.Stack] {
	return &stackDefaultingService{}
}

func (s *stackDefaultingService) PopulateDefaultValues(resource *models.Stack) (*models.Stack, error) {
	for i := range resource.StackResources {
		applyStackResourceWorkloadTypeDefault(resource.StackResources[i])
		applyStackResourcePortDefaults(resource.StackResources[i])
		normalizeStackResourceReplicas(resource.StackResources[i])
	}
	applyStackSettingsDefaults(resource)
	return resource, nil
}

func applyStackSettingsDefaults(stack *models.Stack) {
	if stack.Settings == nil {
		stack.Settings = &models.StackSettings{}
	}
	if stack.Settings.ReleaseRetentionLimit <= 0 {
		stack.Settings.ReleaseRetentionLimit = models.DefaultReleaseRetentionLimit
	}
	if stack.Settings.MinSuccessfulReleases <= 0 {
		stack.Settings.MinSuccessfulReleases = models.DefaultMinSuccessfulReleases
	}
}

func applyStackResourcePortDefaults(resource *models.StackResource) {
	if resource == nil || len(resource.Ports) == 0 {
		return
	}
	for i := range resource.Ports {
		if resource.Ports[i].ExposedToPublic && resource.Ports[i].Protocol == "" {
			resource.Ports[i].Protocol = models.PortProtocolHTTP
		}
	}
}

// applyStackResourceWorkloadTypeDefault fills in the OpenAPI default for
// workload_type. Stackfile-built stacks (PR previews, compose import) omit it,
// and the enum has no value for "" — an empty one fails client-side validation
// and reads as a deleted resource. Runs before the replica normalization below
// so both see the same workload type.
func applyStackResourceWorkloadTypeDefault(resource *models.StackResource) {
	if resource == nil || resource.WorkloadType != "" {
		return
	}
	resource.WorkloadType = models.WorkloadTypeService
}

// singleInstanceReplicas is the replica count hard-coded for workloads that
// always run exactly one instance per (scheduled) run.
const singleInstanceReplicas = int32(1)

// normalizeStackResourceReplicas hard-codes replicas to 1 for Job and
// CronJob workloads regardless of what the client sent: those workloads
// always run a single instance per run, so the persisted value stays
// truthful instead of carrying a count that would never apply.
func normalizeStackResourceReplicas(resource *models.StackResource) {
	if resource == nil {
		return
	}
	switch resource.WorkloadType {
	case models.WorkloadTypeJob, models.WorkloadTypeCronJob:
		replicas := singleInstanceReplicas
		resource.Replicas = &replicas
	}
}
