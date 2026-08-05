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
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"golang.org/x/sync/errgroup"
)

const (
	// manifestStateTTL bounds how long a manifest-flow state stays valid.
	manifestStateTTL = 15 * time.Minute
	// githubHost is the host github_app integrations cover.
	githubHost = "github.com"

	githubAppsBaseURL    = "https://github.com"
	manifestCallbackPath = "/api/v1/git-integrations/github/manifest/callback"
	githubWebhookPath    = "/api/v1/webhooks/github"
	// setupCompleteRedirect is where the browser lands after a platform-app
	// install. setup_action tells the SPA it is the install popup, so it can
	// notify its opener and close.
	setupCompleteRedirect = "/git-integrations?setup_action=install"

	// GitHub webhook event names.
	GitHubEventInstallation = "installation"
	GitHubEventPullRequest  = "pull_request"
	GitHubEventPush         = "push"
	// issue_comment and label are subscribed ahead of use so already-created
	// apps receive them when the hub starts handling them; unknown events are
	// dropped by the webhook handler until then. GitHub delivers PR comments
	// under issue_comment.
	GitHubEventIssueComment = "issue_comment"
	GitHubEventLabel        = "label"
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
	}

	// Platform mode allows re-running the flow on an installed integration:
	// that is how a second GitHub account (personal + org) gets added, since
	// the shared-app webhook cannot attribute new installations to an org.
	if s.platformApp != nil {
		return s.platformAppInstallFlow(ctx, integration)
	}
	if integration.Status == models.GitIntegrationStatusInstalled {
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
			githubapp.PermContents:     githubapp.PermLevelRead,
			githubapp.PermMetadata:     githubapp.PermLevelRead,
			githubapp.PermPullRequests: githubapp.PermLevelWrite,
			// Issues read is what allows the issue_comment event below.
			githubapp.PermIssues: githubapp.PermLevelRead,
		},
		// GitHub rejects "installation" as a default_event — installation
		// lifecycle deliveries are sent to every app automatically, so only
		// subscribable events belong here.
		DefaultEvents: []string{GitHubEventPush, GitHubEventPullRequest, GitHubEventIssueComment, GitHubEventLabel},
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

// platformAppInstallFlow is the manifest flow's shortcut when the hub runs one
// platform-wide GitHub App: there is nothing to create on GitHub, so the user
// goes straight to the app's install page. The state comes back on the setup
// callback and is what binds the new installation to this org. The row stores
// no credentials — platform creds resolve from config at read time.
func (s *gitIntegrationService) platformAppInstallFlow(ctx context.Context, integration *models.GitIntegration) (*models.GitHubAppManifestFlow, *errors.ServiceError) {
	state := fmt.Sprintf("%s:%s", uuid.NewString(), integration.ID)
	if serr := s.oauthStates.Create(ctx, &models.OAuthState{
		State:     state,
		Provider:  models.OAuthProviderGitHubAppInstall,
		CreatedAt: time.Now().UTC(),
	}); serr != nil {
		return nil, serr
	}

	return &models.GitHubAppManifestFlow{
		GitHubURL: fmt.Sprintf("%s?state=%s", githubAppInstallURL(s.platformApp.Slug), state),
		State:     state,
	}, nil
}

