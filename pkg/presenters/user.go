package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func PresentUser(in *models.User) openapi.User {
	res := openapi.User{}
	res.SetEmail(in.Email)
	if in.Organisation != nil {
		res.SetOrganisation(in.Organisation.Name)
	}
	res.SetId(in.ID)
	res.SetName(in.Name)
	res.SetRole(PresentRole(in.Role))
	res.SetOrganisationId(in.OrganisationID)
	return res
}

func PresentUserWithTeams(in *models.User, memberships []*models.TeamMembership) openapi.User {
	res := PresentUser(in)
	teams := make([]openapi.UserTeamMembership, len(memberships))
	for i, m := range memberships {
		tm := openapi.UserTeamMembership{}
		tm.SetTeamId(m.TeamID)
		tm.SetRole(string(m.Role))
		if m.Team != nil {
			tm.SetTeamName(m.Team.Name)
			tm.SetDefaultTeam(m.Team.DefaultTeam)
		}
		teams[i] = tm
	}
	res.SetTeams(teams)
	return res
}

func PresentRole(in models.UserRole) openapi.UserRole {
	switch in {
	case models.OrgAdminRole:
		return openapi.ORG_ADMIN
	case models.OrgMemberRole:
		return openapi.ORG_MEMBER
	default:
		return openapi.ORG_MEMBER
	}
}

func ConvertUser(in *openapi.UserSignupRequest) *models.User {
	res := &models.User{
		Name:     in.Name,
		Email:    in.Email,
		Password: in.GetPassword(),
	}
	if in.Organisation != nil {
		res.Organisation = &models.Organisation{
			Name:    in.Organisation.GetName(),
			Domains: ConvertDomains(in.Organisation.GetDomains()),
		}
	}
	return res
}
