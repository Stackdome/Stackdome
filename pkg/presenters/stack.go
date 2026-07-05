package presenters

import (
	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
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
		Settings:       presentStackSettings(s),
		Status:         presentStackStatus(s.Status),
		CreatedAt:      &s.CreatedAt,
		UpdatedAt:      &s.UpdatedAt,
	}
}

func presentStackSettings(s *models.Stack) *openapi.StackSettings {
	effective := s.EffectiveSettings()
	return &openapi.StackSettings{
		ReleaseRetentionLimit: ptr.To(int32(effective.ReleaseRetentionLimit)),
		MinSuccessfulReleases: ptr.To(int32(effective.MinSuccessfulReleases)),
		DeployTimeoutMinutes:  ptr.To(int32(effective.DeployTimeoutMinutes)),
	}
}

func PresentStackConnections(connections models.StackConnections) []openapi.StackConnection {
	return presentConnections(connections)
}

func PresentStackConnection(connection *models.StackConnection) openapi.StackConnection {
	if connection == nil {
		return openapi.StackConnection{}
	}
	return presentConnections(models.StackConnections{*connection})[0]
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
	if status.TargetRevision != "" {
		res.TargetRevision = &status.TargetRevision
	}
	if status.LastConverged != nil {
		res.LastConverged = &openapi.StackConvergenceRecord{
			Revision:  &status.LastConverged.Revision,
			ReleaseId: &status.LastConverged.ReleaseID,
			At:        &status.LastConverged.At,
		}
	}
	if len(status.Resources) > 0 {
		summaries := make([]openapi.StackResourceSummary, len(status.Resources))
		for i, r := range status.Resources {
			summaries[i] = openapi.StackResourceSummary{
				Name:              &r.Name,
				Phase:             ptr.To(string(r.Phase)),
				ObservedRevision:  &r.ObservedRevision,
				ConvergedRevision: &r.ConvergedRevision,
				AvailableReplicas: &r.AvailableReplicas,
				UpdatedReplicas:   &r.UpdatedReplicas,
				Replicas:          &r.Replicas,
				Missing:           &r.Missing,
				Message:           &r.Message,
			}
		}
		res.Resources = summaries
	}
	if status.LastValidationRun != nil && !status.LastValidationRun.Passed {
		res.Message = &status.LastValidationRun.Message
	}
	return res
}

func presentStackResources(resources []*models.StackResource) []openapi.StackResource {
	result := make([]openapi.StackResource, len(resources))
	for i, r := range resources {
		result[i] = PresentStackResource(r)
	}
	return result
}

