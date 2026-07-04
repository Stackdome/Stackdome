package presenters

import (
	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"k8s.io/utils/ptr"
)

// presentPreviewStatus converts a model PreviewStackStatus into the generated
// API response status object.
func presentPreviewStatus(s models.PreviewStackStatus) *openapi.PreviewStackStatus {
	phase := string(s.Phase)
	status := &openapi.PreviewStackStatus{
		Phase: &phase,
	}
	if s.Reason != "" {
		status.Reason = &s.Reason
	}
	if s.Message != "" {
		status.Message = &s.Message
	}
	if s.Outputs != nil {
		outputs := openapi.PreviewStackStatusOutputs{}
		if s.Outputs.CommitSHA != "" {
			outputs.CommitSha = &s.Outputs.CommitSHA
		}
		if len(s.Outputs.URLs) > 0 {
			urls := make([]openapi.PreviewStackStatusOutputsUrlsInner, len(s.Outputs.URLs))
			for i, u := range s.Outputs.URLs {
				resource := u.Resource
				url := u.URL
				urls[i] = openapi.PreviewStackStatusOutputsUrlsInner{
					Resource: &resource,
					Url:      &url,
				}
			}
			outputs.Urls = urls
		}
		status.Outputs = &outputs
	}
	return status
}

// ConvertPreviewStackCreate converts an API create request into a PreviewStack model.
func ConvertPreviewStackCreate(req *openapi.PreviewStackCreate) *models.PreviewStack {
	preview := &models.PreviewStack{
		StackPreviewConfigID: req.ConfigId,
		PRNumber:             req.GetPrNumber(),
		Branch:               req.Branch,
		CommitSHA:            req.GetCommit(),
		Source:               models.PreviewStackSourceManual,
		StackfileContent:     req.StackfileContent,
		ImageOverrides:       models.ImageOverrides(req.GetImageOverrides()),
	}
	return preview
}

// PresentPreviewStack converts a PreviewStack model into an API response.
func PresentPreviewStack(p *models.PreviewStack) openapi.PreviewStack {
	source := string(p.Source)
	overrides := map[string]string(p.ImageOverrides)

	result := openapi.PreviewStack{
		Id:                &p.ID,
		OrganisationId:    &p.OrganisationID,
		TeamId:            &p.TeamID,
		UserId:            &p.UserID,
		ConfigId:          &p.StackPreviewConfigID,
		StackId:           p.StackID,
		Name:              &p.Name,
		PrNumber:          &p.PRNumber,
		Branch:            &p.Branch,
		Commit:            &p.CommitSHA,
		Source:            &source,
		Status:            presentPreviewStatus(p.Status),
		Labels:            presentLabels(p.Labels),
		Annotations:       presentAnnotations(p.Annotations),
		DeletionTimestamp: p.DeletionTimestamp,
		CreatedAt:         &p.CreatedAt,
		UpdatedAt:         &p.UpdatedAt,
	}

	if len(overrides) > 0 {
		result.ImageOverrides = &overrides
	}

	return result
}

// PresentPreviewStackList converts a paginated result of PreviewStack models
// into an API list response.
func PresentPreviewStackList(result *stores.PaginatedResult[*models.PreviewStack]) openapi.PreviewStackList {
	items := make([]openapi.PreviewStack, len(result.Items))
	for i, p := range result.Items {
		items[i] = PresentPreviewStack(p)
	}
	return openapi.PreviewStackList{
		Items:      items,
		Total:      ptr.To(int32(result.Total)),
		Page:       ptr.To(int32(result.Page)),
		PageSize:   ptr.To(int32(result.PageSize)),
		TotalPages: ptr.To(int32(result.TotalPages)),
	}
}
