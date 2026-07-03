package git

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

// ResolveGitRepoRevision ensures the revision has a resolved commit SHA.
// If Commit is already set, it returns as-is. Otherwise it resolves the
// branch HEAD or tag SHA using the provided git client.
func ResolveGitRepoRevision(ctx context.Context, client GitClient, repoURL string, rev models.GitRepoRevision) (models.GitRepoRevision, error) {
	if rev.Commit != "" {
		return rev, nil
	}

	switch rev.Type() {
	case models.Branch:
		if rev.Commit != "" {
			return rev, nil
		}
		result, err := client.GetBranchHeadSHA(ctx, repoURL, rev.Branch)
		if err != nil {
			return rev, fmt.Errorf("resolve branch '%s': %w", rev.Branch, err)
		}
		rev.Commit = result.HeadSHA
		return rev, nil

	case models.Tag:
		sha, err := client.GetTagSHA(ctx, repoURL, rev.Tag)
		if err != nil {
			return rev, fmt.Errorf("resolve tag '%s': %w", rev.Tag, err)
		}
		rev.Commit = sha
		return rev, nil

	default:
		return rev, fmt.Errorf("git revision has no commit, tag, or branch")
	}
}