// HandleGitHubAppSetup finishes the platform-app install: GitHub redirects the
// browser to the app's setup URL with the new installation id and the state
// issued when the flow started. Unauthenticated: the state is the proof of
// initiation, exactly as in the manifest callback.
func (s *gitIntegrationService) HandleGitHubAppSetup(ctx context.Context, installationID int64, state string) (string, *errors.ServiceError) {
	if installationID == 0 || state == "" {
		return "", errors.BadRequest("installation_id and state are required")
	}

	record, serr := s.oauthStates.Consume(ctx, state, models.OAuthProviderGitHubAppInstall)
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

	creds, serr := s.appCredentials(integration)
	if serr != nil {
		return "", serr
	}
	// The id from the query string is not trusted on its own: confirm with
	// GitHub that this installation really belongs to our app.
	in, err := s.githubApp.GetInstallation(ctx, creds, installationID)
	if err != nil {
		return "", errors.NotFound("installation %d was not found on the platform GitHub App: %s", installationID, err.Error())
	}
	if in.Suspended {
		return "", errors.BadRequest("installation %d is suspended on GitHub", installationID)
	}

	if _, serr := s.installations.Upsert(ctx, &models.GitInstallation{
		GitIntegrationID:    integration.ID,
		InstallationID:      in.ID,
		AccountLogin:        in.AccountLogin,
		AccountType:         models.GitAccountType(in.AccountType),
		RepositorySelection: in.RepositorySelection,
	}); serr != nil {
		return "", serr
	}
	if serr := syncInstallationStatus(ctx, s.store, s.installations, integration); serr != nil {
		return "", serr
	}
	return s.externalURL + setupCompleteRedirect, nil
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
		if serr := s.reconcileInstallations(ctx, integration); serr != nil {
			return nil, serr
		}
	}

	return s.installations.ListByIntegrationID(ctx, integrationID)
}

// ListRepositories proxies repository discovery through the installation,
// minting a token per call. When installationUUID is set it lists that one
// installation's page; when empty it aggregates page N across every
// installation of the integration.
func (s *gitIntegrationService) ListRepositories(ctx context.Context, integrationID string, page int, installationUUID string) (*githubapp.RepoPage, *errors.ServiceError) {
	integration, creds, serr := s.installedApp(ctx, integrationID, auth.ActionRead)
	if serr != nil {
		return nil, serr
	}

	if installationUUID != "" {
		installation, serr := s.installations.GetByIntegrationAndID(ctx, integration.ID, installationUUID)
		if serr != nil {
			return nil, serr
		}
		pageResult, err := s.githubApp.ListInstallationRepos(ctx, creds, installation.InstallationID, page)
		if err != nil {
			return nil, errors.BadRequest("failed to list repositories: %s", err.Error())
		}
		return pageResult, nil
	}

	installations, serr := s.installations.ListByIntegrationID(ctx, integration.ID)
	if serr != nil {
		return nil, serr
	}
	return s.aggregateRepos(ctx, creds, installations, page)
}

