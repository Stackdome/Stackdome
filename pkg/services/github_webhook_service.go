//go:generate mockgen -destination=../mocks/mock_github_webhook_service.go -package=mocks github.com/Stackdome/stackdome/pkg/services GitHubWebhookService
package services

import (
	"context"
	"strconv"

	"github.com/Stackdome/stackdome/pkg/clients/githubapp"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/google/go-github/v88/github"
)

const (
	githubInstallationActCreated   = "created"
	githubInstallationActDeleted   = "deleted"
	githubInstallationActSuspend   = "suspend"
	githubInstallationActUnsuspend = "unsuspend"

	// maxWebhookLogValueLen bounds unverified webhook values in log lines.
	maxWebhookLogValueLen = 64
)

// truncateForLog bounds a webhook-supplied value before it reaches a log line
// emitted ahead of signature verification.
func truncateForLog(s string) string {
	if len(s) > maxWebhookLogValueLen {
		return s[:maxWebhookLogValueLen] + "..."
	}
	return s
}

// GitHubWebhookService is the ingress for GitHub webhook deliveries. It is
// deliberately separate from GitIntegrationService so webhook dispatch can
// depend on the preview webhook router without a construction cycle.
type GitHubWebhookService interface {
	// ProcessGitHubWebhook resolves the integration a delivery belongs to,
	// HMAC-verifies the payload against that integration's webhook secret,
	// and dispatches installation and pull_request events.
	ProcessGitHubWebhook(ctx context.Context, event string, payload []byte, signature string) *errors.ServiceError
}

type GitHubWebhookServiceSpec struct {
	Store             stores.GitIntegrationStore
	InstallationStore stores.GitInstallationStore
	EncryptionService EncryptionService
	PreviewWebhook    PreviewWebhookService
	Logger            logger.Logger
}

type gitHubWebhookService struct {
	store          stores.GitIntegrationStore
	installations  stores.GitInstallationStore
	encryption     EncryptionService
	previewWebhook PreviewWebhookService
	logger         logger.Logger
}

func NewGitHubWebhookService(spec GitHubWebhookServiceSpec) GitHubWebhookService {
	return &gitHubWebhookService{
		store:          spec.Store,
		installations:  spec.InstallationStore,
		encryption:     spec.EncryptionService,
		previewWebhook: spec.PreviewWebhook,
		logger:         spec.Logger,
	}
}

// ProcessGitHubWebhook handles installation lifecycle and pull_request
// deliveries. The event is parsed with go-github, the integration is resolved
// from the payload, and the delivery is HMAC-verified against that
// integration's webhook secret before any state changes. Unmatched or
// uninteresting events are dropped. suspend/unsuspend are treated as
// uninstall/reinstall so a suspended installation is never used to mint
// tokens.
func (s *gitHubWebhookService) ProcessGitHubWebhook(ctx context.Context, event string, payload []byte, signature string) *errors.ServiceError {
	if event != GitHubEventInstallation && event != GitHubEventPullRequest {
		// push (and everything else) is accepted and dropped for now.
		s.logger.Info(ctx, "github webhook: ignoring '%s' event", truncateForLog(event))
		return nil
	}

	parsed, err := github.ParseWebHook(event, payload)
	if err != nil {
		s.logger.Warn(ctx, "github webhook: malformed '%s' payload: %s", event, truncateForLog(err.Error()))
		return errors.BadRequest("malformed webhook payload: %s", err.Error())
	}

	integration, creds, serr := s.webhookIntegration(ctx, parsed)
	if serr != nil {
		s.logger.Warn(ctx, "github webhook: cannot resolve integration for '%s' event: %s", event, serr.Reason)
		return serr
	}
	if err := github.ValidateSignature(signature, payload, []byte(creds.WebhookSecret)); err != nil {
		s.logger.WithField(logger.FieldOrgID, integration.OrganisationID).Warn(ctx, "github webhook: signature verification failed ('%s' event)", event)
		return errors.Forbidden("webhook signature verification failed")
	}
	s.logger.WithField(logger.FieldOrgID, integration.OrganisationID).Info(ctx, "github webhook: verified '%s' delivery", event)

	switch e := parsed.(type) {
	case *github.InstallationEvent:
		return s.handleInstallationEvent(ctx, integration, e)
	case *github.PullRequestEvent:
		return s.handlePullRequestEvent(ctx, integration, e)
	default:
		return nil
	}
}

