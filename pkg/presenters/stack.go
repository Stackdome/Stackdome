package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"k8s.io/utils/ptr"
)

func PresentStackList(stacks []*models.Stack) []openapi.Stack {
	result := make([]openapi.Stack, len(stacks))
	for i, s := range stacks {
		result[i] = PresentStack(s)
	}
	return result
}

func PresentStack(s *models.Stack) openapi.Stack {
	return openapi.Stack{
		Id:             &s.ID,
		OrganisationId: &s.OrganisationID,
		UserId:         &s.UserID,
		Name:           s.Name,
		Namespace:      &s.Namespace,
		Labels:         presentLabels(s.Labels),
		Revision:       &s.CrRevision,
		Annotations:    presentAnnotations(s.Annotations),
		Spec:           presentStackSpec(s),
		Status:         presentStackStatus(s.Status),
		CreatedAt:      &s.CreatedAt,
		UpdatedAt:      &s.UpdatedAt,
	}
}

func presentStackSpec(w *models.Stack) openapi.StackSpec {
	return openapi.StackSpec{
		StackResources: presentStackResources(w.StackResources),
		Volumes:        presentVolumes(w.Volumes, true),
	}
}

func presentStackStatus(status *models.StackStatus) *openapi.StackStatus {
	if status == nil {
		return nil
	}
	res := &openapi.StackStatus{
		State:            ptr.To(string(status.State)),
		ObservedRevision: &status.ObservedCrRevision,
		Conditions:       presentConditions(status.Conditions),
	}
	if status.LastValidationRun != nil && !status.LastValidationRun.Passed {
		res.Message = &status.LastValidationRun.Message
	}
	return res
}

func presentStackResources(resources []*models.StackResource) []openapi.StackResource {
	result := make([]openapi.StackResource, len(resources))
	for i, r := range resources {
		result[i] = presentStackResource(r)
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
		Stateful:        &r.StateFul,
		Status:          presentStackResourceStatus(r.Status),
	}
}

func presentBuildConfig(config *models.BuildConfigSpec) *openapi.StackResourceBuildSpec {
	if config == nil {
		return nil
	}
	return &openapi.StackResourceBuildSpec{
		ContextPathWithinSource: config.ContextPathWithinSource,
		DockerfilePath:          config.DockerfilePath,
		ImageRepository:         presentImageRepository(config),
		SourceContext:           presentSourceContext(config.SourceContext),
		SourceRevision:          presentSourceRevision(config.SourceRevision),
		RegistryPushSecret:      presentSecretRef(config.RegistrySecretRef),
	}
}

func presentImageRepository(in *models.BuildConfigSpec) openapi.ImageRepository {
	if in == nil {
		return openapi.ImageRepository{}
	}
	res := openapi.ImageRepository{}

	if in.BuildImageRepository.UseInClusterRegistry {
		res.UseInternalRegistry = openapi.PtrBool(true)
	} else {
		res.UseInternalRegistry = openapi.PtrBool(false)
		res.ExternalImageRepoUrl = &in.ImageRepositoryUrl
	}
	return res
}

func presentSourceRevision(revision models.BuildSourceRevision) openapi.BuildSourceRevision {
	if revision.Volume != nil {
		return openapi.BuildSourceRevision{
			VolumeSourceRevision: &openapi.BuildSourceRevisionVolumeSourceRevision{
				CurrentVolumeHash: revision.Volume.CurrentVolumeHash,
			},
		}
	} else if revision.Git != nil {
		res := openapi.BuildSourceRevision{
			GitRepoRevision: &openapi.GitRepoRevision{},
		}

		switch {
		case revision.Git.Branch != nil:
			res.GitRepoRevision.Branch = &openapi.GitRepoRevisionBranch{
				Name: &revision.Git.Branch.Name,
			}
			if revision.Git.Branch.HeadSha != "" {
				res.GitRepoRevision.Branch.HeadSha = &revision.Git.Branch.HeadSha
			}
		case revision.Git.Tag != "":
			res.GitRepoRevision.Tag = &revision.Git.Tag
		case revision.Git.Commit != "":
			res.GitRepoRevision.Commit = &revision.Git.Commit
		}
		return res
	}
	return openapi.BuildSourceRevision{}
}

