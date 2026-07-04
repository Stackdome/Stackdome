package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	gitclient "github.com/ashishmax31/stackdome-api-server/pkg/clients/git"
	"github.com/ashishmax31/stackdome-api-server/pkg/clients/githubapp"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/google/uuid"
)

const (
	// manifestStateTTL bounds how long a manifest-flow state stays valid.
	manifestStateTTL = 15 * time.Minute
	// githubHost is the host github_app integrations cover.
	githubHost = "github.com"

	githubAppsBaseURL          = "https://github.com"
	manifestCallbackPath       = "/api/v1/git-integrations/github/manifest/callback"
	githubWebhookPath          = "/api/v1/webhooks/github"
	githubWebhookSignaturePref = "sha256="

	// GitHub webhook event and action names.
	GitHubEventInstallation      = "installation"
	GitHubEventPush              = "push"
	githubInstallationActCreated = "created"
	githubInstallationActDeleted = "deleted"
)

// CreateGitHubAppManifest starts the manifest flow: it creates (or reuses) the
// org's pending github_app integration row and returns the manifest payload
// plus the single-use state for the redirect round trip.
func (s *gitIntegrationService) CreateGitHubAppManifest(ctx context.Context, organisationID string) (*models.GitHubAppManifestFlow, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, organisationID, auth.ResourceGitIntegrations, "", auth.ActionCreate); permErr != nil {
		return nil, permErr
	}
	if s.externalURL == "" {
		return nil, errors.BadRequest("SERVER_EXTERNAL_URL is not configured; the GitHub App flow requires an externally reachable hub URL")
	}

	integration, serr := s.store.GetGitHubAppForOrg(ctx, organisationID)
	if serr != nil {
		if !serr.Is404() {
			return nil, serr
		}
		integration, serr = s.store.Create(ctx, &models.GitIntegration{
			OrganisationID: organisationID,
			Type:           models.GitIntegrationTypeGitHubApp,
			Host:           githubHost,
			Status:         models.GitIntegrationStatusPendingInstall,
			DataHash:       gitIntegrationDataHash(nil),
		})
		if serr != nil {
			return nil, serr
		}
	} else if integration.Status == models.GitIntegrationStatusInstalled {
		return nil, errors.Conflict("a GitHub App is already installed for this organisation")
	}

	state := fmt.Sprintf("%s:%s", uuid.NewString(), integration.ID)
	if serr := s.oauthStates.Create(ctx, &models.OAuthState{
		State:     state,
		Provider:  models.OAuthProviderGitHubAppManifest,
		CreatedAt: time.Now().UTC(),
	}); serr != nil {
		return nil, serr
	}

	hub := strings.TrimSuffix(s.externalURL, "/")
	manifest := map[string]any{
		"name":         githubAppName(organisationID),
		"url":          hub,
		"redirect_url": hub + manifestCallbackPath,
		"setup_url":    hub,
		"public":       true,
		"hook_attributes": map[string]any{
			"url": hub + githubWebhookPath,
		},
		"default_permissions": map[string]any{
			"contents": "read",
			"metadata": "read",
		},
		"default_events": []string{GitHubEventPush, GitHubEventInstallation},
	}

	return &models.GitHubAppManifestFlow{
		Manifest:  manifest,
		GitHubURL: fmt.Sprintf("%s/settings/apps/new?state=%s", githubAppsBaseURL, state),
		State:     state,
	}, nil
}

func githubAppName(organisationID string) string {
	suffix := organisationID
	if len(suffix) > 8 {
		suffix = suffix[:8]
	}
	return "stackdome-" + suffix
}

// HandleGitHubManifestCallback finishes the manifest flow: it validates the
// single-use state, converts the temporary code into app credentials, seals
// them onto the integration, and returns the GitHub install URL to redirect
// the browser to. Unauthenticated: the state is the proof of initiation.
func (s *gitIntegrationService) HandleGitHubManifestCallback(ctx context.Context, code, state string) (string, *errors.ServiceError) {
	if code == "" || state == "" {
		return "", errors.BadRequest("code and state are required")
	}

	record, serr := s.oauthStates.Consume(ctx, state, models.OAuthProviderGitHubAppManifest)
	if serr != nil {
		return "", serr
	}
	if time.Since(record.CreatedAt) > manifestStateTTL {
		return "", errors.BadRequest("the GitHub App setup link has expired; restart the flow")
	}

	_, integrationID, found := strings.Cut(record.State, ":")
	if !found || integrationID == "" {
		return "", errors.BadRequest("malformed state parameter")
	}

	integration, serr := s.store.GetByID(ctx, integrationID)
	if serr != nil {
		return "", serr
	}
	if integration.Type != models.GitIntegrationTypeGitHubApp {
		return "", errors.BadRequest("state does not reference a GitHub App integration")
	}

	creds, err := s.githubApp.ConvertManifestCode(ctx, code)
	if err != nil {
		return "", errors.BadRequest("failed to convert GitHub App manifest code: %s", err.Error())
	}

	integration.Auth = &models.GitIntegrationAuth{
		GitHubApp: &models.GitHubAppCredentials{
			AppID:         creds.AppID,
			Slug:          creds.Slug,
			PEM:           creds.PEM,
			WebhookSecret: creds.WebhookSecret,
			ClientID:      creds.ClientID,
			ClientSecret:  creds.ClientSecret,
		},
	}
	if serr := s.sealIntegration(integration); serr != nil {
		return "", serr
	}
	// Installation hasn't happened yet; the webhook (or a refresh sync) flips
	// the status to installed.
	integration.Status = models.GitIntegrationStatusPendingInstall
	if _, serr := s.store.Update(ctx, integration); serr != nil {
		return "", serr
	}

	return githubAppInstallURL(creds.Slug), nil
}

