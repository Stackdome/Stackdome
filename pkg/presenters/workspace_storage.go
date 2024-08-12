package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"k8s.io/utils/ptr"
)

func PresentWorkspaceStorageList(wss []*models.WorkspaceStorage) []openapi.WorkspaceStorage {
	result := make([]openapi.WorkspaceStorage, len(wss))
	for i, ws := range wss {
		result[i] = PresentWorkspaceStorage(ws)
	}
	return result
}

func PresentWorkspaceStorage(ws *models.WorkspaceStorage) openapi.WorkspaceStorage {
	return openapi.WorkspaceStorage{
		Id:             &ws.ID,
		OrganisationId: &ws.OrganisationID,
		Name:           ws.Name,
		Namespace:      ws.Namespace,
		Labels:         presentLabels(ws.Labels),
		Annotations:    presentAnnotations(ws.Annotations),
		SshConfig:      presentSSHConfig(ws.SSHConfig),
		Volumes:        presentVolumes(ws.Volumes),
		Status:         presentWorkspaceStorageStatus(ws.Status),
		State:          ptr.To(presentWorkpaceStorageState(ws.State)),
		CreatedAt:      &ws.CreatedAt,
		UpdatedAt:      &ws.UpdatedAt,
	}
}

func presentWorkpaceStorageState(state models.WorkspaceStorageState) openapi.WorkspaceStorageState {
	switch state {
	case models.WorkspaceStorageStateCreated:
		return openapi.STORAGE_CREATED
	case models.WorkspaceStorageStateFailed:
		return openapi.STORAGE_FAILED
	case models.WorkspaceStorageStateReady:
		return openapi.STORAGE_READY
	case models.WorkspaceStorageStateCreating:
		return openapi.STORAGE_CREATING
	default:
		return openapi.STORAGE_PENDING
	}
}

func PresentVolume(v *models.Volume) openapi.Volume {
	if v == nil {
		return openapi.Volume{}
	}
	return openapi.Volume{
		Id:          &v.ID,
		Name:        v.Name,
		Spec:        presentWorkspaceVolumeSpec(v),
		Labels:      presentLabels(v.Labels),
		Annotations: presentAnnotations(v.Annotations),
		Status:      presentVolumeStatus(v.VolumeStatus),
	}
}

func presentLabels(labels models.Labels) []openapi.Label {
	if len(labels) == 0 {
		return nil
	}
	result := make([]openapi.Label, len(labels))
	for i, label := range labels {
		result[i] = openapi.Label{
			Key:   label.Key,
			Value: label.Value,
		}
	}
	return result
}

func presentAnnotations(annotations models.Annotations) []openapi.Annotation {
	if len(annotations) == 0 {
		return nil
	}
	result := make([]openapi.Annotation, len(annotations))
	for i, annotation := range annotations {
		result[i] = openapi.Annotation{
			Key:   annotation.Key,
			Value: annotation.Value,
		}
	}
	return result
}

func presentSSHConfig(config *models.SSHConfig) openapi.SSHConfig {
	if config == nil {
		return openapi.SSHConfig{}
	}
	return openapi.SSHConfig{
		PublicKey: config.PublicKey,
	}
}

func presentVolumes(volumes []models.Volume) []openapi.Volume {
	if len(volumes) == 0 {
		return nil
	}
	result := make([]openapi.Volume, len(volumes))
	for i, volume := range volumes {
		result[i] = PresentVolume(&volume)
	}
	return result
}

func presentWorkspaceStorageStatus(status *models.WorkspaceStorageStatus) *openapi.WorkspaceStorageStatus {
	if status == nil {
		return nil
	}
	return &openapi.WorkspaceStorageStatus{
		Conditions:               presentConditions(status.Conditions),
		Phase:                    &status.Phase,
		StorageServerServiceName: &status.StorageServerServiceName,
	}
}

func presentWorkspaceVolumeSpec(spec *models.Volume) openapi.WorkspaceVolumeSpec {
	if spec == nil {
		return openapi.WorkspaceVolumeSpec{}
	}
	res := openapi.WorkspaceVolumeSpec{
		Size:          spec.Size,
		StorageClass:  &spec.StorageClass,
		SyncBeforeUse: &spec.SyncBeforeUse,
		Source:        &openapi.VolumeSource{},
	}
	if spec.BuildSource != nil && len(spec.BuildSource) > 0 {
		res.Source.SourceType = openapi.BUILD_ARTIFACT
		res.Source.BuildSource = presentBuildArtifacts(spec.BuildSource)
	}
	if spec.LocalSource != nil {
		res.Source.SourceType = openapi.LOCAL
		res.Source.LocalSource = presentLocalSource(spec.LocalSource)
	}

	return res
}

func presentLocalSource(source *models.LocalSource) *openapi.LocalSource {
	if source == nil {
		return nil
	}
	return &openapi.LocalSource{
		Path: source.Path,
		Sync: source.Sync,
	}
}

