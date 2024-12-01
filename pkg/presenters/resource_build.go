package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func PresentWorkspaceResourceBuildList(in []*models.WorkspaceResourceBuild) []openapi.WorkspaceResourceBuild {
	result := make([]openapi.WorkspaceResourceBuild, len(in))
	for i, w := range in {
		result[i] = PresentWorkspaceResourceBuild(w)
	}
	return result
}

func PresentWorkspaceResourceBuild(in *models.WorkspaceResourceBuild) openapi.WorkspaceResourceBuild {
	return openapi.WorkspaceResourceBuild{
		Id:                  &in.ID,
		Namespace:           &in.Namespace,
		WorkspaceId:         &in.WorkspaceID,
		WorkspaceResourceId: &in.WorkspaceResourceID,
		SourceHash:          &in.BuildSourceHash,
		ImageRegistry:       &in.ImageRegistry,
		Status:              presentWorkspaceResourceBuildStatus(in.Status),
	}
}

func presentWorkspaceResourceBuildStatus(status *models.WorkspaceResourceBuildStatus) *openapi.ResourceBuildStatus {
	if status == nil {
		return nil
	}
	return &openapi.ResourceBuildStatus{
		State:           &status.State,
		Conditions:      presentConditions(status.Conditions),
		ImageUrl:        &status.ImageURL,
		BuildSourceHash: &status.BuildSourceHash,
	}
}
