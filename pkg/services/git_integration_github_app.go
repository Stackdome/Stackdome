package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Stackdome/stackdome/pkg/auth"
	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/clients/githubapp"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/google/go-github/v88/github"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

const (
	// manifestStateTTL bounds how long a manifest-flow state stays valid.
	manifestStateTTL = 15 * time.Minute
	// githubHost is the host github_app integrations cover.
	githubHost = "github.com"

	githubAppsBaseURL    = "https://github.com"
	manifestCallbackPath = "/api/v1/git-integrations/github/manifest/callback"
	githubWebhookPath    = "/api/v1/webhooks/github"

	// GitHub webhook event and action names.
	GitHubEventInstallation        = "installation"
	GitHubEventPush                = "push"
	githubInstallationActCreated   = "created"
	githubInstallationActDeleted   = "deleted"
	githubInstallationActSuspend   = "suspend"
	githubInstallationActUnsuspend = "unsuspend"
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

	org, serr := s.organisations.Get(ctx, organisationID)
	if serr != nil {
		return nil, serr
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
	manifest := githubapp.AppManifest{
		Name:           githubAppName(org.Name, organisationID),
		URL:            hub,
		RedirectURL:    hub + manifestCallbackPath,
		SetupURL:       hub,
		Public:         true,
		HookAttributes: githubapp.AppHookAttributes{URL: hub + githubWebhookPath},
		DefaultPermissions: map[string]string{
			githubapp.PermContents: githubapp.PermLevelRead,
			githubapp.PermMetadata: githubapp.PermLevelRead,
		},
		DefaultEvents: []string{GitHubEventPush, GitHubEventInstallation},
	}
	// The API contract exposes the manifest as a free-form object, so marshal
	// the typed manifest into the generic map the model carries.
	manifestMap, err := manifestToMap(manifest)
	if err != nil {
		return nil, err
	}

	return &models.GitHubAppManifestFlow{
		Manifest:  manifestMap,
		GitHubURL: fmt.Sprintf("%s/settings/apps/new?state=%s", githubAppsBaseURL, state),
		State:     state,
	}, nil
}

func manifestToMap(manifest githubapp.AppManifest) (map[string]any, *errors.ServiceError) {
	raw, err := json.Marshal(manifest)
	if err != nil {
		return nil, errors.GeneralError("failed to encode GitHub App manifest: %s", err.Error())
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, errors.GeneralError("failed to encode GitHub App manifest: %s", err.Error())
	}
	return out, nil
}

// githubAppName derives the default GitHub App name from the org name (the
// user can edit it on GitHub's create page). GitHub App names are globally
// unique and limited to ~34 chars of [a-z0-9-], so the org name is slugified
// and the org id is a fallback when the name has no usable characters. In a
// future shared-app (SaaS) mode this per-org naming is unused — one platform
// app is named once.
func githubAppName(orgName, organisationID string) string {
	const prefix = "stackdome-"
	s := slug.Make(orgName)
	if max := 34 - len(prefix); len(s) > max {
		s = strings.Trim(s[:max], "-")
	}
	if s == "" {
		s = strings.ToLower(organisationID)
		if len(s) > 8 {
			s = s[:8]
		}
	}
	return prefix + s
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
				AccountType:         models.GitAccountType(installation.AccountType),
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

// ProcessGitHubWebhook handles installation lifecycle deliveries. The event is
// parsed with go-github, the integration is located by the app ID in the
// payload, and the delivery is HMAC-verified against that integration's webhook
// secret before any state changes. Unmatched or uninteresting events are
// dropped. suspend/unsuspend are treated as uninstall/reinstall so a suspended
// installation is never used to mint tokens.
func (s *gitIntegrationService) ProcessGitHubWebhook(ctx context.Context, event string, payload []byte, signature string) *errors.ServiceError {
	if event != GitHubEventInstallation {
		// push (and everything else) is accepted and dropped for now.
		return nil
	}

	parsed, err := github.ParseWebHook(event, payload)
	if err != nil {
		return errors.BadRequest("malformed webhook payload: %s", err.Error())
	}
	installEvent, ok := parsed.(*github.InstallationEvent)
	if !ok {
		return nil
	}
	install := installEvent.GetInstallation()
	if install.GetAppID() == 0 {
		return errors.BadRequest("webhook payload has no installation app id")
	}

	integration, creds, serr := s.findAppByID(ctx, install.GetAppID())
	if serr != nil {
		return serr
	}
	if err := github.ValidateSignature(signature, payload, []byte(creds.WebhookSecret)); err != nil {
		return errors.Forbidden("webhook signature verification failed")
	}

	switch installEvent.GetAction() {
	case githubInstallationActCreated, githubInstallationActUnsuspend:
		if _, serr := s.installations.Upsert(ctx, &models.GitInstallation{
			GitIntegrationID:    integration.ID,
			InstallationID:      install.GetID(),
			AccountLogin:        install.GetAccount().GetLogin(),
			AccountType:         models.GitAccountType(install.GetAccount().GetType()),
			RepositorySelection: install.GetRepositorySelection(),
		}); serr != nil {
			return serr
		}
	case githubInstallationActDeleted, githubInstallationActSuspend:
		if serr := s.installations.DeleteByInstallationID(ctx, integration.ID, install.GetID()); serr != nil {
			return serr
		}
	default:
		return nil
	}

	return s.syncInstallationStatus(ctx, integration)
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
