package presenters

import (
	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
)

func ConvertRegistryCredential(in *openapi.RegistryCredential) *models.RegistryCredential {
	return &models.RegistryCredential{
		Host:     in.GetHost(),
		Purpose:  convertRegistryCredentialPurpose(in.GetPurpose()),
		Username: in.GetUsername(),
		Password: in.GetPassword(),
	}
}

func convertRegistryCredentialPurpose(p openapi.RegistryCredentialPurpose) models.RegistryCredentialPurpose {
	switch p {
	case openapi.PULL:
		return models.RegistryCredentialPurposePull
	case openapi.PUSH:
		return models.RegistryCredentialPurposePush
	case openapi.BOTH:
		return models.RegistryCredentialPurposeBoth
	default:
		return ""
	}
}

// ConvertRegistryCredentialVerifyPurpose returns the requested purpose, or
// empty when the request omits it so the service falls back to the
// credential's own purpose.
func ConvertRegistryCredentialVerifyPurpose(in *openapi.RegistryCredentialVerifyRequest) models.RegistryCredentialPurpose {
	if in.Purpose == nil {
		return ""
	}
	return convertRegistryCredentialPurpose(in.GetPurpose())
}

func presentRegistryCredentialPurpose(p models.RegistryCredentialPurpose) openapi.RegistryCredentialPurpose {
	switch p {
	case models.RegistryCredentialPurposePull:
		return openapi.PULL
	case models.RegistryCredentialPurposePush:
		return openapi.PUSH
	default:
		return openapi.BOTH
	}
}

// PresentRegistryCredential never includes the password: the transient
// plaintext field is dropped and the encrypted form is not exposed.
func PresentRegistryCredential(in *models.RegistryCredential) openapi.RegistryCredential {
	res := openapi.RegistryCredential{}
	res.SetId(in.ID)
	res.SetHost(in.Host)
	res.SetPurpose(presentRegistryCredentialPurpose(in.Purpose))
	res.SetUsername(in.Username)
	res.SetOrganisationId(in.OrganisationID)
	res.SetCreatedAt(in.CreatedAt)
	res.SetUpdatedAt(in.UpdatedAt)
	return res
}

func PresentRegistryCredentialList(in []*models.RegistryCredential) []openapi.RegistryCredential {
	var res []openapi.RegistryCredential
	for _, credential := range in {
		res = append(res, PresentRegistryCredential(credential))
	}
	return res
}
