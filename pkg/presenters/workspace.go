package presenters

import (
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func PresentWorkspaceList(workspaces []*models.Workspace) []openapi.Workspace {
	result := make([]openapi.Workspace, len(workspaces))
	for i, w := range workspaces {
		result[i] = PresentWorkspace(w)
	}
	return result
}

func PresentWorkspace(w *models.Workspace) openapi.Workspace {
	return openapi.Workspace{
		Id:             &w.ID,
		OrganisationId: &w.OrganisationID,
		UserId:         &w.UserID,
		Name:           w.Name,
		Namespace:      &w.Namespace,
		Labels:         presentLabels(w.Labels),
		Annotations:    presentAnnotations(w.Annotations),
		Version:        openapi.PtrInt32(int32(w.Version)),
		Spec:           presentWorkspaceSpec(w),
		Status:         presentWorkspaceStatus(w.Status),
		CreatedAt:      &w.CreatedAt,
		UpdatedAt:      &w.UpdatedAt,
	}
}

func presentWorkspaceSpec(w *models.Workspace) openapi.WorkspaceSpec {
	return openapi.WorkspaceSpec{
		Resources: presentWorkspaceResources(w.WorkspaceResources),
	}
}

func presentWorkspaceStatus(status *models.WorkspaceStatus) *openapi.WorkspaceStatus {
	if status == nil {
		return nil
	}
	return &openapi.WorkspaceStatus{
		State:           &status.State,
		ObservedVersion: openapi.PtrInt32(int32(status.ObservedVersion)),
		Conditions:      presentConditions(status.Conditions),
	}
}

func presentWorkspaceResources(resources []*models.WorkspaceResource) []openapi.WorkspaceResource {
	result := make([]openapi.WorkspaceResource, len(resources))
	for i, r := range resources {
		result[i] = presentWorkspaceResource(r)
	}
	return result
}

func presentWorkspaceResource(r *models.WorkspaceResource) openapi.WorkspaceResource {
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
		Status:          presentResourceStatus(r.Status),
	}
}

func presentBuildConfig(config *models.BuildConfig) *openapi.BuildConfig {
	if config == nil {
		return nil
	}
	return &openapi.BuildConfig{
		SourceVolumeId: config.SourceVolumeID,
		ContextPath:    config.ContextPath,
		DockerfilePath: config.DockerfilePath,
		SourceHash:     config.SourceHash,
	}
}

func presentPrebuiltConfig(config *models.PrebuiltConfig) *openapi.PrebuiltConfig {
	if config == nil {
		return nil
	}
	return &openapi.PrebuiltConfig{
		Image: config.ImageName + ":" + config.Tag,
	}
}

func presentInitConfig(config *models.InitConfig) *openapi.InitConfig {
	if config == nil {
		return nil
	}
	return &openapi.InitConfig{
		Command: config.Command,
		Args:    config.Args,
	}
}

func presentExecutionConfig(config *models.ExecutionConfig) *openapi.ExecutionConfig {
	if config == nil {
		return nil
	}
	return &openapi.ExecutionConfig{
		Command:              config.Command,
		Args:                 config.Args,
		EnvironmentVariables: presentEnvVars(config.Env),
	}
}

func presentEnvVars(envVars []models.EnvVar) []openapi.EnvVar {
	result := make([]openapi.EnvVar, len(envVars))
	for i, env := range envVars {
		result[i] = openapi.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		}
	}
	return result
}

func presentVolumeMounts(mounts []*models.VolumeMount) []openapi.VolumeMount {
	result := make([]openapi.VolumeMount, len(mounts))
	for i, mount := range mounts {
		result[i] = openapi.VolumeMount{
			WorkspaceResourceId: &mount.WorkspaceResourceID,
			WorkspaceStorageId:  &mount.WorkspaceStorageID,
			SourceVolumeId:      mount.SourceVolumeID,
			SourceSubPath:       &mount.SourceSubPath,
			TargetPath:          mount.TargetPath,
		}
	}
	return result
}

func presentDependencies(deps models.Dependencies) []string {
	if deps == nil {
		return nil
	}
	var result []string
	for _, dep := range deps {
		result = append(result, dep)
	}
	return result
}

func presentLifecycleConfig(config *models.LifecycleConfig) *openapi.LifecycleConfig {
	if config == nil {
		return nil
	}
	return &openapi.LifecycleConfig{
		LastRestartRequestTime: config.LastRestartRequestTime,
	}
}

func presentPorts(ports models.Ports) []openapi.Port {
	result := make([]openapi.Port, len(ports))
	for i, port := range ports {
		result[i] = openapi.Port{
			Number:          int32(port.Number),
			Protocol:        &port.Protocol,
			ExposedToPublic: &port.ExposedToPublic,
			SubdomainPrefix: &port.SubdomainPrefix,
		}
	}
	return result
}

