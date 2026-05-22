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
		TeamId:         &s.TeamID,
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
		Connections:    presentConnections(w.Connections),
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
		EnvFromAddons:        presentEnvFromAddons(config.EnvFromAddons),
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

func presentEnvFromAddons(sources []models.AddonEnvSource) []openapi.AddonEnvSource {
	if sources == nil {
		return nil
	}
	result := make([]openapi.AddonEnvSource, len(sources))
	for i, s := range sources {
		if s.Postgres != nil {
			pgSource := &openapi.PostgresAddonEnvSource{
				AddonId:    s.Postgres.AddonID,
				EnvMapping: s.Postgres.EnvMapping,
			}
			if s.Postgres.Database != "" {
				pgSource.SetDatabase(s.Postgres.Database)
			}
			if s.Postgres.Superuser {
				pgSource.SetSuperuser(true)
			}
			result[i] = openapi.AddonEnvSource{
				Postgres: pgSource,
			}
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
			Name:            port.Name,
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
		LastFailure:      presentStackResourceFailure(status.LastFailure),
	}
}

// ConvertWorkspace converts an API Workspace object to a model
func ConvertStack(w *openapi.Stack) *models.Stack {
	return &models.Stack{
		Name:           w.Name,
		Labels:         convertLabels(w.Labels),
		Annotations:    convertAnnotations(w.Annotations),
		Connections:    convertConnections(w.Spec.Connections),
		StackResources: convertStackResources(w.Spec.StackResources),
		Volumes:        convertVolumes(w.Spec.Volumes),
	}
}

func presentConnections(connections models.StackConnections) []openapi.StackConnection {
	if connections == nil {
		return nil
	}
	result := make([]openapi.StackConnection, len(connections))
	for i, connection := range connections {
		result[i] = openapi.StackConnection{
			Kind: connection.Kind.String(),
			From: presentTopologyNodeRef(connection.From),
			To:   presentTopologyNodeRef(connection.To),
		}
		if connection.Id != "" {
			result[i].SetId(connection.Id)
		}
		if len(connection.Mappings) > 0 {
			result[i].SetMappings(presentConnectionMappings(connection.Mappings))
		}
		if len(connection.Config) > 0 {
			result[i].SetConfig(connection.Config)
		}
	}
	return result
}

func presentTopologyNodeRef(ref models.TopologyNodeRef) openapi.TopologyNodeRef {
	result := openapi.TopologyNodeRef{Type: string(ref.Type)}
	if ref.Id != "" {
		result.SetId(ref.Id)
	}
	if ref.Name != "" {
		result.SetName(ref.Name)
	}
	return result
}

func presentConnectionMappings(mappings []models.ConnectionMapping) []openapi.ConnectionMapping {
	result := make([]openapi.ConnectionMapping, len(mappings))
	for i, mapping := range mappings {
		result[i] = openapi.ConnectionMapping{
			Target: presentConnectionTarget(mapping.Target),
			Value:  presentValueRef(mapping.Value),
		}
	}
	return result
}

func presentConnectionTarget(target models.ConnectionTarget) openapi.ConnectionTarget {
	result := openapi.ConnectionTarget{Type: string(target.Type)}
	if target.Name != "" {
		result.SetName(target.Name)
	}
	if target.Path != "" {
		result.SetPath(target.Path)
	}
	return result
}

func presentValueRef(value models.ValueRef) openapi.ValueRef {
	result := openapi.ValueRef{}
	if value.Output != "" {
		result.SetOutput(value.Output)
	}
	if value.Template != "" {
		result.SetTemplate(value.Template)
	}
	if len(value.Values) > 0 {
		result.SetValues(presentOutputValueRefs(value.Values))
	}
	return result
}

func presentOutputValueRefs(values map[string]models.OutputValueRef) map[string]openapi.OutputValueRef {
	result := make(map[string]openapi.OutputValueRef, len(values))
	for key, value := range values {
		result[key] = openapi.OutputValueRef{Output: value.Output}
	}
	return result
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
		EnvFromAddons:      convertEnvFromAddons(config.EnvFromAddons),
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

func convertEnvFromAddons(sources []openapi.AddonEnvSource) []models.AddonEnvSource {
	if sources == nil {
		return nil
	}
	result := make([]models.AddonEnvSource, len(sources))
	for i, s := range sources {
		if s.Postgres != nil {
			result[i] = models.AddonEnvSource{
				Postgres: &models.PostgresAddonEnvSource{
					AddonID:    s.Postgres.AddonId,
					Database:   s.Postgres.GetDatabase(),
					Superuser:  s.Postgres.GetSuperuser(),
					EnvMapping: s.Postgres.EnvMapping,
				},
			}
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
			Name:   port.Name,
			Number: int(port.Number),
		}
		result[i].Protocol = port.GetProtocol()
		result[i].ExposedToPublic = port.GetExposedToPublic()
		result[i].SubdomainPrefix = port.GetSubdomainPrefix()
	}
	return models.Ports(result)
}

func convertConnections(connections []openapi.StackConnection) models.StackConnections {
	if connections == nil {
		return nil
	}
	result := make(models.StackConnections, len(connections))
	for i, connection := range connections {
		result[i] = models.StackConnection{
			Id:       connection.GetId(),
			Kind:     models.ConnectionKind(connection.Kind),
			From:     convertTopologyNodeRef(connection.From),
			To:       convertTopologyNodeRef(connection.To),
			Mappings: convertConnectionMappings(connection.Mappings),
			Config:   connection.GetConfig(),
		}
	}
	return result
}

func convertTopologyNodeRef(ref openapi.TopologyNodeRef) models.TopologyNodeRef {
	return models.TopologyNodeRef{
		Type: models.TopologyNodeType(ref.Type),
		Id:   ref.GetId(),
		Name: ref.GetName(),
	}
}

func convertConnectionMappings(mappings []openapi.ConnectionMapping) []models.ConnectionMapping {
	if mappings == nil {
		return nil
	}
	result := make([]models.ConnectionMapping, len(mappings))
	for i, mapping := range mappings {
		result[i] = models.ConnectionMapping{
			Target: convertConnectionTarget(mapping.Target),
			Value:  convertValueRef(mapping.Value),
		}
	}
	return result
}

func convertConnectionTarget(target openapi.ConnectionTarget) models.ConnectionTarget {
	return models.ConnectionTarget{
		Type: models.ConnectionTargetType(target.Type),
		Name: target.GetName(),
		Path: target.GetPath(),
	}
}

func convertValueRef(value openapi.ValueRef) models.ValueRef {
	result := models.ValueRef{
		Output:   value.GetOutput(),
		Template: value.GetTemplate(),
	}
	if values, ok := value.GetValuesOk(); ok && values != nil {
		result.Values = convertOutputValueRefs(*values)
	}
	return result
}

func convertOutputValueRefs(values map[string]openapi.OutputValueRef) map[string]models.OutputValueRef {
	result := make(map[string]models.OutputValueRef, len(values))
	for key, value := range values {
		result[key] = models.OutputValueRef{Output: value.Output}
	}
	return result
}
