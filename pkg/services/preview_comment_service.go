//go:generate mockgen -source=preview_comment_service.go -destination=preview_comment_service_mock.go -package=services
package services

import (
	"context"
	stderrors "errors"
	"fmt"
	"strconv"
	"strings"

	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/clients/githubapp"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
)

const previewCommentMarker = "<!-- stackdome-preview -->"

// PreviewCommentService maintains the sticky PR comment for a preview stack.
// Perm-less: reached only from the preview worker, never from user requests.
type PreviewCommentService interface {
	// InternalUpsertComment renders the comment for the preview's current
	// state and creates or edits the sticky PR comment. Best-effort: callers
	// log the returned error and never fail reconciliation on it.
	InternalUpsertComment(ctx context.Context, preview *models.PreviewStack) error
}

type PreviewCommentServiceSpec struct {
	PreviewStackStore stores.PreviewStackStore
	ConfigStore       stores.StackPreviewConfigStore
	GitIntegrations   GitIntegrationService
	Commenter         githubapp.PullRequestCommenter
	Logger            logger.Logger
}

type previewCommentService struct {
	previews        stores.PreviewStackStore
	configs         stores.StackPreviewConfigStore
	gitIntegrations GitIntegrationService
	commenter       githubapp.PullRequestCommenter
	logger          logger.Logger
}

func NewPreviewCommentService(spec PreviewCommentServiceSpec) PreviewCommentService {
	return &previewCommentService{
		previews:        spec.PreviewStackStore,
		configs:         spec.ConfigStore,
		gitIntegrations: spec.GitIntegrations,
		commenter:       spec.Commenter,
		logger:          spec.Logger,
	}
}

func (s *previewCommentService) InternalUpsertComment(ctx context.Context, preview *models.PreviewStack) error {
	if preview.DeletionTimestamp != nil && preview.GitHubCommentID == 0 {
		return nil // never announced; nothing to tear down
	}

	config, sErr := s.configs.GetByID(ctx, preview.StackPreviewConfigID)
	if sErr != nil {
		return fmt.Errorf("failed to load preview config: %w", sErr)
	}

	owner, repo, ok := gitclient.ParseGitHubRepoURL(config.GitRepository.RepoURL)
	if !ok {
		return nil // not a GitHub repo; comments don't apply
	}

	prNumber, err := strconv.Atoi(preview.PRNumber)
	if err != nil {
		s.logger.Warn(ctx, "preview comment: non-numeric PR number '%s' for preview %s", preview.PRNumber, preview.ID)
		return nil
	}

	mint, sErr := s.gitIntegrations.InternalMintForRepo(ctx, config.OrganisationID, config.GitRepository.RepoURL)
	if sErr != nil {
		if sErr.Is404() {
			return nil // no App installed for this repo
		}
		return fmt.Errorf("failed to mint installation token: %w", sErr)
	}

	body := renderPreviewComment(preview)

	if preview.GitHubCommentID != 0 {
		err := s.commenter.EditComment(ctx, mint.Token, owner, repo, preview.GitHubCommentID, body)
		if err == nil {
			return nil
		}
		if !stderrors.Is(err, githubapp.ErrCommentNotFound) {
			return err
		}
		// Comment was deleted by a human — fall through and re-create.
	}

	id, err := s.commenter.CreateComment(ctx, mint.Token, owner, repo, prNumber, body)
	if err != nil {
		return err
	}
	preview.GitHubCommentID = id
	if _, sErr := s.previews.Update(ctx, preview); sErr != nil {
		return fmt.Errorf("failed to persist comment id: %w", sErr)
	}
	return nil
}

// renderPreviewComment produces the full sticky-comment markdown for the
// preview's current state: deleted > failed > live.
func renderPreviewComment(preview *models.PreviewStack) string {
	var b strings.Builder

	switch {
	case preview.DeletionTimestamp != nil:
		b.WriteString("### 🗑️ Preview environment deleted\n")
	case preview.Status.Phase == models.PreviewStackPhaseFailed:
		b.WriteString("### 🔴 Preview deploy failed\n")
		if preview.Status.Message != "" {
			fmt.Fprintf(&b, "\n%s\n", preview.Status.Message)
		}
		writeOutputs(&b, preview)
	default:
		b.WriteString("### 🟢 Preview environment is live\n")
		writeOutputs(&b, preview)
	}

	b.WriteString("\n" + previewCommentMarker + "\n")
	return b.String()
}

func writeOutputs(b *strings.Builder, preview *models.PreviewStack) {
	outputs := preview.Status.Outputs
	if outputs == nil {
		return
	}
	if outputs.CommitSHA != "" {
		fmt.Fprintf(b, "\nDeployed commit: `%s`\n", outputs.CommitSHA)
	}
	if len(outputs.URLs) == 0 {
		return
	}
	b.WriteString("\n| Resource | URL |\n| --- | --- |\n")
	for _, u := range outputs.URLs {
		fmt.Fprintf(b, "| %s | %s |\n", u.Resource, u.URL)
	}
}
