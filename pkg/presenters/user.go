package presenters

import (
	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
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

func PresentUserWithProjects(in *models.User, memberships []*models.ProjectMembership) openapi.User {
	res := PresentUser(in)
	projects := make([]openapi.UserProjectMembership, len(memberships))
	for i, m := range memberships {
		tm := openapi.UserProjectMembership{}
		tm.SetProjectId(m.ProjectID)
		tm.SetRole(string(m.Role))
		if m.Project != nil {
			tm.SetProjectName(m.Project.Name)
			tm.SetDefaultProject(m.Project.DefaultProject)
		}
		projects[i] = tm
	}
	res.SetProjects(projects)
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
