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
	return presentStackResource(r)
}

func presentStackResourceStatus(status *models.StackResourceStatus) *openapi.StackResourceStatus {
	if status == nil {
		return nil
	}
	res := &openapi.StackResourceStatus{
		State:                         ptr.To(string(status.State)),
		ObservedRevision:              &status.ObservedCrRevision,
		Conditions:                    presentConditions(status.Conditions),
		PublicIngress:                 presentIngress(status.PublicIngresses),
		InternalServiceName:           status.InternalServiceName,
		LastRestartRequestProcessedAt: status.LastRestartRequestProcessedAt,
		LastFailure:                   presentStackResourceFailure(status.LastFailure),
		Replicas:                      &status.Replicas,
		AvailableReplicas:             &status.AvailableReplicas,
		UpdatedReplicas:               &status.UpdatedReplicas,
		LastRunTime:                   status.LastRunTime,
		LastRunSucceeded:              status.LastRunSucceeded,
	}
	return res
}

func presentStackResourceFailure(f *models.StackResourceFailure) *openapi.StackResourceFailure {
	if f == nil {
		return nil
	}
	return &openapi.StackResourceFailure{
		Type:          ptr.To(string(f.Type)),
		Container:     presentContainerFailureDetail(f.Container),
		InitContainer: presentContainerFailureDetail(f.InitContainer),
		Build:         presentBuildFailureDetail(f.Build),
	}
}

func presentContainerFailureDetail(d *models.ContainerFailureDetail) *openapi.ContainerFailureDetail {
	if d == nil {
		return nil
	}
	return &openapi.ContainerFailureDetail{
		FailureType:  ptr.To(d.FailureType),
		Reason:       ptr.To(d.Reason),
		Message:      ptr.To(d.Message),
		RestartCount: ptr.To(d.RestartCount),
		ExitCode:     d.ExitCode,
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

func presentOutputDescriptors(outputs []models.OutputDescriptor) []openapi.OutputDescriptor {
	result := make([]openapi.OutputDescriptor, len(outputs))
	for i, output := range outputs {
		result[i] = openapi.OutputDescriptor{
			Name:      output.Name,
			Type:      string(output.Type),
			Sensitive: output.Sensitive,
		}
	}
	return result
}

func presentStackResource(r *models.StackResource) openapi.StackResource {
	return openapi.StackResource{
		Id:              &r.ID,
		StackId:         &r.StackID,
		Name:            r.Name,
		Labels:          presentLabels(r.Labels),
		Annotations:     presentAnnotations(r.Annotations),
		BuildSpec:       presentBuildConfig(r.BuildConfig),
		ImageSpec:       presentImageConfig(r.ImageConfig),
		InitSpec:        presentInitConfig(r.Init),
		ExecutionConfig: presentExecutionConfig(r.ExecutionConfig),
		VolumeMounts:    presentVolumeMounts(r.VolumeMounts),
		DependsOn:       presentDependencies(r.DependsOn),
		LifecycleConfig: presentLifecycleConfig(r.LifecycleConfig),
		Ports:           presentPorts(r.Ports),
		Outputs:         presentOutputDescriptors(r.EnsureDeclaredOutputs()),
		WorkloadType:    (*string)(&r.WorkloadType),
		Schedule:        &r.Schedule,
		Replicas:        r.Replicas,
		Status:          presentStackResourceStatus(r.Status),
	}
}