// presentSource reconstructs the intent-level source union from the internal
// build/image specs. Credential material is never echoed.
func presentSource(r *models.StackResource) *openapi.SourceSpec {
	switch {
	case r.ImageConfig != nil:
		img := openapi.NewImageSource(r.ImageConfig.Image)
		if r.ImageConfig.RegistryCredentialID != "" {
			img.SetRegistryCredentialsId(r.ImageConfig.RegistryCredentialID)
		}
		return &openapi.SourceSpec{Image: img}
	case r.BuildConfig != nil:
		bc := r.BuildConfig
		switch {
		case bc.SourceContext.Git != nil:
			git := openapi.NewGitSource(bc.SourceContext.Git.RepoURL)
			if rev := bc.SourceRevision.Git; rev != nil {
				if rev.Branch != "" {
					git.SetBranch(rev.Branch)
				}
				if rev.Tag != "" {
					git.SetTag(rev.Tag)
				}
				if rev.Commit != "" {
					git.SetCommit(rev.Commit)
				}
			}
			if bc.DockerfilePath != "" {
				git.SetDockerfilePath(bc.DockerfilePath)
			}
			if bc.ContextPathWithinSource != "" {
				git.SetBuildContext(bc.ContextPathWithinSource)
			}
			if bc.SourceContext.Git.IntegrationID != "" {
				git.SetIntegrationId(bc.SourceContext.Git.IntegrationID)
			}
			if !bc.BuildImageRepository.UseInClusterRegistry && bc.BuildImageRepository.ExternalImageRef != "" {
				push := openapi.NewPushTarget(bc.BuildImageRepository.ExternalImageRef)
				if bc.PushRegistryCredentialID != "" {
					push.SetRegistryCredentialsId(bc.PushRegistryCredentialID)
				}
				git.SetPush(*push)
			}
			return &openapi.SourceSpec{Git: git}
		case bc.SourceContext.Volume != nil:
			vol := openapi.NewVolumeBuildSource()
			if bc.SourceContext.Volume.SourceVolumeID != "" {
				vol.SetVolumeId(bc.SourceContext.Volume.SourceVolumeID)
			}
			if bc.SourceContext.Volume.SourceVolumeName != "" {
				vol.SetVolumeName(bc.SourceContext.Volume.SourceVolumeName)
			}
			if bc.SourceRevision.Volume != nil && bc.SourceRevision.Volume.CurrentVolumeHash != "" {
				vol.SetCurrentVolumeHash(bc.SourceRevision.Volume.CurrentVolumeHash)
			}
			if bc.DockerfilePath != "" {
				vol.SetDockerfilePath(bc.DockerfilePath)
			}
			if bc.ContextPathWithinSource != "" {
				vol.SetBuildContext(bc.ContextPathWithinSource)
			}
			return &openapi.SourceSpec{Volume: vol}
		}
	}
	return nil
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
		case revision.Git.Branch != "":
			res.GitRepoRevision.Branch = &revision.Git.Branch
			if revision.Git.Commit != "" {
				res.GitRepoRevision.Commit = &revision.Git.Commit
			}
		case revision.Git.Tag != "":
			res.GitRepoRevision.Tag = &revision.Git.Tag
			if revision.Git.Commit != "" {
				res.GitRepoRevision.Commit = &revision.Git.Commit
			}
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
			RepoUrl: context.Git.RepoURL,
		}
	}
	return res
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
		result[i] = openapi.EnvVar{Name: env.Name}
		if env.Value != "" {
			result[i].SetValue(env.Value)
		}
		if env.SelfOutput != "" {
			result[i].SetSelfOutput(env.SelfOutput)
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
		State:             ptr.To(string(status.State)),
		ObservedRevision:  &status.ObservedCrRevision,
		Conditions:        presentConditions(status.Conditions),
		LastFailure:       presentStackResourceFailure(status.LastFailure),
		Replicas:          &status.Replicas,
		AvailableReplicas: &status.AvailableReplicas,
		UpdatedReplicas:   &status.UpdatedReplicas,
		LastRunTime:       status.LastRunTime,
		LastRunSucceeded:  status.LastRunSucceeded,
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
		Settings:       convertStackSettings(w.Settings),
	}
}

func convertStackSettings(s *openapi.StackSettings) *models.StackSettings {
	if s == nil {
		return nil
	}
	result := &models.StackSettings{}
	if s.ReleaseRetentionLimit != nil {
		result.ReleaseRetentionLimit = int(*s.ReleaseRetentionLimit)
	}
	if s.MinSuccessfulReleases != nil {
		result.MinSuccessfulReleases = int(*s.MinSuccessfulReleases)
	}
	if s.DeployTimeoutMinutes != nil {
		result.DeployTimeoutMinutes = int(*s.DeployTimeoutMinutes)
	}
	return result
}

