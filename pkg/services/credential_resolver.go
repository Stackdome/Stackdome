package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/clients"
	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/clients/githubapp"
	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

// The resolver types live in pkg/credentials so packages the services layer
// depends on (e.g. the stack validator) can consume them without an import
// cycle. Aliased here so service-layer consumers keep the CredentialResolver
// naming.
type (
	CredentialResolver         = credentials.Resolver
	CredentialSource           = credentials.Source
	RegistryPurpose            = credentials.RegistryPurpose
	ResolvedGitCredential      = credentials.ResolvedGitCredential
	ResolvedRegistryCredential = credentials.ResolvedRegistryCredential
)

const (
	CredentialSourceSecretRef   = credentials.SourceSecretRef
	CredentialSourceIntegration = credentials.SourceIntegration
	CredentialSourceAnonymous   = credentials.SourceAnonymous

	RegistryPurposePull = credentials.RegistryPurposePull
	RegistryPurposePush = credentials.RegistryPurposePush
)

//go:generate mockgen -source=credential_resolver.go -destination=credential_resolver_mock_test.go -package=services
type credentialSecretFetcher interface {
	InternalGetByID(ctx context.Context, ID string) (*models.Secret, *errors.ServiceError)
}

// gitIntegrationResolverSource is the org-level git integration tier.
type gitIntegrationResolverSource interface {
	InternalGetForHost(ctx context.Context, organisationID, host string) (*models.GitIntegration, *errors.ServiceError)
	InternalGetByID(ctx context.Context, ID string) (*models.GitIntegration, *errors.ServiceError)
	// InternalMintForRepo mints a GitHub App installation token when the
	// org's app covers the repository owner (404 to fall through).
	InternalMintForRepo(ctx context.Context, organisationID, repoURL string) (*models.GitHubAppMintResult, *errors.ServiceError)
	// InternalMintCloneToken mints for an explicitly selected integration.
	InternalMintCloneToken(ctx context.Context, integrationID, repoURL string) (*models.GitHubAppMintResult, *errors.ServiceError)
}

// registryCredentialResolverSource is the org-level registry credential tier.
type registryCredentialResolverSource interface {
	InternalGetForHost(ctx context.Context, organisationID, host string, purpose models.RegistryCredentialPurpose) (*models.RegistryCredential, *errors.ServiceError)
	InternalGetByID(ctx context.Context, ID string) (*models.RegistryCredential, *errors.ServiceError)
}

type CredentialResolverSpec struct {
	SecretService credentialSecretFetcher
	// RegistryCredentialService is optional; when unset, registry resolution
	// falls through from explicit refs straight to anonymous.
	RegistryCredentialService registryCredentialResolverSource
	// GitIntegrationService is optional; when unset, git resolution falls
	// through from explicit refs straight to anonymous.
	GitIntegrationService gitIntegrationResolverSource
}

type credentialResolver struct {
	secretService       credentialSecretFetcher
	registryCredentials registryCredentialResolverSource
	gitIntegrations     gitIntegrationResolverSource
}

func NewCredentialResolver(spec CredentialResolverSpec) CredentialResolver {
	return &credentialResolver{
		secretService:       spec.SecretService,
		registryCredentials: spec.RegistryCredentialService,
		gitIntegrations:     spec.GitIntegrationService,
	}
}

// GitCredentials resolves git auth for repoURL with fixed precedence:
// explicit secret ref > explicit integration ID > org-level integration
// (GitHub App mint, then host match) > anonymous.
func (r *credentialResolver) GitCredentials(ctx context.Context, orgID, repoURL string, selector credentials.GitAuthSelector) (*ResolvedGitCredential, *errors.ServiceError) {
	switch {
	case selector.SecretRef != nil:
		secret, serr := r.secretService.InternalGetByID(ctx, selector.SecretRef.SecretID)
		if serr != nil {
			return nil, serr
		}

		resolved := &ResolvedGitCredential{
			Source:   CredentialSourceSecretRef,
			SecretID: secret.ID,
			DataHash: secret.DataHash,
			Secret:   secret,
		}
		if token, ok := secret.Data[models.TokenSecretKey]; ok && token != "" {
			resolved.Credentials = gitclient.GitCredentials{Token: token}
			return resolved, nil
		}
		resolved.Credentials = gitclient.GitCredentials{
			Username: secret.Data[models.UsernameSecretKey],
			Password: secret.Data[models.PasswordSecretKey],
		}
		return resolved, nil

	case selector.IntegrationID != "":
		if r.gitIntegrations == nil {
			return nil, errors.GeneralError("git integration resolution is not configured")
		}
		integration, serr := r.gitIntegrations.InternalGetByID(ctx, selector.IntegrationID)
		if serr != nil {
			return nil, serr
		}
		if integration.OrganisationID != orgID {
			return nil, errors.NotFound("git integration with id '%s' not found", selector.IntegrationID)
		}
		if integration.Type == models.GitIntegrationTypeGitHubApp {
			mint, serr := r.gitIntegrations.InternalMintCloneToken(ctx, integration.ID, repoURL)
			if serr != nil {
				if serr.Is404() {
					return nil, errors.BadRequest("the GitHub App installation does not cover repository '%s'", repoURL)
				}
				return nil, serr
			}
			return resolvedFromGitHubAppMint(mint), nil
		}
		return resolvedFromGitIntegration(integration), nil

	default:
		if r.gitIntegrations == nil {
			break // git integration tier not configured; fall through to anonymous
		}
		host := gitclient.RepoHost(repoURL)
		if gitclient.IsGithubHost(host) {
			mint, serr := r.gitIntegrations.InternalMintForRepo(ctx, orgID, repoURL)
			if serr == nil {
				return resolvedFromGitHubAppMint(mint), nil
			}
			if !serr.Is404() {
				return nil, serr
			}
		}
		if host != "" {
			integration, serr := r.gitIntegrations.InternalGetForHost(ctx, orgID, host)
			if serr == nil {
				return resolvedFromGitIntegration(integration), nil
			}
			if !serr.Is404() {
				return nil, serr
			}
		}
	}

	return &ResolvedGitCredential{Source: CredentialSourceAnonymous}, nil
}

