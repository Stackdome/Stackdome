package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"k8s.io/utils/ptr"
)

func PresentWorkspaceResourceList(resources []*models.WorkspaceResource) []openapi.WorkspaceResource {
	result := make([]openapi.WorkspaceResource, len(resources))
	for i, r := range resources {
		result[i] = PresentWorkspaceResource(r)
	}
	return result
}

func PresentWorkspaceResource(r *models.WorkspaceResource) openapi.WorkspaceResource {
	return openapi.WorkspaceResource{
		Id:              &r.ID,
		WorkspaceId:     &r.WorkspaceID,
		Name:            r.Name,
		Labels:          presentLabels(r.Labels),
		Annotations:     presentAnnotations(r.Annotations),
		Version:         openapi.PtrInt32(int32(r.Version)),
		ImageRegistry:   r.ImageRegistry,
		Build:           presentBuildConfig(r.Build),
		Prebuilt:        presentPrebuiltConfig(r.Prebuilt),
		Init:            presentInitConfig(r.Init),
		ExecutionConfig: presentExecutionConfig(r.ExecutionConfig),
		VolumeMounts:    presentVolumeMounts(r.VolumeMounts),
		DependsOn:       presentDependencies(r.DependsOn),
		LifecycleConfig: presentLifecycleConfig(r.LifecycleConfig),
		Ports:           presentPorts(r.Ports),
		Stateful:        &r.StateFul,
		Status:          presentWorkspaceResourceStatus(r.Status),
	}
}

func presentWorkspaceResourceStatus(status *models.WorkspaceResourceStatus) *openapi.ResourceStatus {
	if status == nil {
		return nil
	}
	return &openapi.ResourceStatus{
		State:               &status.State,
		ObservedVersion:     ptr.To(int32(status.ObservedVersion)),
		Conditions:          presentConditions(status.Conditions),
		PublicIngress:       presentIngress(status.PublicIngresses),
		InternalServiceName: status.InternalServiceName,
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