func ConvertStackConnection(connection *openapi.StackConnection) *models.StackConnection {
	if connection == nil {
		return nil
	}
	converted := convertConnections([]openapi.StackConnection{*connection})
	if len(converted) == 0 {
		return nil
	}
	return &converted[0]
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
		if connection.ID != "" {
			result[i].SetId(connection.ID)
		}
		if len(connection.Mappings) > 0 {
			result[i].SetMappings(presentConnectionMappings(connection.Mappings))
		}
		if apiConfig := presentConnectionConfig(connection.Kind, connection.From, connection.Config); apiConfig != nil {
			result[i].SetConfig(*apiConfig)
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

func ConvertStackResource(r *openapi.StackResource) *models.StackResource {
	return convertStackResource(r)
}

func convertStackResource(r *openapi.StackResource) *models.StackResource {
	buildConfig, imageConfig := convertSource(r.Source)
	return &models.StackResource{
		Name:            r.Name,
		Labels:          convertLabels(r.Labels),
		Annotations:     convertAnnotations(r.Annotations),
		BuildConfig:     buildConfig,
		ImageConfig:     imageConfig,
		Init:            convertInitConfig(r.InitSpec),
		ExecutionConfig: convertExecutionConfig(r.ExecutionConfig),
		VolumeMounts:    convertVolumeMounts(r.VolumeMounts),
		DependsOn:       convertDependencies(r.DependsOn),
		LifecycleConfig: convertLifecycleConfig(r.LifecycleConfig),
		Ports:           convertPorts(r.Ports),
		WorkloadType:    models.WorkloadType(r.GetWorkloadType()),
		Schedule:        r.GetSchedule(),
		Replicas:        r.Replicas,
	}
}

// convertSource maps the intent-level source union onto the internal
// build/image specs.
func convertSource(src *openapi.SourceSpec) (*models.BuildConfigSpec, *models.ImageConfigSpec) {
	if src == nil {
		return nil, nil
	}
	switch {
	case src.Image != nil:
		return nil, &models.ImageConfigSpec{
			Image:                src.Image.Ref,
			RegistryCredentialID: src.Image.GetRegistryCredentialsId(),
		}
	case src.Git != nil:
		git := src.Git
		res := &models.BuildConfigSpec{
			SourceContext: models.BuildContextSource{
				Git: &models.GitBuildSource{
					RepoURL:       git.RepoUrl,
					IntegrationID: git.GetIntegrationId(),
				},
			},
			SourceRevision: models.BuildSourceRevision{
				Git: &models.GitRevision{
					Branch: git.GetBranch(),
					Tag:    git.GetTag(),
					Commit: git.GetCommit(),
				},
			},
			DockerfilePath:          buildPathOrDefault(git.GetDockerfilePath(), models.DefaultDockerfilePath),
			ContextPathWithinSource: buildPathOrDefault(git.GetBuildContext(), models.DefaultBuildContext),
		}
		if git.Push == nil {
			res.BuildImageRepository = models.BuildImageRepository{
				UseInClusterRegistry: true,
				InsecureRegistry:     true,
			}
		} else {
			res.BuildImageRepository.ExternalImageRef = git.Push.Repository
			res.PushRegistryCredentialID = git.Push.GetRegistryCredentialsId()
		}
		return res, nil
	case src.Volume != nil:
		vol := src.Volume
		res := &models.BuildConfigSpec{
			SourceContext: models.BuildContextSource{
				Volume: &models.VolumeBuildSource{
					SourceVolumeID:   vol.GetVolumeId(),
					SourceVolumeName: vol.GetVolumeName(),
				},
			},
			SourceRevision: models.BuildSourceRevision{
				Volume: &models.VolumeRevision{
					CurrentVolumeHash: vol.GetCurrentVolumeHash(),
				},
			},
			DockerfilePath:          buildPathOrDefault(vol.GetDockerfilePath(), models.DefaultDockerfilePath),
			ContextPathWithinSource: buildPathOrDefault(vol.GetBuildContext(), models.DefaultBuildContext),
			BuildImageRepository: models.BuildImageRepository{
				UseInClusterRegistry: true,
				InsecureRegistry:     true,
			},
		}
		return res, nil
	}
	return nil, nil
}

func buildPathOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
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
		Command: config.Command,
		Args:    config.Args,
		Env:     convertEnvVars(config.EnvironmentVariables),
	}
}

