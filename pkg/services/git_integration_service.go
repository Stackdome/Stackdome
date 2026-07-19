package services

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/Stackdome/stackdome/pkg/auth"
	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/clients/githubapp"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
)

//go:generate mockgen -source=git_integration_service.go -destination=git_integration_service_mock_test.go -package=services -exclude_interfaces GitIntegrationService
type verifyGitClientProvider interface {
	ClientFor(repoURL string, creds gitclient.GitCredentials) (gitclient.GitClient, error)
}

type defaultVerifyGitClientProvider struct{}

func (defaultVerifyGitClientProvider) ClientFor(repoURL string, creds gitclient.GitCredentials) (gitclient.GitClient, error) {
	return gitclient.NewGitClientForRepo(repoURL, creds)
}

//go:generate mockgen -destination=../mocks/mock_git_integration_service.go -package=mocks github.com/Stackdome/stackdome/pkg/services GitIntegrationService
type GitIntegrationService interface {
	Create(ctx context.Context, integration *models.GitIntegration) (*models.GitIntegration, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.GitIntegration, *errors.ServiceError)
	ListByOrgID(ctx context.Context, organisationID string) ([]*models.GitIntegration, *errors.ServiceError)
	Update(ctx context.Context, ID string, integration *models.GitIntegration) (*models.GitIntegration, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	Verify(ctx context.Context, ID, repoURL string) *errors.ServiceError

	// InternalGetForHost returns the decrypted git_credentials integration for
	// (org, host). No permission checks; used by the credential resolver.
	InternalGetForHost(ctx context.Context, organisationID, host string) (*models.GitIntegration, *errors.ServiceError)
	// InternalGetByID returns the decrypted integration by ID. No permission
	// checks; used by the credential resolver for explicit overrides.
	InternalGetByID(ctx context.Context, ID string) (*models.GitIntegration, *errors.ServiceError)

	// GitHub App manifest flow, installations, and discovery.
	CreateGitHubAppManifest(ctx context.Context, organisationID string) (*models.GitHubAppManifestFlow, *errors.ServiceError)
	HandleGitHubManifestCallback(ctx context.Context, code, state string) (string, *errors.ServiceError)
	ListInstallations(ctx context.Context, integrationID string, refresh bool) ([]*models.GitInstallation, *errors.ServiceError)
	ListRepositories(ctx context.Context, integrationID string, page int, installationUUID string) (*githubapp.RepoPage, *errors.ServiceError)
	GetRepository(ctx context.Context, integrationID, owner, repo string) (*githubapp.Repo, *errors.ServiceError)
	ListRepositoryBranches(ctx context.Context, integrationID, owner, repo string) ([]string, *errors.ServiceError)
	ProcessGitHubWebhook(ctx context.Context, event string, payload []byte, signature string) *errors.ServiceError

	// InternalMintForRepo mints an installation token for the org's installed
	// GitHub App when it covers the repository owner (404 to fall through).
	InternalMintForRepo(ctx context.Context, organisationID, repoURL string) (*models.GitHubAppMintResult, *errors.ServiceError)
	// InternalMintCloneToken re-mints a token for a known integration; used
	// by the hub-side token refresher.
	InternalMintCloneToken(ctx context.Context, integrationID, repoURL string) (*models.GitHubAppMintResult, *errors.ServiceError)
}

type GitIntegrationServiceSpec struct {
	Store             stores.GitIntegrationStore
	InstallationStore stores.GitInstallationStore
	OAuthStateStore   stores.OAuthStateStore
	OrganisationStore stores.OrganisationStore
	AtomicExecutor    stores.AtomicExecutor
	GitHubAppClient   githubapp.Client
	EncryptionService EncryptionService
	Permissions       auth.PermissionService
	PreviewWebhook    PreviewWebhookService
	Logger            logger.Logger
	// ExternalURL is the hub's externally reachable base URL, required for
	// the GitHub App manifest flow.
	ExternalURL string
	// GitClients is optional; it defaults to real git clients.
	GitClients verifyGitClientProvider
}

type gitIntegrationService struct {
	store             stores.GitIntegrationStore
	installations     stores.GitInstallationStore
	oauthStates       stores.OAuthStateStore
	organisations     stores.OrganisationStore
	atomic            stores.AtomicExecutor
	githubApp         githubapp.Client
	encryptionService EncryptionService
	permissions       auth.PermissionService
	previewWebhook    PreviewWebhookService
	logger            logger.Logger
	externalURL       string
	gitClients        verifyGitClientProvider
}

func NewGitIntegrationService(spec GitIntegrationServiceSpec) GitIntegrationService {
	gitClients := spec.GitClients
	if gitClients == nil {
		gitClients = defaultVerifyGitClientProvider{}
	}
	return &gitIntegrationService{
		store:             spec.Store,
		installations:     spec.InstallationStore,
		oauthStates:       spec.OAuthStateStore,
		organisations:     spec.OrganisationStore,
		atomic:            spec.AtomicExecutor,
		githubApp:         spec.GitHubAppClient,
		encryptionService: spec.EncryptionService,
		permissions:       spec.Permissions,
		previewWebhook:    spec.PreviewWebhook,
		logger:            spec.Logger,
		externalURL:       strings.TrimSuffix(spec.ExternalURL, "/"),
		gitClients:        gitClients,
	}
}

func (s *gitIntegrationService) SetPreviewWebhook(pw PreviewWebhookService) {
	s.previewWebhook = pw
}

func (s *gitIntegrationService) Create(ctx context.Context, integration *models.GitIntegration) (*models.GitIntegration, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, integration.OrganisationID, auth.ResourceGitIntegrations, "", auth.ActionCreate); permErr != nil {
		return nil, permErr
	}

	if integration.Type == "" {
		integration.Type = models.GitIntegrationTypeGitCredentials
	}
	if !integration.Type.Valid() {
		return nil, errors.BadRequest("unsupported git integration type '%s'", integration.Type)
	}
	if integration.Type == models.GitIntegrationTypeGitHubApp {
		return nil, errors.BadRequest("github_app integrations are created through the GitHub App manifest flow")
	}
	if integration.Host == "" {
		return nil, errors.BadRequest("host is required")
	}
	integration.Host = gitclient.RepoHost(integration.Host)
	if integration.Auth == nil || integration.Auth.Empty() {
		return nil, errors.BadRequest("auth credentials are required")
	}
	if integration.Auth.Token != "" && integration.Auth.Basic != nil {
		return nil, errors.BadRequest("only one of token or basic auth may be set")
	}
	if integration.Auth.Basic != nil && (integration.Auth.Basic.Username == "" || integration.Auth.Basic.Password == "") {
		return nil, errors.BadRequest("basic auth requires username and password")
	}

	if _, serr := s.store.GetByOrgAndHost(ctx, integration.OrganisationID, integration.Host); serr == nil {
		return nil, errors.Conflict("a git integration for host '%s' already exists", integration.Host)
	} else if !serr.Is404() {
		return nil, serr
	}

	integration.Status = models.GitIntegrationStatusActive
	if serr := s.sealIntegration(integration); serr != nil {
		return nil, serr
	}

	return s.store.Create(ctx, integration)
}

