package clients

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
)

type GitClient interface {
	CheckAccess(ctx context.Context, repoURL string) (bool, error)
	GetBranchHeadSHA(ctx context.Context, repoURL, branch string) (*RepoResult, error)
	GetTagSHA(ctx context.Context, repoURL, tag string) (string, error)
	CheckTagExists(ctx context.Context, repoURL, tag string) (bool, error)
}

type gitClient struct {
	auth transport.AuthMethod
}

type RepoResult struct {
	HeadSHA string
	Branch  string
}

func NewGitClient(username, password string) (GitClient, error) {
	var auth transport.AuthMethod
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password must be provided for authentication")
	}
	if username != "" && password != "" {
		auth = &http.BasicAuth{
			Username: username,
			Password: password,
		}
	}

	return &gitClient{
		auth: auth,
	}, nil
}

func NewGitClientWithToken(token string) (GitClient, error) {
	var auth transport.AuthMethod
	if token == "" {
		return nil, fmt.Errorf("token must be provided for authentication")
	}
	if token != "" {
		auth = &http.BasicAuth{
			Username: "token",
			Password: token,
		}
	}

	return &gitClient{
		auth: auth,
	}, nil
}

func NewGitClientAnonymous() (GitClient, error) {
	return &gitClient{
		auth: nil,
	}, nil
}

func (g *gitClient) CheckAccess(ctx context.Context, repoURL string) (bool, error) {
	// Create a remote to list references
	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoURL},
	})

	// List references to check clone access and find the branch
	_, err := rem.List(&git.ListOptions{
		Auth: g.auth,
	})
	if err != nil {
		if isGitAuthError(err) {
			return false, fmt.Errorf("authentication failed: %v", err)
		} else if isGitNotFoundError(err) {
			return false, fmt.Errorf("repository not found: %v", err)
		}
		return false, fmt.Errorf("failed to access git repo: %v", err)
	}

	return true, nil
}

// verify branch existence and get HEAD SHA
func (g *gitClient) GetBranchHeadSHA(ctx context.Context, repoURL, branch string) (*RepoResult, error) {
	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoURL},
	})

	ref, err := rem.List(&git.ListOptions{
		Auth: g.auth,
	})
	if err != nil {
		return nil, err
	}
	result := &RepoResult{
		Branch: branch,
	}
	branchRefName := plumbing.NewBranchReferenceName(branch)
	for _, r := range ref {
		if r.Name() == branchRefName {
			result.HeadSHA = r.Hash().String()
			return result, nil
		}
	}

	return nil, fmt.Errorf("branch '%s' not found in repository", branch)
}

func (g *gitClient) GetTagSHA(ctx context.Context, repoURL, tag string) (string, error) {
	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoURL},
	})
	refs, err := rem.List(&git.ListOptions{
		Auth: g.auth,
	})
	if err != nil {
		return "", err
	}

	tagRefName := plumbing.NewTagReferenceName(tag)
	for _, r := range refs {
		if r.Name() == tagRefName {
			return r.Hash().String(), nil
		}
	}
	return "", fmt.Errorf("tag '%s' not found in repository", tag)
}

func (g *gitClient) CheckTagExists(ctx context.Context, repoURL, tag string) (bool, error) {
	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoURL},
	})
	ref, err := rem.List(&git.ListOptions{
		Auth: g.auth,
	})
	if err != nil {
		return false, err
	}

	tagRefName := plumbing.NewTagReferenceName(tag)

	for _, r := range ref {
		if r.Name() == tagRefName {
			return true, nil
		}
	}
	return false, nil
}

// Helper function to check if error is authentication related
func isGitAuthError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "authentication") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "permission denied") ||
		strings.Contains(errStr, "access denied") ||
		strings.Contains(errStr, "credentials") ||
		strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "403")
}

// Helper function to check if error is not found related
func isGitNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "repository does not exist") ||
		strings.Contains(errStr, "could not read") ||
		strings.Contains(errStr, "404")
}