func githubAppInstallURL(slug string) string {
	return fmt.Sprintf("%s/apps/%s/installations/new", githubAppsBaseURL, slug)
}

// ListInstallations returns the stored installations; refresh re-lists them
// from GitHub first (fallback for missed webhooks) and syncs status.
func (s *gitIntegrationService) ListInstallations(ctx context.Context, integrationID string, refresh bool) ([]*models.GitInstallation, *errors.ServiceError) {
	integration, serr := s.store.GetByID(ctx, integrationID)
	if serr != nil {
		return nil, serr
	}
	if permErr := s.permissions.Check(ctx, integration.OrganisationID, auth.ResourceGitIntegrations, integrationID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	if integration.Type != models.GitIntegrationTypeGitHubApp {
		return nil, errors.BadRequest("installations only exist for GitHub App integrations")
	}

	if refresh {
		creds, serr := s.appCredentials(integration)
		if serr != nil {
			return nil, serr
		}
		installations, err := s.githubApp.ListInstallations(ctx, creds)
		if err != nil {
			return nil, errors.BadRequest("failed to list installations from GitHub: %s", err.Error())
		}
		if serr := s.installations.DeleteByIntegrationID(ctx, integrationID); serr != nil {
			return nil, serr
		}
		for _, installation := range installations {
			if _, serr := s.installations.Upsert(ctx, &models.GitInstallation{
				GitIntegrationID:    integrationID,
				InstallationID:      installation.ID,
				AccountLogin:        installation.AccountLogin,
				AccountType:         installation.AccountType,
				RepositorySelection: installation.RepositorySelection,
			}); serr != nil {
				return nil, serr
			}
		}
		if serr := s.syncInstallationStatus(ctx, integration); serr != nil {
			return nil, serr
		}
	}

	return s.installations.ListByIntegrationID(ctx, integrationID)
}

// ListRepositories proxies repository discovery through the installation,
// minting a token per call.
func (s *gitIntegrationService) ListRepositories(ctx context.Context, integrationID, query string, page int, installationID int64) (*githubapp.RepoPage, *errors.ServiceError) {
	integration, creds, serr := s.installedApp(ctx, integrationID, auth.ActionRead)
	if serr != nil {
		return nil, serr
	}

	if installationID == 0 {
		installations, serr := s.installations.ListByIntegrationID(ctx, integration.ID)
		if serr != nil {
			return nil, serr
		}
		if len(installations) == 0 {
			return nil, errors.BadRequest("the GitHub App has no installations yet")
		}
		installationID = installations[0].InstallationID
	}

	pageResult, err := s.githubApp.ListInstallationRepos(ctx, creds, installationID, query, page)
	if err != nil {
		return nil, errors.BadRequest("failed to list repositories: %s", err.Error())
	}
	return pageResult, nil
}

func (s *gitIntegrationService) GetRepository(ctx context.Context, integrationID, owner, repo string) (*githubapp.Repo, *errors.ServiceError) {
	integration, creds, serr := s.installedApp(ctx, integrationID, auth.ActionRead)
	if serr != nil {
		return nil, serr
	}
	installation, serr := s.installations.GetByIntegrationAndAccount(ctx, integration.ID, owner)
	if serr != nil {
		return nil, serr
	}
	repository, err := s.githubApp.GetRepo(ctx, creds, installation.InstallationID, owner, repo)
	if err != nil {
		return nil, errors.BadRequest("failed to get repository: %s", err.Error())
	}
	return repository, nil
}

func (s *gitIntegrationService) ListRepositoryBranches(ctx context.Context, integrationID, owner, repo string) ([]string, *errors.ServiceError) {
	integration, creds, serr := s.installedApp(ctx, integrationID, auth.ActionRead)
	if serr != nil {
		return nil, serr
	}
	installation, serr := s.installations.GetByIntegrationAndAccount(ctx, integration.ID, owner)
	if serr != nil {
		return nil, serr
	}
	branches, err := s.githubApp.ListBranches(ctx, creds, installation.InstallationID, owner, repo)
	if err != nil {
		return nil, errors.BadRequest("failed to list branches: %s", err.Error())
	}
	return branches, nil
}

// githubWebhookPayload is the subset of GitHub webhook payloads we act on.
type githubWebhookPayload struct {
	Action       string `json:"action"`
	Installation struct {
		ID    int64 `json:"id"`
		AppID int64 `json:"app_id"`

		Account struct {
			Login string `json:"login"`
			Type  string `json:"type"`
		} `json:"account"`
		RepositorySelection string `json:"repository_selection"`
	} `json:"installation"`
}

// ProcessGitHubWebhook handles installation lifecycle deliveries. The
// integration is located by the app ID in the payload and the delivery is
// HMAC-verified against that integration's webhook secret before any state
// changes. Unmatched or uninteresting events are dropped.
func (s *gitIntegrationService) ProcessGitHubWebhook(ctx context.Context, event string, payload []byte, signature string) *errors.ServiceError {
	if event != GitHubEventInstallation {
		// push (and everything else) is accepted and dropped for now.
		return nil
	}

	var parsed githubWebhookPayload
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return errors.BadRequest("malformed webhook payload: %s", err.Error())
	}
	if parsed.Installation.AppID == 0 {
		return errors.BadRequest("webhook payload has no installation app id")
	}

	integration, creds, serr := s.findAppByID(ctx, parsed.Installation.AppID)
	if serr != nil {
		return serr
	}
	if !verifyWebhookSignature(payload, signature, creds.WebhookSecret) {
		return errors.Forbidden("webhook signature verification failed")
	}

	switch parsed.Action {
	case githubInstallationActCreated:
		if _, serr := s.installations.Upsert(ctx, &models.GitInstallation{
			GitIntegrationID:    integration.ID,
			InstallationID:      parsed.Installation.ID,
			AccountLogin:        parsed.Installation.Account.Login,
			AccountType:         parsed.Installation.Account.Type,
			RepositorySelection: parsed.Installation.RepositorySelection,
		}); serr != nil {
			return serr
		}
	case githubInstallationActDeleted:
		if serr := s.installations.DeleteByInstallationID(ctx, integration.ID, parsed.Installation.ID); serr != nil {
			return serr
		}
	default:
		return nil
	}

	return s.syncInstallationStatus(ctx, integration)
}