func presentSourceContext(context models.BuildContextSource) openapi.BuildSourceContext {
	res := openapi.BuildSourceContext{}
	switch {
	case context.Volume != nil:
		res.Volume = &openapi.BuildSourceContextVolume{
			Id:   context.Volume.SourceVolumeID,
			Name: &context.Volume.SourceVolumeName,
		}
	case context.Git != nil:
		res.GitRepo = &openapi.BuildSourceContextGitRepo{
			RepoUrl:   context.Git.RepoURL,
			GitSecret: presentSecretRef(context.Git.GitSecretRef),
		}
	}
	return res
}

func presentImageConfig(config *models.ImageConfigSpec) *openapi.ImageSpec {
	if config == nil {
		return nil
	}
	return &openapi.ImageSpec{
		Image:      config.Image,
		PullSecret: presentSecretRef(config.PullSecretRef),
	}
}
func presentSecretRef(ref *models.SecretReference) *openapi.SecretRef {
	if ref == nil {
		return nil
	}
	return &openapi.SecretRef{
		SecretId: ref.SecretID,
	}
}

func presentInitConfig(config *models.InitConfig) *openapi.InitSpec {
	if config == nil {
		return nil
	}
	return &openapi.InitSpec{
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
			StackResourceId:  &mount.StackResourceID,
			SourceVolumeName: mount.SourceVolumeName,
			SourceVolumeType: presentVolumeSourceType(mount.SourceVolumeType),
			SourceSubPath:    &mount.SourceSubPath,
			TargetPath:       mount.TargetPath,
		}
	}
	return result
}

func presentVolumeSourceType(sourceType models.SourceVolumeType) *openapi.VolumeMountSourceType {
	switch sourceType {
	case models.BuildArtifactSyncedVolume:
		return openapi.BUILD_ARTIFACT_SYNCED_VOLUME.Ptr()
	case models.RemoteDirSyncedVolume:
		return openapi.REMOTE_DIR_SYNCED_VOLUME.Ptr()
	case models.GitRepoVolume:
		return openapi.GIT_REPO_SYNCED_VOLUME.Ptr()
	default:
		return openapi.EMPTY_VOLUME.Ptr()
	}
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
		RestartRequestTime: config.RestartRequestTime,
	}
}

func presentPorts(ports models.Ports) []openapi.Port {
	result := make([]openapi.Port, len(ports))
	for i, port := range ports {
		result[i] = openapi.Port{
			Number:          int32(port.Number),
			Protocol:        &port.Protocol,
			ExposedToPublic: port.ExposedToPublic,
			SubdomainPrefix: &port.SubdomainPrefix,
		}
	}
	return result
}

func presentResourceStatus(status *models.StackResourceStatus) *openapi.StackResourceStatus {
	if status == nil {
		return nil
	}
	return &openapi.StackResourceStatus{
		State:            ptr.To(string(status.State)),
		ObservedRevision: &status.ObservedCrRevision,
		Conditions:       presentConditions(status.Conditions),
	}
}

// ConvertWorkspace converts an API Workspace object to a model
func ConvertStack(w *openapi.Stack) *models.Stack {
	return &models.Stack{
		Name:           w.Name,
		Labels:         convertLabels(w.Labels),
		Annotations:    convertAnnotations(w.Annotations),
		StackResources: convertStackResources(w.Spec.StackResources),
		Volumes:        convertVolumes(w.Spec.Volumes),
	}
}

func convertStackResources(resources []openapi.StackResource) []*models.StackResource {
	result := make([]*models.StackResource, len(resources))
	for i, r := range resources {
		result[i] = convertStackResource(&r)
	}
	return result
}

func convertStackResource(r *openapi.StackResource) *models.StackResource {
	return &models.StackResource{
		Name:            r.Name,
		Labels:          convertLabels(r.Labels),
		Annotations:     convertAnnotations(r.Annotations),
		BuildConfig:     convertBuildConfig(r.BuildSpec),
		ImageConfig:     convertImageConfig(r.ImageSpec),
		Init:            convertInitConfig(r.InitSpec),
		ExecutionConfig: convertExecutionConfig(r.ExecutionConfig),
		VolumeMounts:    convertVolumeMounts(r.VolumeMounts),
		DependsOn:       convertDependencies(r.DependsOn),
		LifecycleConfig: convertLifecycleConfig(r.LifecycleConfig),
		Ports:           convertPorts(r.Ports),
		StateFul:        r.GetStateful(),
	}
}

