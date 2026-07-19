package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
)

// PullRequestAction is the GitHub pull_request webhook action that triggered
// a PullRequestEvent.
type PullRequestAction string

const (
	PRActionOpened      PullRequestAction = "opened"
	PRActionReopened    PullRequestAction = "reopened"
	PRActionSynchronize PullRequestAction = "synchronize"
	PRActionClosed      PullRequestAction = "closed"
)

// PullRequestEvent carries the fields of a GitHub pull_request webhook
// payload that are needed to route preview stack provisioning.
type PullRequestEvent struct {
	OrganisationID string
	RepoURL        string
	Branch         string // PR head ref
	BaseBranch     string // PR base ref
	HeadSHA        string
	PRNumber       string
	Action         PullRequestAction
}

//go:generate mockgen -destination=preview_webhook_service_mock.go -package=services -self_package=github.com/Stackdome/stackdome/pkg/services github.com/Stackdome/stackdome/pkg/services PreviewWebhookService
type PreviewWebhookService interface {
	HandlePullRequest(ctx context.Context, ev PullRequestEvent) *errors.ServiceError
}

type previewWebhookService struct {
	configs  stores.StackPreviewConfigStore
	previews stores.PreviewStackStore
	service  PreviewStackService
	logger   logger.Logger
}

type PreviewWebhookServiceSpec struct {
	ConfigStore       stores.StackPreviewConfigStore
	PreviewStackStore stores.PreviewStackStore
	PreviewService    PreviewStackService
	Logger            logger.Logger
}

func NewPreviewWebhookService(spec PreviewWebhookServiceSpec) PreviewWebhookService {
	return &previewWebhookService{
		configs:  spec.ConfigStore,
		previews: spec.PreviewStackStore,
		service:  spec.PreviewService,
		logger:   spec.Logger,
	}
}

// Per-delivery outcome labels for the webhook log line.
const (
	prOutcomeCreated      = "created preview"
	prOutcomeSynced       = "synced preview"
	prOutcomeDeleted      = "deleting preview"
	prOutcomeNoConfig     = "dropped: no preview config for repo"
	prOutcomeBaseMismatch = "dropped: PR base branch does not match config"
	prOutcomeNoPreview    = "dropped: no preview to act on"
	prOutcomeUnhandled    = "dropped: unhandled action"
	prOutcomeCapReached   = "skipped: active preview cap reached"
)

// HandlePullRequest routes the PR event and logs exactly one line per
// delivery with the outcome — webhook flows are otherwise invisible when
// they drop events by design.
func (s *previewWebhookService) HandlePullRequest(ctx context.Context, ev PullRequestEvent) *errors.ServiceError {
	outcome, sErr := s.routePullRequest(ctx, ev)
	if sErr != nil {
		s.logger.Error(ctx, "preview webhook: %s PR #%s on '%s': %s", ev.Action, ev.PRNumber, ev.RepoURL, sErr.Reason)
		return sErr
	}
	s.logger.Info(ctx, "preview webhook: %s PR #%s on '%s': %s", ev.Action, ev.PRNumber, ev.RepoURL, outcome)
	return nil
}

// routePullRequest resolves the preview config for the PR's repo, gates on
// the config's configured base branch, and routes the PR action to the
// matching perm-less preview stack entrypoint. Events for repos/branches with
// no matching config, or with no existing preview to act on, are dropped
// (nil error) rather than treated as failures: the webhook has no user to
// report an error to and most PRs are simply not previewed. A cap-reached
// rejection is also a drop — failing the delivery would make GitHub back off
// (and eventually disable the webhook) over expected capacity behavior.
func (s *previewWebhookService) routePullRequest(ctx context.Context, ev PullRequestEvent) (string, *errors.ServiceError) {
	config, sErr := s.configs.GetByOrgAndRepo(ctx, ev.OrganisationID, models.NormalizeRepoURL(ev.RepoURL))
	if sErr != nil {
		if sErr.Is404() {
			return prOutcomeNoConfig, nil
		}
		return "", sErr
	}

	if config.GitRepository.BaseBranch != ev.BaseBranch {
		return prOutcomeBaseMismatch, nil
	}

	existing, sErr := s.previews.GetByConfigAndPR(ctx, config.ID, ev.PRNumber)
	if sErr != nil {
		if !sErr.Is404() {
			return "", sErr
		}
		existing = nil
	}

	switch ev.Action {
	case PRActionOpened, PRActionReopened:
		if existing != nil {
			return prOutcomeSynced, s.service.InternalSyncFromWebhook(ctx, existing.ID, ev.HeadSHA)
		}
		return s.createPreview(ctx, config, ev)
	case PRActionSynchronize:
		if existing == nil {
			return s.createPreview(ctx, config, ev)
		}
		return prOutcomeSynced, s.service.InternalSyncFromWebhook(ctx, existing.ID, ev.HeadSHA)
	case PRActionClosed:
		if existing == nil {
			return prOutcomeNoPreview, nil
		}
		return prOutcomeDeleted, s.service.InternalDeleteFromWebhook(ctx, existing.ID)
	default:
		return prOutcomeUnhandled, nil
	}
}

func (s *previewWebhookService) createPreview(ctx context.Context, config *models.StackPreviewConfig, ev PullRequestEvent) (string, *errors.ServiceError) {
	_, sErr := s.service.InternalCreateFromWebhook(ctx, config, ev.PRNumber, ev.Branch, ev.HeadSHA)
	if sErr != nil {
		if sErr.Code == errors.ErrorTooManyRequests {
			return prOutcomeCapReached, nil
		}
		return "", sErr
	}
	return prOutcomeCreated, nil
}
