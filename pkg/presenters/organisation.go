package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func PresentOrganisation(in *models.Organisation) openapi.Organisation {
	return openapi.Organisation{
		Id:        &in.ID,
		Name:      &in.Name,
		Domains:   PresentDomains(in.Domains),
		CreatedAt: &in.CreatedAt,
		UpdatedAt: &in.UpdatedAt,
	}

}

func PresentDomains(in []*models.OrganisationDomain) []openapi.DomainName {
	domains := make([]openapi.DomainName, len(in))
	for i, domain := range in {
		domains[i] = openapi.DomainName{
			Id:   &domain.ID,
			Fqdn: &domain.Domain,
		}
	}
	return domains
}

func ConvertOrganisation(in openapi.Organisation) *models.Organisation {
	res := &models.Organisation{}
	if in.HasDomains() {
		res.Domains = ConvertDomains(in.Domains)
	}
	if in.HasName() {
		res.Name = *in.Name
	}
	return res
}

func ConvertDomains(in []openapi.DomainName) []*models.OrganisationDomain {
	domains := make([]*models.OrganisationDomain, len(in))
	for i, domain := range in {
		domains[i] = &models.OrganisationDomain{
			Domain: domain.GetFqdn(),
		}
	}
	return domains
}