func resolvedFromGitIntegration(integration *models.GitIntegration) *ResolvedGitCredential {
	return &ResolvedGitCredential{
		Source:        CredentialSourceIntegration,
		Credentials:   gitCredentialsFromAuth(integration.Auth),
		DataHash:      integration.DataHash,
		IntegrationID: integration.ID,
		Host:          integration.Host,
	}
}

func resolvedFromGitHubAppMint(mint *models.GitHubAppMintResult) *ResolvedGitCredential {
	mintedAt := mint.MintedAt
	expiresAt := mint.ExpiresAt
	return &ResolvedGitCredential{
		Source: CredentialSourceIntegration,
		Credentials: gitclient.GitCredentials{
			Username: githubapp.CloneUsername,
			Password: mint.Token,
		},
		DataHash:       mint.DataHash,
		IntegrationID:  mint.IntegrationID,
		Host:           mint.Host,
		TokenMintedAt:  &mintedAt,
		TokenExpiresAt: &expiresAt,
	}
}

// RegistryCredentials resolves registry auth for an image or repository ref
// with fixed precedence: explicit secret ref > explicit org registry
// credential ID > org-level registry credential matched by normalized host >
// anonymous.
func (r *credentialResolver) RegistryCredentials(ctx context.Context, orgID, ref string, purpose RegistryPurpose, selector credentials.RegistryAuthSelector) (*ResolvedRegistryCredential, *errors.ServiceError) {
	if selector.SecretRef != nil {
		secret, serr := r.secretService.InternalGetByID(ctx, selector.SecretRef.SecretID)
		if serr != nil {
			return nil, serr
		}

		return &ResolvedRegistryCredential{
			Source:   CredentialSourceSecretRef,
			Username: secret.Data[models.UsernameSecretKey],
			Password: secret.Data[models.PasswordSecretKey],
			SecretID: secret.ID,
			DataHash: secret.DataHash,
			Secret:   secret,
			Purpose:  purpose,
		}, nil
	}

	if selector.RegistryCredentialID != "" {
		if r.registryCredentials == nil {
			return nil, errors.GeneralError("registry credential resolution is not configured")
		}
		cred, serr := r.registryCredentials.InternalGetByID(ctx, selector.RegistryCredentialID)
		if serr != nil {
			return nil, serr
		}
		if cred.OrganisationID != orgID {
			return nil, errors.NotFound("registry credential with id '%s' not found", selector.RegistryCredentialID)
		}
		if !cred.Purpose.Covers(models.RegistryCredentialPurpose(purpose)) {
			return nil, errors.BadRequest("registry credential '%s' has purpose '%s' and cannot be used for '%s'", cred.ID, cred.Purpose, purpose)
		}
		return resolvedFromRegistryCredential(cred, purpose), nil
	}

	if r.registryCredentials != nil {
		if host, err := clients.NormalizeRegistryHost(ref); err == nil {
			cred, serr := r.registryCredentials.InternalGetForHost(ctx, orgID, host, models.RegistryCredentialPurpose(purpose))
			if serr == nil {
				return resolvedFromRegistryCredential(cred, purpose), nil
			}
			if !serr.Is404() {
				return nil, serr
			}
		}
	}

	return &ResolvedRegistryCredential{Source: CredentialSourceAnonymous}, nil
}

func resolvedFromRegistryCredential(cred *models.RegistryCredential, purpose RegistryPurpose) *ResolvedRegistryCredential {
	return &ResolvedRegistryCredential{
		Source:       CredentialSourceIntegration,
		Username:     cred.Username,
		Password:     cred.Password,
		DataHash:     cred.DataHash,
		CredentialID: cred.ID,
		Host:         cred.Host,
		Purpose:      purpose,
	}
}
