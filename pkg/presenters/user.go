package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func PresentUser(in *models.User) openapi.User {
	res := openapi.User{}
	res.SetEmail(in.Email)
	res.SetOrganisation(in.Organisation.Name)
	res.SetId(in.ID)
	res.SetName(in.Name)
	res.SetRole(PresentRole(in.Role))
	res.SetOrganisationId(in.OrganisationID)
	return res
}

func PresentRole(in models.Role) openapi.UserRole {
	switch in {
	case models.OrganisationAdminRole:
		return openapi.ORGANISATION_ADMIN
	case models.PlatformAdminRole:
		return openapi.PLATFORM_ADMIN
	default:
		return openapi.USER
	}
}

func ConvertUser(in *openapi.UserSignupRequest) *models.User {
	res := &models.User{
		Name:     in.Name,
		Email:    in.Email,
		Password: in.GetPassword(),
		Role:     convertRole(in.GetRole()),
	}
	if in.HasOrganisationId() {
		res.OrganisationID = *in.OrganisationId
	} else if in.HasOrganisation() {
		res.Organisation = &models.Organisation{
			Name:    in.Organisation.GetName(),
			Domains: ConvertDomains(in.Organisation.GetDomains()),
		}
	}
	return res
}

func convertRole(in openapi.UserRole) models.Role {
	switch in {
	case openapi.ORGANISATION_ADMIN:
		return models.OrganisationAdminRole
	case openapi.PLATFORM_ADMIN:
		return models.PlatformAdminRole
	default:
		return models.UserRole
	}
}
