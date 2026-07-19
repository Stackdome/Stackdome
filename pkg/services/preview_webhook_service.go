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

// HandlePullRequest resolves the preview config for the PR's repo, gates on
// the config's configured base branch, and routes the PR action to the
// matching perm-less preview stack entrypoint. Events for repos/branches with
// no matching config, or with no existing preview to act on, are dropped
// (nil error) rather than treated as failures: the webhook has no user to
// report an error to and most PRs are simply not previewed.
func (s *previewWebhookService) HandlePullRequest(ctx context.Context, ev PullRequestEvent) *errors.ServiceError {
	config, sErr := s.configs.GetByOrgAndRepo(ctx, ev.OrganisationID, models.NormalizeRepoURL(ev.RepoURL))
	if sErr != nil {
		if sErr.Is404() {
			return nil
		}
		return sErr
	}

	if config.GitRepository.BaseBranch != ev.BaseBranch {
		return nil
	}

	existing, sErr := s.previews.GetByConfigAndPR(ctx, config.ID, ev.PRNumber)
	if sErr != nil {
		if !sErr.Is404() {
			return sErr
		}
		existing = nil
	}

	switch ev.Action {
	case PRActionOpened, PRActionReopened:
		if existing != nil {
			return s.service.InternalSyncFromWebhook(ctx, existing.ID, ev.HeadSHA)
		}
		_, sErr := s.service.InternalCreateFromWebhook(ctx, config, ev.PRNumber, ev.Branch, ev.HeadSHA)
		return s.dropCapReached(ctx, ev, sErr)
	case PRActionSynchronize:
		if existing == nil {
			_, sErr := s.service.InternalCreateFromWebhook(ctx, config, ev.PRNumber, ev.Branch, ev.HeadSHA)
			return s.dropCapReached(ctx, ev, sErr)
		}
		return s.service.InternalSyncFromWebhook(ctx, existing.ID, ev.HeadSHA)
	case PRActionClosed:
		if existing == nil {
			return nil
		}
		return s.service.InternalDeleteFromWebhook(ctx, existing.ID)
	default:
		return nil
	}
}

// dropCapReached converts a max-active-previews rejection into a logged skip:
// failing the delivery would make GitHub back off (and eventually disable the
// webhook) over expected capacity behavior.
func (s *previewWebhookService) dropCapReached(ctx context.Context, ev PullRequestEvent, sErr *errors.ServiceError) *errors.ServiceError {
	if sErr != nil && sErr.Code == errors.ErrorTooManyRequests {
		s.logger.Warn(ctx, "preview webhook: skipping PR #%s on '%s': %s", ev.PRNumber, ev.RepoURL, sErr.Reason)
		return nil
	}
	return sErr
}
