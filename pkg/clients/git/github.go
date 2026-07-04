package git

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/google/go-github/v88/github"
)

// GitHubClient provides GitHub-specific API operations.
type gitHubClient struct {
	client *github.Client
}

var _ GitClient = (*gitHubClient)(nil)

func newGitHubClientWithToken(token string) *gitHubClient {
	// No URL option, so NewClient cannot error.
	client, _ := github.NewClient(github.WithAuthToken(token))
	return &gitHubClient{client: client}
}

func newGitHubClientWithBasicAuth(username, password string) *gitHubClient {
	transport := &github.BasicAuthTransport{
		Username: username,
		Password: password,
	}
	client, _ := github.NewClient(github.WithHTTPClient(transport.Client()))
	return &gitHubClient{client: client}
}

func newGitHubClientAnonymous() *gitHubClient {
	client, _ := github.NewClient()
	return &gitHubClient{client: client}
}

func (g *gitHubClient) CheckAccess(ctx context.Context, repoURL string) (bool, error) {
	owner, repo, ok := ParseGitHubRepoURL(repoURL)
	if !ok {
		return false, fmt.Errorf("not a GitHub repository URL: %s", repoURL)
	}
	_, resp, err := g.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		if resp != nil {
			if isRateLimited(resp) {
				return false, fmt.Errorf("rate limited: %v: %w", err, ErrRateLimited)
			}
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

func (g *gitHubClient) GetDefaultBranch(ctx context.Context, repoURL string) (string, error) {
	owner, repo, ok := ParseGitHubRepoURL(repoURL)
	if !ok {
		return "", fmt.Errorf("not a GitHub repository URL: %s", repoURL)
	}
	repository, resp, err := g.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		if resp != nil {
			if isRateLimited(resp) {
				return "", fmt.Errorf("rate limited: %v: %w", err, ErrRateLimited)
			}
			switch resp.StatusCode {
			case http.StatusNotFound:
				return "", fmt.Errorf("repository not found: %v: %w", err, ErrNotFound)
			case http.StatusUnauthorized, http.StatusForbidden:
				return "", fmt.Errorf("authentication failed: %v: %w", err, ErrAuthFailed)
			}
		}
		return "", fmt.Errorf("failed to get repository: %w", err)
	}
	if repository.GetDefaultBranch() == "" {
		return "", fmt.Errorf("repository has no default branch")
	}
	return repository.GetDefaultBranch(), nil
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

	ref, resp, err := g.client.Repositories.GetBranch(ctx, owner, repo, branch, 1)
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

func (g *gitHubClient) FetchFile(ctx context.Context, repoURL, ref, filePath string) ([]byte, error) {
	owner, repo, ok := ParseGitHubRepoURL(repoURL)
	if !ok {
		return nil, fmt.Errorf("not a GitHub repository URL: %s", repoURL)
	}

	fc, _, resp, err := g.client.Repositories.GetContents(ctx, owner, repo, filePath, &github.RepositoryContentGetOptions{
		Ref: ref,
	})
	if err != nil {
		if resp != nil {
			if isRateLimited(resp) {
				return nil, fmt.Errorf("rate limited: %v: %w", err, ErrRateLimited)
			}
			if resp.StatusCode == http.StatusNotFound {
				return nil, fmt.Errorf("file '%s' not found at ref '%s': %w", filePath, ref, ErrNotFound)
			}
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

func isRateLimited(resp *github.Response) bool {
	if resp == nil {
		return false
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return true
	}
	if resp.StatusCode == http.StatusForbidden && resp.Rate.Remaining == 0 {
		return true
	}
	return false
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