// aggregateRepos fetches page N from every installation concurrently and merges
// the results in stable installation order.
func (s *gitIntegrationService) aggregateRepos(ctx context.Context, creds *githubapp.AppCredentials, installations []*models.GitInstallation, page int) (*githubapp.RepoPage, *errors.ServiceError) {
	if page <= 0 {
		page = 1
	}
	merged := &githubapp.RepoPage{Page: page, Repos: []githubapp.Repo{}}
	if len(installations) == 0 {
		return merged, nil
	}

	pages := make([]*githubapp.RepoPage, len(installations))
	group, ctx := errgroup.WithContext(ctx)
	for i, installation := range installations {
		i, installation := i, installation
		group.Go(func() error {
			result, err := s.githubApp.ListInstallationRepos(ctx, creds, installation.InstallationID, page)
			if err != nil {
				return err
			}
			pages[i] = result
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, errors.BadRequest("failed to list repositories: %s", err.Error())
	}

	for _, p := range pages {
		merged.Repos = append(merged.Repos, p.Repos...)
		merged.TotalCount += p.TotalCount
		merged.HasNext = merged.HasNext || p.HasNext
	}
	return merged, nil
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
	if integration.Status != models.GitIntegrationStatusInstalled {
		return nil, nil, errors.BadRequest("the GitHub App setup has not been completed yet")
	}
	creds, serr := s.appCredentials(integration)
	if serr != nil {
		return nil, nil, serr
	}
	return integration, creds, nil
}

// appCredentials unseals the integration and extracts the app credentials.
func (s *gitIntegrationService) appCredentials(integration *models.GitIntegration) (*githubapp.AppCredentials, *errors.ServiceError) {
	return unsealAppCredentials(s.encryptionService, s.platformApp, integration)
}

// unsealAppCredentials returns the GitHub App credentials for an integration.
// Shared by the integration service and the webhook ingress service.
//   - Sealed auth on the row (BYO app): decrypt and use it.
//   - No sealed auth + platform app configured: the platform credentials, read
//     from config every time so a key rotation is just an env change.
func unsealAppCredentials(enc EncryptionService, platform *githubapp.AppCredentials, integration *models.GitIntegration) (*githubapp.AppCredentials, *errors.ServiceError) {
	if integration.EncryptedAuth == "" {
		if platform != nil {
			return platform, nil
		}
		return nil, errors.BadRequest("the GitHub App setup has not been completed yet")
	}
	if integration.Auth == nil {
		if serr := unsealGitIntegration(enc, integration); serr != nil {
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

// reconcileInstallations re-lists installations from GitHub (the webhook-miss
// fallback) and reconciles the local table to match, atomically: live
// installations are upserted in place, suspended or removed ones are deleted,
// and the integration status is synced. The GitHub call runs outside the
// transaction so a slow API response never holds the DB transaction open.
func (s *gitIntegrationService) reconcileInstallations(ctx context.Context, integration *models.GitIntegration) *errors.ServiceError {
	creds, serr := s.appCredentials(integration)
	if serr != nil {
		return serr
	}
	remote, err := s.githubApp.ListInstallations(ctx, creds)
	if err != nil {
		return errors.BadRequest("failed to list installations from GitHub: %s", err.Error())
	}

	// A platform-backed integration (no sealed auth of its own) shares the app
	// with every org, so GitHub's list spans all of them. Only installations
	// the setup callback already bound to this integration are its to
	// reconcile. BYO apps own their whole list.
	sharedApp := s.platformApp != nil && integration.EncryptedAuth == ""
	owned := map[int64]bool{}
	if sharedApp {
		local, serr := s.installations.ListByIntegrationID(ctx, integration.ID)
		if serr != nil {
			return serr
		}
		for _, l := range local {
			owned[l.InstallationID] = true
		}
	}

	return s.atomic.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		live := make(map[int64]bool, len(remote))
		for _, in := range remote {
			if sharedApp && !owned[in.ID] {
				continue
			}
			if in.Suspended {
				// Suspended installations can't mint tokens; treat as absent so
				// they are pruned below, matching the webhook suspend handling.
				continue
			}
			live[in.ID] = true
			if _, serr := s.installations.Upsert(ctx, &models.GitInstallation{
				GitIntegrationID:    integration.ID,
				InstallationID:      in.ID,
				AccountLogin:        in.AccountLogin,
				AccountType:         models.GitAccountType(in.AccountType),
				RepositorySelection: in.RepositorySelection,
			}); serr != nil {
				return serr
			}
		}

		local, serr := s.installations.ListByIntegrationID(ctx, integration.ID)
		if serr != nil {
			return serr
		}
		for _, l := range local {
			if !live[l.InstallationID] {
				if serr := s.installations.DeleteByInstallationID(ctx, integration.ID, l.InstallationID); serr != nil {
					return serr
				}
			}
		}

		return syncInstallationStatus(ctx, s.store, s.installations, integration)
	})
}

// syncInstallationStatus flips the integration between installed and
// pending_install based on whether any installations exist. Shared by the
// integration service and the webhook ingress service.
func syncInstallationStatus(ctx context.Context, store stores.GitIntegrationStore, installationStore stores.GitInstallationStore, integration *models.GitIntegration) *errors.ServiceError {
	installations, serr := installationStore.ListByIntegrationID(ctx, integration.ID)
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
	_, serr = store.Update(ctx, integration)
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
