package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func ConvertSecret(in *openapi.Secret) *models.Secret {
	res := &models.Secret{
		Name:        in.GetName(),
		Description: in.GetDescription(),
		Type:        convertSecretType(in.GetType()),
		Keys:        keys(convertSecretData(in.GetData())),
		Data:        convertSecretData(in.GetData()),
	}
	return res
}

func convertSecretData(data []openapi.SecretData) map[string]string {
	res := make(map[string]string)
	for _, d := range data {
		res[d.GetKey()] = d.GetValue()
	}
	return res
}

func convertSecretType(t openapi.SecretType) models.SecretType {
	switch t {
	case openapi.GENERIC:
		return models.SecretTypeGeneric
	case openapi.DOCKER_REGISTRY:
		return models.SecretTypeDockerRegistry
	case openapi.GIT_CREDENTIALS:
		return models.SecretTypeGitCredentials
	case openapi.USERNAME_PASSWORD:
		return models.SecretTypeUsernamePassword
	case openapi.TOKEN:
		return models.SecretTypeToken
	case openapi.SSH_KEY:
		return models.SecretTypeSSHKey
	default:
		return models.SecretTypeGeneric
	}
}

func PresentSecret(in *models.Secret) openapi.Secret {
	res := openapi.Secret{}
	res.SetId(in.ID)
	res.SetName(in.Name)
	// we only present the keys in the secret, not the values
	res.SetData(presentSecretData(in.Keys))
	res.SetType(presentSecretType(in.Type))
	res.SetDescription(in.Description)
	res.SetOrganisationId(in.OrganisationID)
	res.SetTeamId(in.TeamID)
	res.SetOutputs(presentOutputDescriptors(in.EnsureDeclaredOutputs()))
	res.SetCreatedAt(in.CreatedAt)
	res.SetUpdatedAt(in.UpdatedAt)
	return res
}

func PresentSecretList(in []*models.Secret) []openapi.Secret {
	var res []openapi.Secret
	for _, secret := range in {
		res = append(res, PresentSecret(secret))
	}
	return res
}

func presentSecretData(keys []string) []openapi.SecretData {
	var res []openapi.SecretData
	for _, key := range keys {
		res = append(res, openapi.SecretData{
			Key:   key,
			Value: "******",
		})
	}
	return res
}

func presentSecretType(t models.SecretType) openapi.SecretType {
	switch t {
	case models.SecretTypeGeneric:
		return openapi.GENERIC
	case models.SecretTypeDockerRegistry:
		return openapi.DOCKER_REGISTRY
	case models.SecretTypeGitCredentials:
		return openapi.GIT_CREDENTIALS
	case models.SecretTypeUsernamePassword:
		return openapi.USERNAME_PASSWORD
	case models.SecretTypeToken:
		return openapi.TOKEN
	case models.SecretTypeSSHKey:
		return openapi.SSH_KEY
	default:
		return openapi.GENERIC
	}
}

func keys(data map[string]string) []string {
	var res []string
	for key := range data {
		res = append(res, key)
	}
	return res
}