func verifyWebhookSignature(payload []byte, signature, secret string) bool {
	if secret == "" || !strings.HasPrefix(signature, githubWebhookSignaturePref) {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(strings.TrimPrefix(signature, githubWebhookSignaturePref)))
}

// InternalMintForRepo mints an installation token for the org's installed
// GitHub App when an installation covers the repository owner. Returns 404
// when no app/installation applies so resolution can fall through.
func (s *gitIntegrationService) InternalMintForRepo(ctx context.Context, organisationID, repoURL string) (*models.GitHubAppMintResult, *errors.ServiceError) {
	integration, serr := s.store.GetGitHubAppForOrg(ctx, organisationID)
	if serr != nil {
		return nil, serr
	}
	if integration.Status != models.GitIntegrationStatusInstalled {
		return nil, errors.NotFound("the organisation's GitHub App is not installed")
	}
	owner, _, ok := gitclient.ParseGitHubRepoURL(repoURL)
	if !ok {
		return nil, errors.NotFound("'%s' is not a GitHub repository", repoURL)
	}
	installation, serr := s.installations.GetByIntegrationAndAccount(ctx, integration.ID, owner)
	if serr != nil {
		return nil, serr
	}
	return s.mintForInstallation(ctx, integration, installation.InstallationID)
}

// InternalMintCloneToken re-mints a token for a known integration and repo;
// used by the hub-side token refresher.
func (s *gitIntegrationService) InternalMintCloneToken(ctx context.Context, integrationID, repoURL string) (*models.GitHubAppMintResult, *errors.ServiceError) {
	integration, serr := s.store.GetByID(ctx, integrationID)
	if serr != nil {
		return nil, serr
	}
	if integration.Type != models.GitIntegrationTypeGitHubApp {
		return nil, errors.BadRequest("integration '%s' is not a GitHub App", integrationID)
	}
	owner, _, ok := gitclient.ParseGitHubRepoURL(repoURL)
	if !ok {
		return nil, errors.BadRequest("'%s' is not a GitHub repository", repoURL)
	}
	installation, serr := s.installations.GetByIntegrationAndAccount(ctx, integration.ID, owner)
	if serr != nil {
		return nil, serr
	}
	return s.mintForInstallation(ctx, integration, installation.InstallationID)
}

