package git

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/google/go-github/v50/github"
)

// GitHubClient provides GitHub-specific API operations.
type gitHubClient struct {
	client *github.Client
}

var _ GitClient = (*gitHubClient)(nil)

func newGitHubClientWithToken(token string) *gitHubClient {
	return &gitHubClient{client: github.NewTokenClient(context.Background(), token)}
}

func newGitHubClientWithBasicAuth(username, password string) *gitHubClient {
	transport := &github.BasicAuthTransport{
		Username: username,
		Password: password,
	}
	return &gitHubClient{client: github.NewClient(transport.Client())}
}

func newGitHubClientAnonymous() *gitHubClient {
	return &gitHubClient{client: github.NewClient(&http.Client{})}
}

func (g *gitHubClient) CheckAccess(ctx context.Context, repoURL string) (bool, error) {
	owner, repo, ok := ParseGitHubRepoURL(repoURL)
	if !ok {
		return false, fmt.Errorf("not a GitHub repository URL: %s", repoURL)
	}
	_, resp, err := g.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		if resp != nil {
			switch resp.StatusCode {
			case http.StatusNotFound:
				return false, fmt.Errorf("repository not found: %v: %w", err, ErrNotFound)
			case http.StatusUnauthorized, http.StatusForbidden:
				return false, fmt.Errorf("authentication failed: %v: %w", err, ErrAuthFailed)
			}
		}
		return false, fmt.Errorf("failed to access git repo: %w", err)
	}
	return true, nil
}

func (g *gitHubClient) GetTagSHA(ctx context.Context, repoURL, tag string) (string, error) {
	owner, repo, ok := ParseGitHubRepoURL(repoURL)
	if !ok {
		return "", fmt.Errorf("not a GitHub repository URL: %s", repoURL)
	}

	sha, resp, err := g.client.Repositories.GetCommitSHA1(ctx, owner, repo, tag, "")
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return "", fmt.Errorf("tag '%s' not found in repository", tag)
		}
		return "", fmt.Errorf("failed to get tag SHA: %w", err)
	}
	return sha, nil
}

func (g *gitHubClient) CheckTagExists(ctx context.Context, repoURL, tag string) (bool, error) {
	owner, repo, ok := ParseGitHubRepoURL(repoURL)
	if !ok {
		return false, fmt.Errorf("not a GitHub repository URL: %s", repoURL)
	}

	_, resp, err := g.client.Git.GetRef(ctx, owner, repo, "tags/"+tag)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check tag existence: %w", err)
	}
	return true, nil
}

func (g *gitHubClient) GetBranchHeadSHA(ctx context.Context, repoURL, branch string) (*RepoResult, error) {
	owner, repo, ok := ParseGitHubRepoURL(repoURL)
	if !ok {
		return nil, fmt.Errorf("not a GitHub repository URL: %s", repoURL)
	}

	ref, resp, err := g.client.Repositories.GetBranch(ctx, owner, repo, branch, true)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("branch '%s' not found in repository: %w", branch, ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get branch head: %w", err)
	}
	if ref.Commit == nil || ref.Commit.SHA == nil {
		return nil, fmt.Errorf("branch '%s' has no commit SHA", branch)
	}

	return &RepoResult{
		HeadSHA: *ref.Commit.SHA,
		Branch:  branch,
	}, nil
}

func (g *gitHubClient) FetchFile(ctx context.Context, repoURL, branch, filePath string) ([]byte, error) {
	owner, repo, ok := ParseGitHubRepoURL(repoURL)
	if !ok {
		return nil, fmt.Errorf("not a GitHub repository URL: %s", repoURL)
	}

	fc, _, resp, err := g.client.Repositories.GetContents(ctx, owner, repo, filePath, &github.RepositoryContentGetOptions{
		Ref: branch,
	})
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("file '%s' not found in branch '%s': %w", filePath, branch, ErrNotFound)
		}
		return nil, fmt.Errorf("failed to fetch file: %w", err)
	}
	if fc == nil {
		return nil, fmt.Errorf("path '%s' is a directory, not a file", filePath)
	}

	content, err := fc.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode file content: %w", err)
	}
	return []byte(content), nil
}

func (g *gitHubClient) CheckFileExists(ctx context.Context, owner, repo, filePath, ref string) (bool, error) {
	_, _, resp, err := g.client.Repositories.GetContents(ctx, owner, repo, filePath, &github.RepositoryContentGetOptions{
		Ref: ref,
	})
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}
	return true, nil
}

func IsGithubHost(host string) bool {
	return host == "github.com" || host == "www.github.com"
}

// ParseGitHubRepoURL extracts owner and repo from a GitHub URL.
// Handles HTTPS, SSH, and SCP-style URLs (git@github.com:owner/repo.git).
// Returns false if the URL is not a GitHub repo.
func ParseGitHubRepoURL(repoURL string) (owner, repo string, ok bool) {
	ep, err := transport.NewEndpoint(repoURL)
	if err != nil || !IsGithubHost(ep.Host) {
		return "", "", false
	}
	path := strings.TrimSuffix(strings.Trim(ep.Path, "/"), ".git")
	owner, repo, ok = strings.Cut(path, "/")
	return owner, repo, ok && owner != "" && repo != ""
}

// IsGithubRepo returns true for gh repos
func IsGithubRepo(repoURL string) bool {
	_, _, ok := ParseGitHubRepoURL(repoURL)
	return ok
}
