package presenters

import (
	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
)

func PresentVolumeList(volumes []*models.Volume, withStatus bool) []openapi.Volume {
	result := make([]openapi.Volume, len(volumes))
	for i, volume := range volumes {
		result[i] = PresentVolume(volume, withStatus)
	}
	return result
}

func PresentVolume(v *models.Volume, withStatus bool) openapi.Volume {
	if v == nil {
		return openapi.Volume{}
	}
	res := openapi.Volume{
		Id:          &v.ID,
		ProjectId:   &v.ProjectID,
		Name:        v.Name,
		Spec:        presentVolumeSpec(v),
		Labels:      presentLabels(v.Labels),
		Annotations: presentAnnotations(v.Annotations),
	}
	if withStatus {
		res.Status = presentVolumeStatus(v.Status)
	}
	return res
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

func presentVolumes(volumes []*models.Volume, withStatus bool) []openapi.Volume {
	if len(volumes) == 0 {
		return nil
	}
	result := make([]openapi.Volume, len(volumes))
	for i, volume := range volumes {
		result[i] = PresentVolume(volume, withStatus)
	}
	return result
}

func presentVolumeSpec(spec *models.Volume) openapi.VolumeSpec {
	if spec == nil {
		return openapi.VolumeSpec{}
	}
	res := openapi.VolumeSpec{
		Size:               spec.Size,
		StorageClass:       &spec.StorageClass,
		NeedsSyncBeforeUse: spec.SyncBeforeUse,
		AccessMode:         presentVolumeAccessMode(spec.AccessMode),
	}
	if spec.VolumeSource != nil && spec.VolumeSource.GitRepoSource != nil {
		res.Source = &openapi.VolumeSource{GitRepoSource: *presentGitRepoSource(spec.VolumeSource.GitRepoSource)}
	}

	return res
}

func presentGitRepoSource(source *models.GitRepoSource) *openapi.GitRepoSource {
	if source == nil {
		return nil
	}
	revision := openapi.GitRepoRevision{}
	switch source.Revision.Type() {
	case models.Commit:
		revision.Commit = &source.Revision.Commit
	case models.Tag:
		revision.Tag = &source.Revision.Tag
	case models.Branch:
		revision.Branch = &source.Revision.Branch
		if source.Revision.Commit != "" {
			revision.Commit = &source.Revision.Commit
		}
	}
	return &openapi.GitRepoSource{
		RepoUrl:  source.RepoUrl,
		Revision: revision,
	}
}

func presentVolumeAccessMode(mode models.VolumeAccessMode) openapi.VolumeAccessMode {
	switch mode {
	case models.READ_WRITE_ONCE:
		return openapi.READ_WRITE_ONCE
	case models.READ_ONLY_MANY:
		return openapi.READ_ONLY_MANY
	case models.READ_WRITE_MANY:
		return openapi.READ_WRITE_MANY
	default:
		return openapi.READ_WRITE_ONCE
	}
}

func presentVolumeStatus(status *models.VolumeStatus) *openapi.VolumeStatus {
	if status == nil {
		return nil
	}
	res := &openapi.VolumeStatus{
		Conditions: presentConditions(status.Conditions),
		Phase:      &status.Phase,
	}
	if status.LastSyncedGitRevision != "" {
		res.LastSyncedGitRevision = &status.LastSyncedGitRevision
	}

	return res
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

func ConvertVolume(v *openapi.Volume) *models.Volume {
	res := &models.Volume{
		Name:          v.Name,
		Labels:        convertLabels(v.Labels),
		Annotations:   convertAnnotations(v.Annotations),
		AccessMode:    convertVolumeAccessMode(v.Spec.AccessMode),
		Size:          v.Spec.Size,
		StorageClass:  v.Spec.GetStorageClass(),
		SyncBeforeUse: v.Spec.GetNeedsSyncBeforeUse(),
	}
	if v.Spec.Source != nil {
		res.VolumeSource = &models.VolumeSource{GitRepoSource: convertGitRepoSource(&v.Spec.Source.GitRepoSource)}
	}
	return res
}

func convertVolumeAccessMode(mode openapi.VolumeAccessMode) models.VolumeAccessMode {
	switch mode {
	case openapi.READ_WRITE_ONCE:
		return models.READ_WRITE_ONCE
	case openapi.READ_ONLY_MANY:
		return models.READ_ONLY_MANY
	case openapi.READ_WRITE_MANY:
		return models.READ_WRITE_MANY
	default:
		return models.READ_WRITE_ONCE
	}
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

func convertVolumes(volumes []openapi.Volume) []*models.Volume {
	result := make([]*models.Volume, len(volumes))
	for i, volume := range volumes {
		result[i] = ConvertVolume(&volume)
	}
	return result
}

func convertGitRepoSource(source *openapi.GitRepoSource) *models.GitRepoSource {
	if source == nil {
		return nil
	}
	revision := models.GitRepoRevision{}
	switch {
	case source.Revision.Commit != nil:
		revision.Commit = source.Revision.GetCommit()
	case source.Revision.Tag != nil:
		revision.Tag = source.Revision.GetTag()
	case source.Revision.Branch != nil:
		revision.Branch = source.Revision.GetBranch()
		if source.Revision.Commit != nil {
			revision.Commit = source.Revision.GetCommit()
		}
	}
	return &models.GitRepoSource{
		RepoUrl:  source.RepoUrl,
		Revision: revision,
	}
}
