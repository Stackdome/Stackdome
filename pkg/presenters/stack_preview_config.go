package presenters

import (
	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"k8s.io/utils/ptr"
)

func PresentStackPreviewConfig(c *models.StackPreviewConfig) openapi.StackPreviewConfig {
	result := openapi.StackPreviewConfig{
		Id:                &c.ID,
		OrganisationId:    &c.OrganisationID,
		TeamId:            &c.TeamID,
		UserId:            &c.UserID,
		Name:              &c.Name,
		Description:       &c.Description,
		StackfilePath:     &c.StackfilePath,
		MaxActivePreviews: ptr.To(int32(c.MaxActivePreviews)),
		Labels:            presentLabels(c.Labels),
		Annotations:       presentAnnotations(c.Annotations),
		CreatedAt:         &c.CreatedAt,
		UpdatedAt:         &c.UpdatedAt,
		GitRepository:     presentPreviewGitRepository(c.GitRepository),
	}
	return result
}

// presentPreviewGitRepository never echoes credential material.
func presentPreviewGitRepository(repo models.PreviewGitRepository) *openapi.PreviewGitRepository {
	res := &openapi.PreviewGitRepository{
		RepoUrl:    repo.RepoURL,
		BaseBranch: &repo.BaseBranch,
	}
	if repo.IntegrationID != "" {
		res.SetIntegrationId(repo.IntegrationID)
	}
	return res
}

func convertPreviewGitRepository(in *openapi.PreviewGitRepository) models.PreviewGitRepository {
	if in == nil {
		return models.PreviewGitRepository{}
	}
	return models.PreviewGitRepository{
		RepoURL:       in.RepoUrl,
		BaseBranch:    in.GetBaseBranch(),
		IntegrationID: in.GetIntegrationId(),
	}
}

func ConvertStackPreviewConfigCreate(req *openapi.StackPreviewConfigCreate) *models.StackPreviewConfig {
	config := &models.StackPreviewConfig{
		Name:              req.Name,
		Description:       req.GetDescription(),
		StackfilePath:     req.GetStackfilePath(),
		MaxActivePreviews: int(req.GetMaxActivePreviews()),
		GitRepository:     convertPreviewGitRepository(&req.GitRepository),
		Labels:            convertLabels(req.Labels),
		Annotations:       convertAnnotations(req.Annotations),
	}
	return config
}

func ConvertStackPreviewConfigUpdate(req *openapi.StackPreviewConfigUpdate) *models.StackPreviewConfig {
	config := &models.StackPreviewConfig{
		Description:       req.GetDescription(),
		StackfilePath:     req.GetStackfilePath(),
		MaxActivePreviews: int(req.GetMaxActivePreviews()),
	}
	if req.GitRepository != nil {
		config.GitRepository = convertPreviewGitRepository(req.GitRepository)
	}
	if req.Labels != nil {
		config.Labels = convertLabels(req.Labels)
	}
	if req.Annotations != nil {
		config.Annotations = convertAnnotations(req.Annotations)
	}
	return config
}

func PresentStackPreviewConfigList(result *stores.PaginatedResult[*models.StackPreviewConfig]) openapi.StackPreviewConfigList {
	items := make([]openapi.StackPreviewConfig, len(result.Items))
	for i, c := range result.Items {
		items[i] = PresentStackPreviewConfig(c)
	}
	return openapi.StackPreviewConfigList{
		Items:      items,
		Total:      ptr.To(int32(result.Total)),
		Page:       ptr.To(int32(result.Page)),
		PageSize:   ptr.To(int32(result.PageSize)),
		TotalPages: ptr.To(int32(result.TotalPages)),
	}
}