func presentResourceStatus(status *models.WorkspaceResourceStatus) *openapi.ResourceStatus {
	if status == nil {
		return nil
	}
	return &openapi.ResourceStatus{
		State:           &status.State,
		ObservedVersion: openapi.PtrInt32(int32(status.ObservedVersion)),
		Conditions:      presentConditions(status.Conditions),
	}
}

// ConvertWorkspace converts an API Workspace object to a model
func ConvertWorkspace(w *openapi.Workspace) *models.Workspace {
	return &models.Workspace{
		Name:               w.Name,
		Labels:             convertLabels(w.Labels),
		Annotations:        convertAnnotations(w.Annotations),
		WorkspaceResources: convertWorkspaceResources(w.Spec.Resources),
	}
}

func convertWorkspaceResources(resources []openapi.WorkspaceResource) []*models.WorkspaceResource {
	result := make([]*models.WorkspaceResource, len(resources))
	for i, r := range resources {
		result[i] = convertWorkspaceResource(&r)
	}
	return result
}

func convertWorkspaceResource(r *openapi.WorkspaceResource) *models.WorkspaceResource {
	return &models.WorkspaceResource{
		Name:            r.Name,
		Labels:          convertLabels(r.Labels),
		Annotations:     convertAnnotations(r.Annotations),
		ImageRegistry:   r.ImageRegistry,
		Build:           convertBuildConfig(r.Build),
		Prebuilt:        convertPrebuiltConfig(r.Prebuilt),
		Init:            convertInitConfig(r.Init),
		ExecutionConfig: convertExecutionConfig(r.ExecutionConfig),
		VolumeMounts:    convertVolumeMounts(r.VolumeMounts),
		DependsOn:       convertDependencies(r.DependsOn),
		LifecycleConfig: convertLifecycleConfig(r.LifecycleConfig),
		Ports:           convertPorts(r.Ports),
		StateFul:        r.GetStateful(),
	}
}

func convertBuildConfig(config *openapi.BuildConfig) *models.BuildConfig {
	if config == nil {
		return nil
	}
	return &models.BuildConfig{
		SourceVolumeID: config.SourceVolumeId,
		ContextPath:    config.ContextPath,
		DockerfilePath: config.DockerfilePath,
		SourceHash:     config.SourceHash,
	}
}

func convertPrebuiltConfig(config *openapi.PrebuiltConfig) *models.PrebuiltConfig {
	if config == nil {
		return nil
	}
	// Assuming the image is in the format "imageName:tag"
	parts := strings.SplitN(config.Image, ":", 2)
	imageName := parts[0]
	tag := "latest"
	if len(parts) > 1 {
		tag = parts[1]
	}
	return &models.PrebuiltConfig{
		ImageName: imageName,
		Tag:       tag,
	}
}

func convertInitConfig(config *openapi.InitConfig) *models.InitConfig {
	if config == nil {
		return nil
	}
	return &models.InitConfig{
		Command: config.Command,
		Args:    config.Args,
	}
}

func convertExecutionConfig(config *openapi.ExecutionConfig) *models.ExecutionConfig {
	if config == nil {
		return nil
	}
	return &models.ExecutionConfig{
		Command: config.Command,
		Args:    config.Args,
		Env:     convertEnvVars(config.EnvironmentVariables),
	}
}

func convertEnvVars(envVars []openapi.EnvVar) []models.EnvVar {
	result := make([]models.EnvVar, len(envVars))
	for i, env := range envVars {
		result[i] = models.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		}
	}
	return result
}

func convertVolumeMounts(mounts []openapi.VolumeMount) []*models.VolumeMount {
	result := make([]*models.VolumeMount, len(mounts))
	for i, mount := range mounts {
		result[i] = &models.VolumeMount{
			SourceVolumeID: mount.SourceVolumeId,
			SourceSubPath:  mount.GetSourceSubPath(),
			TargetPath:     mount.TargetPath,
		}
	}
	return result
}

func convertDependencies(deps []string) models.Dependencies {
	if deps == nil {
		return nil
	}
	result := models.Dependencies(deps)
	return result
}

func convertLifecycleConfig(config *openapi.LifecycleConfig) *models.LifecycleConfig {
	if config == nil {
		return nil
	}
	return &models.LifecycleConfig{
		LastRestartRequestTime: config.LastRestartRequestTime,
	}
}

func convertPorts(ports []openapi.Port) []models.Port {
	result := make([]models.Port, len(ports))
	for i, port := range ports {
		result[i] = models.Port{
			Number:          int(port.Number),
			Protocol:        *port.Protocol,
			ExposedToPublic: *port.ExposedToPublic,
		}
	}
	return models.Ports(result)
}
