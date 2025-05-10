package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"k8s.io/utils/ptr"
)

func PresentStackResourceList(resources []*models.StackResource) []openapi.StackResource {
	result := make([]openapi.StackResource, len(resources))
	for i, r := range resources {
		result[i] = PresentStackResource(r)
	}
	return result
}

func PresentStackResource(r *models.StackResource) openapi.StackResource {
	return openapi.StackResource{
		Id:              &r.ID,
		StackId:         &r.StackID,
		Name:            r.Name,
		Labels:          presentLabels(r.Labels),
		Annotations:     presentAnnotations(r.Annotations),
		Version:         openapi.PtrInt32(int32(r.Version)),
		BuildSpec:       presentBuildConfig(r.BuildConfig),
		ImageSpec:       presentImageConfig(r.ImageConfig),
		InitSpec:        presentInitConfig(r.Init),
		ExecutionConfig: presentExecutionConfig(r.ExecutionConfig),
		VolumeMounts:    presentVolumeMounts(r.VolumeMounts),
		DependsOn:       presentDependencies(r.DependsOn),
		LifecycleConfig: presentLifecycleConfig(r.LifecycleConfig),
		Ports:           presentPorts(r.Ports),
		Stateful:        &r.StateFul,
		Status:          presentStackResourceStatus(r.Status),
	}
}

func presentStackResourceStatus(status *models.StackResourceStatus) *openapi.StackResourceStatus {
	if status == nil {
		return nil
	}
	return &openapi.StackResourceStatus{
		State:                         &status.State,
		ObservedVersion:               ptr.To(int32(status.ObservedVersion)),
		Conditions:                    presentConditions(status.Conditions),
		PublicIngress:                 presentIngress(status.PublicIngresses),
		InternalServiceName:           status.InternalServiceName,
		LastRestartRequestProcessedAt: status.LastRestartRequestProcessedAt,
	}
}

func presentIngress(ingresses []models.Ingress) []openapi.Ingress {
	result := make([]openapi.Ingress, len(ingresses))
	for i, ingress := range ingresses {
		result[i] = openapi.Ingress{
			Url:        &ingress.URL,
			TargetPort: ptr.To(int32(ingress.TargetPort)),
		}
	}
	return result
}
