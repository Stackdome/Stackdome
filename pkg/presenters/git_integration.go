package presenters

import (
	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/clients/githubapp"
	"github.com/Stackdome/stackdome/pkg/models"
)

func ConvertGitIntegration(in *openapi.GitIntegration) *models.GitIntegration {
	res := &models.GitIntegration{
		Type: convertGitIntegrationType(in.GetType()),
		Host: in.GetHost(),
	}
	if in.Auth != nil {
		auth := &models.GitIntegrationAuth{
			Token: in.Auth.GetToken(),
		}
		if in.Auth.Basic != nil {
			auth.Basic = &models.GitIntegrationBasicAuth{
				Username: in.Auth.Basic.Username,
				Password: in.Auth.Basic.Password,
			}
		}
		res.Auth = auth
	}
	return res
}

func convertGitIntegrationType(t openapi.GitIntegrationType) models.GitIntegrationType {
	switch t {
	case openapi.GIT_INTEGRATION_TYPE_GITHUB_APP:
		return models.GitIntegrationTypeGitHubApp
	default:
		// The schema default; also applied when the field is omitted.
		return models.GitIntegrationTypeGitCredentials
	}
}

func presentGitIntegrationType(t models.GitIntegrationType) openapi.GitIntegrationType {
	switch t {
	case models.GitIntegrationTypeGitHubApp:
		return openapi.GIT_INTEGRATION_TYPE_GITHUB_APP
	default:
		return openapi.GIT_INTEGRATION_TYPE_CREDENTIALS
	}
}

// PresentGitIntegration never includes auth material: neither the transient
// decrypted blob nor the encrypted form is exposed.
func PresentGitIntegration(in *models.GitIntegration) openapi.GitIntegration {
	res := openapi.GitIntegration{}
	res.SetId(in.ID)
	res.SetType(presentGitIntegrationType(in.Type))
	res.SetHost(in.Host)
	res.SetStatus(string(in.Status))
	res.SetOrganisationId(in.OrganisationID)
	if in.EncryptedAuth != "" {
		res.SetCredentialsConfigured(true)
	}
	if in.InstallURL != "" {
		res.SetInstallUrl(in.InstallURL)
	}
	res.SetCreatedAt(in.CreatedAt)
	res.SetUpdatedAt(in.UpdatedAt)
	return res
}

func PresentGitIntegrationList(in []*models.GitIntegration) []openapi.GitIntegration {
	var res []openapi.GitIntegration
	for _, integration := range in {
		res = append(res, PresentGitIntegration(integration))
	}
	return res
}

// PresentGitHubAppManifestFlow exposes only what the UI needs to start the
// GitHub App creation; app credentials never appear here (they don't exist
// until the callback).
func PresentGitHubAppManifestFlow(in *models.GitHubAppManifestFlow) openapi.GitHubAppManifestFlow {
	res := openapi.GitHubAppManifestFlow{}
	res.SetManifest(in.Manifest)
	res.SetGithubUrl(in.GitHubURL)
	res.SetState(in.State)
	return res
}

func PresentGitInstallation(in *models.GitInstallation) openapi.GitInstallation {
	res := openapi.GitInstallation{}
	res.SetId(in.ID)
	res.SetInstallationId(in.InstallationID)
	res.SetAccountLogin(in.AccountLogin)
	res.SetAccountType(string(in.AccountType))
	res.SetRepositorySelection(in.RepositorySelection)
	res.SetCreatedAt(in.CreatedAt)
	return res
}

func PresentGitInstallationList(in []*models.GitInstallation) []openapi.GitInstallation {
	var res []openapi.GitInstallation
	for _, installation := range in {
		res = append(res, PresentGitInstallation(installation))
	}
	return res
}

func PresentGitRepository(in githubapp.Repo) openapi.GitRepository {
	res := openapi.GitRepository{}
	res.SetFullName(in.FullName)
	res.SetCloneUrl(in.CloneURL)
	res.SetDefaultBranch(in.DefaultBranch)
	res.SetPrivate(in.Private)
	res.SetOwner(in.OwnerLogin)
	if in.PushedAt != nil {
		res.SetPushedAt(*in.PushedAt)
	}
	return res
}

func PresentGitRepositoryPage(in *githubapp.RepoPage) openapi.GitRepositoryPage {
	res := openapi.GitRepositoryPage{}
	var items []openapi.GitRepository
	for _, repo := range in.Repos {
		items = append(items, PresentGitRepository(repo))
	}
	res.SetItems(items)
	res.SetPage(int32(in.Page))
	res.SetTotalCount(int32(in.TotalCount))
	res.SetHasNext(in.HasNext)
	return res
}