func presentBuildArtifacts(artifacts []models.BuildArtifactSource) []openapi.BuildArtifact {
	if len(artifacts) == 0 {
		return nil
	}
	result := make([]openapi.BuildArtifact, len(artifacts))
	for i, artifact := range artifacts {
		result[i] = openapi.BuildArtifact{
			ResourceRef:     artifact.ResourceName,
			SourcePath:      artifact.SourcePath,
			DestinationPath: artifact.DestinationPath,
		}
	}
	return result
}

func presentVolumeStatus(status *models.VolumeStatus) *openapi.VolumeStatus {
	if status == nil {
		return nil
	}
	return &openapi.VolumeStatus{
		Conditions:         presentConditions(status.Conditions),
		Phase:              &status.Phase,
		BuildArtifactSyncs: presentBuildArtifactSyncInfo(status.BuildArtifactSyncs),
	}
}

func presentConditions(conditions []models.Condition) []openapi.Condition {
	if len(conditions) == 0 {
		return nil
	}
	result := make([]openapi.Condition, len(conditions))
	for i, condition := range conditions {
		result[i] = openapi.Condition{
			Type:               &condition.Type,
			Status:             &condition.Status,
			LastTransitionTime: &condition.LastTransitionTime,
			Reason:             &condition.Reason,
			Message:            &condition.Message,
		}
	}
	return result
}

func presentBuildArtifactSyncInfo(info []models.BuildArtifactSyncInfo) []openapi.BuildArtifactSyncInfo {
	if len(info) == 0 {
		return nil
	}
	result := make([]openapi.BuildArtifactSyncInfo, len(info))

	for i, syncInfo := range info {
		result[i] = openapi.BuildArtifactSyncInfo{
			ResourceName: &syncInfo.ResourceName,
			BuildId:      &syncInfo.BuildID,
			Status:       &syncInfo.Status,
		}
	}
	return result
}

// Converters

func ConvertWorkspaceStorage(ws *openapi.WorkspaceStorage) *models.WorkspaceStorage {
	return &models.WorkspaceStorage{
		Name:        ws.Name,
		Namespace:   ws.Namespace,
		Labels:      convertLabels(ws.Labels),
		Annotations: convertAnnotations(ws.Annotations),
		SSHConfig:   convertSSHConfig(ws.SshConfig),
		Volumes:     convertVolumes(ws.Volumes),
		CreatedAt:   ws.GetCreatedAt(),
		UpdatedAt:   ws.GetUpdatedAt(),
	}
}

func ConvertVolume(v *openapi.Volume) models.Volume {
	res := models.Volume{
		ID:            v.Name,
		Name:          v.Name,
		Labels:        convertLabels(v.Labels),
		Annotations:   convertAnnotations(v.Annotations),
		Size:          v.Spec.Size,
		StorageClass:  v.Spec.GetStorageClass(),
		SyncBeforeUse: v.Spec.GetSyncBeforeUse(),
	}
	if v.Spec.Source != nil {
		switch v.Spec.Source.SourceType {
		case openapi.LOCAL:
			res.LocalSource = convertLocalSource(v.Spec.Source.LocalSource)
		case openapi.BUILD_ARTIFACT:
			res.BuildSource = convertBuildArtifacts(v.Spec.Source.BuildSource)
		}
	}
	return res
}

func convertLabels(labels []openapi.Label) models.Labels {
	result := make(models.Labels, len(labels))
	for i, label := range labels {
		result[i] = models.Label{
			Key:   label.Key,
			Value: label.Value,
		}
	}
	return result
}

func convertAnnotations(annotations []openapi.Annotation) models.Annotations {
	result := make(models.Annotations, len(annotations))
	for i, annotation := range annotations {
		result[i] = models.Annotation{
			Key:   annotation.Key,
			Value: annotation.Value,
		}
	}
	return result
}

func convertSSHConfig(config openapi.SSHConfig) *models.SSHConfig {
	return &models.SSHConfig{
		PublicKey: config.PublicKey,
	}
}

func convertVolumes(volumes []openapi.Volume) []models.Volume {
	result := make([]models.Volume, len(volumes))
	for i, volume := range volumes {
		result[i] = ConvertVolume(&volume)
	}
	return result
}

func convertLocalSource(source *openapi.LocalSource) *models.LocalSource {
	if source == nil {
		return nil
	}
	return &models.LocalSource{
		Path: source.Path,
		Sync: source.Sync,
	}
}

func convertBuildArtifacts(artifacts []openapi.BuildArtifact) []models.BuildArtifactSource {
	result := make([]models.BuildArtifactSource, len(artifacts))
	for i, artifact := range artifacts {
		result[i] = models.BuildArtifactSource{
			ResourceName:    artifact.ResourceRef,
			SourcePath:      artifact.SourcePath,
			DestinationPath: artifact.DestinationPath,
		}
	}
	return result
}