// webhookIntegration resolves the integration a delivery belongs to. Only
// installation events carry the full installation object with app_id;
// repository events (pull_request) carry the slim {id, node_id} reference, so
// they resolve through the stored installation row instead.
func (s *gitHubWebhookService) webhookIntegration(ctx context.Context, parsed any) (*models.GitIntegration, *githubapp.AppCredentials, *errors.ServiceError) {
	switch e := parsed.(type) {
	case *github.InstallationEvent:
		appID := e.GetInstallation().GetAppID()
		if appID == 0 {
			return nil, nil, errors.BadRequest("installation payload has no app id")
		}
		return s.findAppByID(ctx, appID)
	case *github.PullRequestEvent:
		installationID := e.GetInstallation().GetID()
		if installationID == 0 {
			return nil, nil, errors.BadRequest("webhook payload has no installation id")
		}
		return s.findAppByInstallation(ctx, installationID)
	default:
		return nil, nil, errors.BadRequest("unsupported webhook event")
	}
}

// findAppByID locates a github_app integration by GitHub app ID.
func (s *gitHubWebhookService) findAppByID(ctx context.Context, appID int64) (*models.GitIntegration, *githubapp.AppCredentials, *errors.ServiceError) {
	integrations, serr := s.store.ListGitHubApps(ctx)
	if serr != nil {
		return nil, nil, serr
	}
	for _, integration := range integrations {
		if integration.EncryptedAuth == "" {
			continue
		}
		creds, serr := unsealAppCredentials(s.encryption, integration)
		if serr != nil {
			continue
		}
		if creds.AppID == appID {
			return integration, creds, nil
		}
	}
	return nil, nil, errors.NotFound("no GitHub App integration matches app id %d", appID)
}

// findAppByInstallation resolves an integration from a GitHub-global
// installation id via the stored installation row.
func (s *gitHubWebhookService) findAppByInstallation(ctx context.Context, installationID int64) (*models.GitIntegration, *githubapp.AppCredentials, *errors.ServiceError) {
	install, serr := s.installations.GetByInstallationID(ctx, installationID)
	if serr != nil {
		return nil, nil, serr
	}
	integration, serr := s.store.GetByID(ctx, install.GitIntegrationID)
	if serr != nil {
		return nil, nil, serr
	}
	creds, serr := unsealAppCredentials(s.encryption, integration)
	if serr != nil {
		return nil, nil, serr
	}
	return integration, creds, nil
}

func (s *gitHubWebhookService) handleInstallationEvent(ctx context.Context, integration *models.GitIntegration, installEvent *github.InstallationEvent) *errors.ServiceError {
	install := installEvent.GetInstallation()
	s.logger.WithField(logger.FieldOrgID, integration.OrganisationID).Info(ctx, "github webhook: installation '%s' by '%s'", installEvent.GetAction(), install.GetAccount().GetLogin())
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

	return syncInstallationStatus(ctx, s.store, s.installations, integration)
}

func (s *gitHubWebhookService) handlePullRequestEvent(ctx context.Context, integration *models.GitIntegration, prEvent *github.PullRequestEvent) *errors.ServiceError {
	pr := prEvent.GetPullRequest()
	// fork PRs carry untrusted code; never build them
	if pr.GetHead().GetRepo().GetID() != prEvent.GetRepo().GetID() {
		s.logger.Info(ctx, "github webhook: dropping fork PR #%d on '%s'", prEvent.GetNumber(), prEvent.GetRepo().GetCloneURL())
		return nil
	}
	ev := PullRequestEvent{
		OrganisationID: integration.OrganisationID,
		RepoURL:        prEvent.GetRepo().GetCloneURL(),
		Branch:         pr.GetHead().GetRef(),
		BaseBranch:     pr.GetBase().GetRef(),
		HeadSHA:        pr.GetHead().GetSHA(),
		PRNumber:       strconv.Itoa(prEvent.GetNumber()),
		Action:         PullRequestAction(prEvent.GetAction()),
	}
	return s.previewWebhook.HandlePullRequest(ctx, ev)
}