func (s *gitIntegrationService) GetByID(ctx context.Context, ID string) (*models.GitIntegration, *errors.ServiceError) {
	integration, serr := s.store.GetByID(ctx, ID)
	if serr != nil {
		return nil, serr
	}
	if permErr := s.permissions.Check(ctx, integration.OrganisationID, auth.ResourceGitIntegrations, ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	s.attachInstallURL(integration)
	return integration, nil
}

func (s *gitIntegrationService) ListByOrgID(ctx context.Context, organisationID string) ([]*models.GitIntegration, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, organisationID, auth.ResourceGitIntegrations, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	integrations, serr := s.store.ListByOrgID(ctx, organisationID)
	if serr != nil {
		return nil, serr
	}
	for _, integration := range integrations {
		s.attachInstallURL(integration)
	}
	return integrations, nil
}

func (s *gitIntegrationService) Update(ctx context.Context, ID string, integration *models.GitIntegration) (*models.GitIntegration, *errors.ServiceError) {
	existing, serr := s.store.GetByID(ctx, ID)
	if serr != nil {
		return nil, serr
	}
	if permErr := s.permissions.Check(ctx, existing.OrganisationID, auth.ResourceGitIntegrations, ID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}

	if existing.Type != models.GitIntegrationTypeGitCredentials {
		return nil, errors.BadRequest("only git_credentials integrations can be updated directly")
	}
	if integration.Type != "" && integration.Type != existing.Type {
		return nil, errors.BadRequest("type cannot be updated")
	}
	if integration.Host != "" && gitclient.RepoHost(integration.Host) != existing.Host {
		return nil, errors.BadRequest("host cannot be updated")
	}

	// Credential rotation: omit auth to keep the stored credentials.
	if integration.Auth != nil && !integration.Auth.Empty() {
		if integration.Auth.Token != "" && integration.Auth.Basic != nil {
			return nil, errors.BadRequest("only one of token or basic auth may be set")
		}
		existing.Auth = integration.Auth
		if serr := s.sealIntegration(existing); serr != nil {
			return nil, serr
		}
	}

	return s.store.Update(ctx, existing)
}

func (s *gitIntegrationService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	integration, serr := s.store.GetByID(ctx, ID)
	if serr != nil {
		return serr
	}
	if permErr := s.permissions.Check(ctx, integration.OrganisationID, auth.ResourceGitIntegrations, ID, auth.ActionDelete); permErr != nil {
		return permErr
	}
	return s.store.Delete(ctx, ID)
}

func (s *gitIntegrationService) Verify(ctx context.Context, ID, repoURL string) *errors.ServiceError {
	integration, serr := s.store.GetByID(ctx, ID)
	if serr != nil {
		return serr
	}
	if permErr := s.permissions.Check(ctx, integration.OrganisationID, auth.ResourceGitIntegrations, ID, auth.ActionRead); permErr != nil {
		return permErr
	}

	if repoURL == "" {
		return errors.BadRequest("repo_url is required")
	}
	if integration.Type != models.GitIntegrationTypeGitCredentials {
		return errors.BadRequest("only git_credentials integrations can be verified directly")
	}
	if host := gitclient.RepoHost(repoURL); host != integration.Host {
		return errors.BadRequest("repository host '%s' does not match integration host '%s'", host, integration.Host)
	}

	if serr := s.unsealIntegration(integration); serr != nil {
		return serr
	}

	client, err := s.gitClients.ClientFor(repoURL, gitCredentialsFromAuth(integration.Auth))
	if err != nil {
		return errors.GeneralError("failed to create git client: %s", err.Error())
	}
	if _, err := client.CheckAccess(ctx, repoURL); err != nil {
		return errors.BadRequest("verification failed for '%s': %s", repoURL, err.Error())
	}
	return nil
}

func (s *gitIntegrationService) InternalGetForHost(ctx context.Context, organisationID, host string) (*models.GitIntegration, *errors.ServiceError) {
	integration, serr := s.store.GetByOrgAndHost(ctx, organisationID, host)
	if serr != nil {
		return nil, serr
	}
	if serr := s.unsealIntegration(integration); serr != nil {
		return nil, serr
	}
	return integration, nil
}

func (s *gitIntegrationService) InternalGetByID(ctx context.Context, ID string) (*models.GitIntegration, *errors.ServiceError) {
	integration, serr := s.store.GetByID(ctx, ID)
	if serr != nil {
		return nil, serr
	}
	if integration.Type != models.GitIntegrationTypeGitCredentials {
		// github_app auth blobs hold app credentials, not clone credentials.
		return integration, nil
	}
	if serr := s.unsealIntegration(integration); serr != nil {
		return nil, serr
	}
	return integration, nil
}

// sealIntegration encrypts the transient auth blob and stamps the data hash.
func (s *gitIntegrationService) sealIntegration(integration *models.GitIntegration) *errors.ServiceError {
	blob, err := integration.Auth.Marshal()
	if err != nil {
		return errors.GeneralError("failed to serialize git integration auth: %s", err.Error())
	}
	encrypted, err := s.encryptionService.EncryptData(blob)
	if err != nil {
		return errors.GeneralError("failed to encrypt git integration auth: %s", err.Error())
	}
	integration.EncryptedAuth = encrypted
	integration.DataHash = gitIntegrationDataHash(blob)
	integration.Auth = nil
	return nil
}

// unsealIntegration decrypts the stored auth blob onto the transient Auth
// field, verifying integrity against the data hash.
func (s *gitIntegrationService) unsealIntegration(integration *models.GitIntegration) *errors.ServiceError {
	decrypted, err := s.encryptionService.DecryptData(integration.EncryptedAuth)
	if err != nil {
		return errors.GeneralError("failed to decrypt git integration auth: %s", err.Error())
	}
	if gitIntegrationDataHash(decrypted) != integration.DataHash {
		return errors.GeneralError("data integrity check failed for git integration %s", integration.ID)
	}
	var auth models.GitIntegrationAuth
	if err := json.Unmarshal(decrypted, &auth); err != nil {
		return errors.GeneralError("failed to deserialize git integration auth: %s", err.Error())
	}
	integration.Auth = &auth
	return nil
}

// gitCredentialsFromAuth maps a decrypted auth blob to git client credentials.
func gitCredentialsFromAuth(auth *models.GitIntegrationAuth) gitclient.GitCredentials {
	if auth == nil {
		return gitclient.GitCredentials{}
	}
	if auth.Token != "" {
		return gitclient.GitCredentials{Token: auth.Token}
	}
	if auth.Basic != nil {
		return gitclient.GitCredentials{Username: auth.Basic.Username, Password: auth.Basic.Password}
	}
	return gitclient.GitCredentials{}
}

func gitIntegrationDataHash(blob []byte) string {
	hash := sha256.Sum256(blob)
	return base64.StdEncoding.EncodeToString(hash[:])
}
