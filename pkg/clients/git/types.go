package git

import (
	"context"
	"fmt"
)

//go:generate mockgen -destination=../../mocks/mock_git_client.go -package=mocks github.com/Stackdome/stackdome/pkg/clients/git GitClient
type GitClient interface {
	CheckAccess(ctx context.Context, repoURL string) (bool, error)
	GetTagSHA(ctx context.Context, repoURL, tag string) (string, error)
	CheckTagExists(ctx context.Context, repoURL, tag string) (bool, error)
	GetBranchHeadSHA(ctx context.Context, repoURL, branch string) (*RepoResult, error)
	GetDefaultBranch(ctx context.Context, repoURL string) (string, error)
	FetchFile(ctx context.Context, repoURL, ref, filePath string) ([]byte, error)
}

// GitCredentials holds optional git authentication material.
// A zero value selects anonymous access.
type GitCredentials struct {
	Token    string
	Username string
	Password string
}

func (c GitCredentials) isEmpty() bool {
	return c.Token == "" && c.Username == "" && c.Password == ""
}

// NewGitClientForRepo returns a GitClient backed by the GitHub REST API for
// github.com repos, or the generic go-git client for all other hosts.
func NewGitClientForRepo(repoURL string, creds GitCredentials) (GitClient, error) {
	if IsGithubRepo(repoURL) {
		return newGitHubClientForCredentials(creds)
	}
	return newGenericGitClientForCredentials(creds)
}

func newGitHubClientForCredentials(creds GitCredentials) (GitClient, error) {
	if creds.isEmpty() {
		return newGitHubClientAnonymous(), nil
	}
	if creds.Token != "" {
		return newGitHubClientWithToken(creds.Token), nil
	}
	if creds.Username != "" && creds.Password != "" {
		return newGitHubClientWithBasicAuth(creds.Username, creds.Password), nil
	}
	return nil, fmt.Errorf("token or username and password must be provided for authentication")
}

func newGenericGitClientForCredentials(creds GitCredentials) (GitClient, error) {
	if creds.isEmpty() {
		return newGitClientAnonymous()
	}
	if creds.Token != "" {
		return newGitClientWithToken(creds.Token)
	}
	return newGitClient(creds.Username, creds.Password)
}
