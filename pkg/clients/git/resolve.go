package git

import (
	"context"
	"fmt"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/go-git/go-git/v5/plumbing/transport"
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

// RepoHost extracts the host (without port) from any git repository URL —
// https, ssh://, or scp-style (git@host:org/repo) — falling back to the raw
// URL when it cannot be parsed.
func RepoHost(repoURL string) string {
	ep, err := transport.NewEndpoint(repoURL)
	if err != nil || ep.Host == "" {
		return repoURL
	}
	// transport.Endpoint keeps the port separately, but strip defensively in
	// case a raw host:port slips through.
	host, _, found := strings.Cut(ep.Host, ":")
	if found {
		return host
	}
	return ep.Host
}
