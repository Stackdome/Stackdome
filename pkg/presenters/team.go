package presenters

import (
	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
)

func PresentTeam(in *models.Team) openapi.Team {
	t := openapi.Team{
		Id:             &in.ID,
		Name:           in.Name,
		OrganisationId: &in.OrganisationID,
		DefaultTeam:    &in.DefaultTeam,
		CreatedAt:      &in.CreatedAt,
		UpdatedAt:      &in.UpdatedAt,
	}
	return t
}

func PresentTeamList(in []*models.Team) []openapi.Team {
	out := make([]openapi.Team, len(in))
	for i, t := range in {
		out[i] = PresentTeam(t)
	}
	return out
}

func PresentTeamMembership(in *models.TeamMembership) openapi.TeamMembership {
	m := openapi.TeamMembership{
		Id:        &in.ID,
		TeamId:    in.TeamID,
		UserId:    in.UserID,
		Role:      string(in.Role),
		CreatedAt: &in.CreatedAt,
	}
	if in.Team != nil {
		t := PresentTeam(in.Team)
		m.Team = &t
	}
	if in.User != nil {
		u := PresentUser(in.User)
		m.User = &u
	}
	return m
}

func PresentTeamMembershipList(in []*models.TeamMembership) []openapi.TeamMembership {
	out := make([]openapi.TeamMembership, len(in))
	for i, m := range in {
		out[i] = PresentTeamMembership(m)
	}
	return out
}

func ConvertTeamRole(role openapi.TeamRole) models.TeamRole {
	return models.TeamRole(role)
}