func convertEnvVars(envVars []openapi.EnvVar) []models.EnvVar {
	result := make([]models.EnvVar, len(envVars))
	for i, env := range envVars {
		result[i] = models.EnvVar{
			Name:       env.Name,
			Value:      env.GetValue(),
			SelfOutput: env.GetSelfOutput(),
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
			ID:       connection.GetId(),
			Kind:     models.ConnectionKind(connection.Kind),
			From:     convertTopologyNodeRef(connection.From),
			To:       convertTopologyNodeRef(connection.To),
			Mappings: convertConnectionMappings(connection.Mappings),
			Config:   convertConnectionConfig(connection.GetConfig()),
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

func presentConnectionConfig(kind models.ConnectionKind, from models.TopologyNodeRef, cfg models.ConnectionConfig) *openapi.StackConnectionConfig {
	if len(cfg) == 0 {
		return nil
	}
	switch kind {
	case models.ConnectionKindVolumeMount:
		vc := openapi.NewVolumeMountConfig(stringFromConfig(cfg, string(models.ConnectionConfigKeyMountPath)))
		if v, ok := cfg[string(models.ConnectionConfigKeySubPath)].(string); ok {
			vc.SubPath = &v
		}
		if v, ok := cfg[string(models.ConnectionConfigKeyReadOnly)].(bool); ok {
			vc.ReadOnly = &v
		}
		return &openapi.StackConnectionConfig{VolumeMountConfig: vc}

	case models.ConnectionKindBuildArtifactSource:
		bc := openapi.NewBuildArtifactSourceConfig(stringFromConfig(cfg, string(models.ConnectionConfigKeySourcePath)))
		if v, ok := cfg[string(models.ConnectionConfigKeyDestinationPath)].(string); ok {
			bc.DestinationPath = &v
		}
		return &openapi.StackConnectionConfig{BuildArtifactSourceConfig: bc}

	case models.ConnectionKindEnv:
		if from.Type != models.TopologyNodeTypePostgresAddon {
			return nil
		}
		pc := openapi.NewPostgresEnvConfig()
		if v, ok := cfg[string(models.ConnectionConfigKeyDatabase)].(string); ok {
			pc.Database = &v
		}
		if v, ok := cfg[string(models.ConnectionConfigKeyCredentialScope)].(string); ok {
			pc.CredentialScope = &v
		}
		if v, ok := cfg[string(models.ConnectionConfigKeySuperuser)].(bool); ok {
			pc.Superuser = &v
		}
		if pc.Database == nil && pc.CredentialScope == nil && pc.Superuser == nil {
			return nil
		}
		return &openapi.StackConnectionConfig{PostgresEnvConfig: pc}
	}
	return nil
}

func convertConnectionConfig(cfg openapi.StackConnectionConfig) models.ConnectionConfig {
	result := make(models.ConnectionConfig)
	switch {
	case cfg.PostgresEnvConfig != nil:
		pc := cfg.PostgresEnvConfig
		if pc.Database != nil {
			result[string(models.ConnectionConfigKeyDatabase)] = *pc.Database
		}
		if pc.CredentialScope != nil {
			result[string(models.ConnectionConfigKeyCredentialScope)] = *pc.CredentialScope
		}
		if pc.Superuser != nil {
			result[string(models.ConnectionConfigKeySuperuser)] = *pc.Superuser
		}
	case cfg.VolumeMountConfig != nil:
		vc := cfg.VolumeMountConfig
		result[string(models.ConnectionConfigKeyMountPath)] = vc.MountPath
		if vc.SubPath != nil {
			result[string(models.ConnectionConfigKeySubPath)] = *vc.SubPath
		}
		if vc.ReadOnly != nil {
			result[string(models.ConnectionConfigKeyReadOnly)] = *vc.ReadOnly
		}
	case cfg.BuildArtifactSourceConfig != nil:
		bc := cfg.BuildArtifactSourceConfig
		result[string(models.ConnectionConfigKeySourcePath)] = bc.SourcePath
		if bc.DestinationPath != nil {
			result[string(models.ConnectionConfigKeyDestinationPath)] = *bc.DestinationPath
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func stringFromConfig(cfg models.ConnectionConfig, key string) string {
	if v, ok := cfg[key].(string); ok {
		return v
	}
	return ""
}
