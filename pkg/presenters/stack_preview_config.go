package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
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
		GitRepository: &openapi.PreviewGitRepository{
			RepoUrl:      c.GitRepository.RepoURL,
			BaseBranch:   &c.GitRepository.BaseBranch,
			GitSecretRef: c.GitRepository.GitSecretID,
		},
	}
	return result
}

func ConvertStackPreviewConfigCreate(req *openapi.StackPreviewConfigCreate) *models.StackPreviewConfig {
	config := &models.StackPreviewConfig{
		Name:              req.Name,
		Description:       req.GetDescription(),
		StackfilePath:     req.GetStackfilePath(),
		MaxActivePreviews: int(req.GetMaxActivePreviews()),
		GitRepository: models.PreviewGitRepository{
			RepoURL:    req.GitRepository.RepoUrl,
			BaseBranch: req.GitRepository.GetBaseBranch(),
		},
		Labels:      convertLabels(req.Labels),
		Annotations: convertAnnotations(req.Annotations),
	}
	if req.GitRepository.GitSecretRef != nil {
		config.GitRepository.GitSecretID = req.GitRepository.GitSecretRef
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
		config.GitRepository = models.PreviewGitRepository{
			RepoURL:    req.GitRepository.RepoUrl,
			BaseBranch: req.GitRepository.GetBaseBranch(),
		}
		if req.GitRepository.GitSecretRef != nil {
			config.GitRepository.GitSecretID = req.GitRepository.GitSecretRef
		}
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
