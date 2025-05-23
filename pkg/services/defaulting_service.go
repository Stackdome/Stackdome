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
		if len(resource.StackResources[i].Ports) > 0 {
			for j := range resource.StackResources[i].Ports {
				if resource.StackResources[i].Ports[j].ExposedToPublic {
					resource.StackResources[i].Ports[j].Protocol = "http"
				}
			}
		}
	}
	return resource, nil
}