func (s *gitIntegrationService) mintForInstallation(ctx context.Context, integration *models.GitIntegration, installationID int64) (*models.GitHubAppMintResult, *errors.ServiceError) {
	creds, serr := s.appCredentials(integration)
	if serr != nil {
		return nil, serr
	}
	token, err := s.githubApp.MintInstallationToken(ctx, creds, installationID)
	if err != nil {
		return nil, errors.GeneralError("failed to mint GitHub App installation token: %s", err.Error())
	}
	return &models.GitHubAppMintResult{
		IntegrationID: integration.ID,
		Host:          integration.Host,
		Token:         token.Value,
		MintedAt:      time.Now().UTC(),
		ExpiresAt:     token.ExpiresAt,
		DataHash:      integration.DataHash,
	}, nil
}

// installedApp loads the integration, checks permissions, and returns the
// unsealed app credentials.
func (s *gitIntegrationService) installedApp(ctx context.Context, integrationID, action string) (*models.GitIntegration, *githubapp.AppCredentials, *errors.ServiceError) {
	integration, serr := s.store.GetByID(ctx, integrationID)
	if serr != nil {
		return nil, nil, serr
	}
	if permErr := s.permissions.Check(ctx, integration.OrganisationID, auth.ResourceGitIntegrations, integrationID, action); permErr != nil {
		return nil, nil, permErr
	}
	if integration.Type != models.GitIntegrationTypeGitHubApp {
		return nil, nil, errors.BadRequest("integration '%s' is not a GitHub App", integrationID)
	}
	creds, serr := s.appCredentials(integration)
	if serr != nil {
		return nil, nil, serr
	}
	return integration, creds, nil
}

// appCredentials unseals the integration and extracts the app credentials.
func (s *gitIntegrationService) appCredentials(integration *models.GitIntegration) (*githubapp.AppCredentials, *errors.ServiceError) {
	if integration.EncryptedAuth == "" {
		return nil, errors.BadRequest("the GitHub App setup has not been completed yet")
	}
	if integration.Auth == nil {
		if serr := s.unsealIntegration(integration); serr != nil {
			return nil, serr
		}
	}
	app := integration.Auth.GitHubApp
	if app == nil {
		return nil, errors.GeneralError("integration '%s' has no GitHub App credentials", integration.ID)
	}
	return &githubapp.AppCredentials{
		AppID:         app.AppID,
		Slug:          app.Slug,
		PEM:           app.PEM,
		WebhookSecret: app.WebhookSecret,
		ClientID:      app.ClientID,
		ClientSecret:  app.ClientSecret,
	}, nil
}

// findAppByID locates a github_app integration by GitHub app ID.
func (s *gitIntegrationService) findAppByID(ctx context.Context, appID int64) (*models.GitIntegration, *githubapp.AppCredentials, *errors.ServiceError) {
	integrations, serr := s.store.ListGitHubApps(ctx)
	if serr != nil {
		return nil, nil, serr
	}
	for _, integration := range integrations {
		if integration.EncryptedAuth == "" {
			continue
		}
		creds, serr := s.appCredentials(integration)
		if serr != nil {
			continue
		}
		if creds.AppID == appID {
			return integration, creds, nil
		}
	}
	return nil, nil, errors.NotFound("no GitHub App integration matches app id %d", appID)
}

// syncInstallationStatus flips the integration between installed and
// pending_install based on whether any installations exist.
func (s *gitIntegrationService) syncInstallationStatus(ctx context.Context, integration *models.GitIntegration) *errors.ServiceError {
	installations, serr := s.installations.ListByIntegrationID(ctx, integration.ID)
	if serr != nil {
		return serr
	}
	status := models.GitIntegrationStatusPendingInstall
	if len(installations) > 0 {
		status = models.GitIntegrationStatusInstalled
	}
	if integration.Status == status {
		return nil
	}
	integration.Status = status
	integration.Auth = nil
	_, serr = s.store.Update(ctx, integration)
	return serr
}

// attachInstallURL populates the transient install URL on github_app rows so
// clients can add more accounts.
func (s *gitIntegrationService) attachInstallURL(integration *models.GitIntegration) {
	if integration.Type != models.GitIntegrationTypeGitHubApp || integration.EncryptedAuth == "" {
		return
	}
	creds, serr := s.appCredentials(integration)
	if serr != nil {
		return
	}
	integration.InstallURL = githubAppInstallURL(creds.Slug)
	integration.Auth = nil
}
