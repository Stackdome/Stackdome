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
		applyStackResourcePortDefaults(resource.StackResources[i])
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
	if stack.Settings.DeployTimeoutMinutes <= 0 {
		stack.Settings.DeployTimeoutMinutes = models.DefaultDeployTimeoutMinutes
	}
}

func applyStackResourcePortDefaults(resource *models.StackResource) {
	if resource == nil || len(resource.Ports) == 0 {
		return
	}
	for i := range resource.Ports {
		if resource.Ports[i].ExposedToPublic && resource.Ports[i].Protocol == "" {
			resource.Ports[i].Protocol = "http"
		}
	}
}
