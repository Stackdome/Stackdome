package presenters

import (
	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
)

func PresentProject(in *models.Project) openapi.Project {
	t := openapi.Project{
		Id:             &in.ID,
		Name:           in.Name,
		OrganisationId: &in.OrganisationID,
		DefaultProject:    &in.DefaultProject,
		CreatedAt:      &in.CreatedAt,
		UpdatedAt:      &in.UpdatedAt,
	}
	return t
}

func PresentProjectList(in []*models.Project) []openapi.Project {
	out := make([]openapi.Project, len(in))
	for i, t := range in {
		out[i] = PresentProject(t)
	}
	return out
}

func PresentProjectMembership(in *models.ProjectMembership) openapi.ProjectMembership {
	m := openapi.ProjectMembership{
		Id:        &in.ID,
		ProjectId:    in.ProjectID,
		UserId:    in.UserID,
		Role:      string(in.Role),
		CreatedAt: &in.CreatedAt,
	}
	if in.Project != nil {
		t := PresentProject(in.Project)
		m.Project = &t
	}
	if in.User != nil {
		u := PresentUser(in.User)
		m.User = &u
	}
	return m
}

func PresentProjectMembershipList(in []*models.ProjectMembership) []openapi.ProjectMembership {
	out := make([]openapi.ProjectMembership, len(in))
	for i, m := range in {
		out[i] = PresentProjectMembership(m)
	}
	return out
}

func ConvertProjectRole(role openapi.ProjectRole) models.ProjectRole {
	return models.ProjectRole(role)
}
