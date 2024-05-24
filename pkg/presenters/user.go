package presenters

import (
	"github.com/ashishmax31/soradev-api-server/pkg/api/openapi"
	"github.com/ashishmax31/soradev-api-server/pkg/models"
)

func PresentUser(in *models.User) openapi.User {
	res := openapi.User{}
	res.SetEmail(in.Email)
	res.SetOrganisation(in.Organisation)
	res.SetId(in.ID)
	res.SetName(in.Name)
	res.SetRole(string(in.Role))
	return res
}

func ConvertUser(in *openapi.UserCreateRequest) *models.User {
	res := &models.User{
		Name:         in.Name,
		Email:        in.Email,
		Organisation: in.GetOrganisation(),
		Password:     in.GetPassword(),
	}
	return res
}
