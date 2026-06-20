package services

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
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
	return resource, nil
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
