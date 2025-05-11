package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func PresentOrganisation(in *models.Organisation) openapi.Organisation {
	return openapi.Organisation{
		Id:         &in.ID,
		Name:       &in.Name,
		DomainName: &in.DomainName,
		IsDefault:  &in.Default,
		CreatedAt:  &in.CreatedAt,
		UpdatedAt:  &in.UpdatedAt,
	}

}

func ConvertOrganisation(in openapi.Organisation) *models.Organisation {
	res := &models.Organisation{
		Default: false,
	}
	if in.HasDomainName() {
		res.DomainName = *in.DomainName
	}
	if in.HasName() {
		res.Name = *in.Name
	}
	return res
}