func convertBuildConfig(config *openapi.StackResourceBuildSpec) *models.BuildConfigSpec {
	if config == nil {
		return nil
	}
	res := &models.BuildConfigSpec{
		ContextPathWithinSource: config.ContextPathWithinSource,
		DockerfilePath:          config.DockerfilePath,
		SourceContext:           convertSourceContext(config.SourceContext),
		SourceRevision:          convertSourceRevision(config.SourceRevision),
	}
	if config.ImageRepository.GetUseInternalRegistry() {
		res.BuildImageRepository = models.BuildImageRepository{
			UseInClusterRegistry: true,
			InsecureRegistry:     true,
		}
	} else {
		res.ImageRepositoryUrl = config.ImageRepository.GetExternalImageRepoUrl()
	}
	res.RegistrySecretRef = convertSecretRef(config.RegistryPushSecret)
	return res
}

func convertSourceContext(context openapi.BuildSourceContext) models.BuildContextSource {
	res := models.BuildContextSource{}
	switch {
	case context.Volume != nil:
		res.Volume = &models.VolumeBuildSource{
			SourceVolumeID:   context.Volume.Id,
			SourceVolumeName: context.Volume.GetName(),
		}
	case context.GitRepo != nil:
		res.Git = &models.GitBuildSource{
			RepoURL:      context.GitRepo.RepoUrl,
			GitSecretRef: convertSecretRef(context.GitRepo.GitSecret),
		}
	}
	return res
}

func convertSourceRevision(revision openapi.BuildSourceRevision) models.BuildSourceRevision {
	if revision.VolumeSourceRevision != nil {
		return models.BuildSourceRevision{
			Volume: &models.VolumeRevision{
				CurrentVolumeHash: revision.VolumeSourceRevision.CurrentVolumeHash,
			},
		}
	} else if revision.GitRepoRevision != nil {
		res := models.BuildSourceRevision{
			Git: &models.GitRevision{},
		}
		switch {
		case revision.GitRepoRevision.Branch != nil:
			res.Git = &models.GitRevision{
				Branch: &models.GitBranch{
					Name:    revision.GitRepoRevision.Branch.GetName(),
					HeadSha: revision.GitRepoRevision.Branch.GetHeadSha(),
				},
			}
		case revision.GitRepoRevision.Tag != nil:
			res.Git = &models.GitRevision{
				Tag: revision.GitRepoRevision.GetTag(),
			}
		case revision.GitRepoRevision.Commit != nil:
			res.Git = &models.GitRevision{
				Commit: revision.GitRepoRevision.GetCommit(),
			}
		}
		return res
	}
	return models.BuildSourceRevision{}
}

func convertImageConfig(config *openapi.ImageSpec) *models.ImageConfigSpec {
	if config == nil {
		return nil
	}

	return &models.ImageConfigSpec{
		Image:         config.Image,
		PullSecretRef: convertSecretRef(config.PullSecret),
	}
}

func convertSecretRef(ref *openapi.SecretRef) *models.SecretReference {
	if ref == nil {
		return nil
	}
	return &models.SecretReference{
		SecretID: ref.SecretId,
	}
}

func convertInitConfig(config *openapi.InitSpec) *models.InitConfig {
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
		Command:            config.Command,
		Args:               config.Args,
		Env:                convertEnvVars(config.EnvironmentVariables),
		EnvVarsFromSecrets: convertEnvVarsFromSecret(config.EnvironmentVariablesFromSecret),
	}
}

func convertEnvVarsFromSecret(envVars []openapi.EnvVarFromSecret) []models.EnvSecretReference {
	if envVars == nil {
		return nil
	}
	result := make([]models.EnvSecretReference, len(envVars))
	for i, env := range envVars {
		result[i] = models.EnvSecretReference{
			SecretID:  env.SecretRef.SecretId,
			SecretKey: env.Key,
			EnvName:   env.Name,
		}
	}
	return result

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
			SourceVolumeName: mount.SourceVolumeName,
			SourceSubPath:    mount.GetSourceSubPath(),
			TargetPath:       mount.TargetPath,
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
		RestartRequestTime: config.RestartRequestTime,
	}
}

func convertPorts(ports []openapi.Port) []models.Port {
	result := make([]models.Port, len(ports))
	for i, port := range ports {
		result[i] = models.Port{
			Number: int(port.Number),
		}
		result[i].Protocol = port.GetProtocol()
		result[i].ExposedToPublic = port.GetExposedToPublic()
		result[i].SubdomainPrefix = port.GetSubdomainPrefix()
	}
	return models.Ports(result)
}
