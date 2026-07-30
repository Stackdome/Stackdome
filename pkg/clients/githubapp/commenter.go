//go:generate mockgen -source=commenter.go -destination=commenter_mock.go -package=githubapp
package githubapp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/go-github/v88/github"
)

// ErrCommentNotFound reports that the target comment no longer exists (e.g. a
// human deleted it); callers re-create instead of failing.
var ErrCommentNotFound = errors.New("comment not found")

// PullRequestCommenter posts and edits PR conversation comments using an
// installation token minted per call.
type PullRequestCommenter interface {
	CreateComment(ctx context.Context, token, owner, repo string, prNumber int, body string) (int64, error)
	EditComment(ctx context.Context, token, owner, repo string, commentID int64, body string) error
}

type PullRequestCommenterSpec struct {
	// BaseURL overrides the GitHub API base URL (tests); empty uses api.github.com.
	BaseURL string
}

type pullRequestCommenter struct {
	baseURL string
}

func NewPullRequestCommenter(spec PullRequestCommenterSpec) PullRequestCommenter {
	return &pullRequestCommenter{baseURL: spec.BaseURL}
}

func (c *pullRequestCommenter) client(token string) (*github.Client, error) {
	opts := []github.ClientOptionsFunc{github.WithAuthToken(token)}
	if c.baseURL != "" {
		base := c.baseURL + "/"
		opts = append(opts, github.WithURLs(&base, &base))
	}
	gh, err := github.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to build commenter client: %w", err)
	}
	return gh, nil
}

func (c *pullRequestCommenter) CreateComment(ctx context.Context, token, owner, repo string, prNumber int, body string) (int64, error) {
	gh, err := c.client(token)
	if err != nil {
		return 0, err
	}
	comment, _, err := gh.Issues.CreateComment(ctx, owner, repo, prNumber, &github.IssueComment{Body: github.Ptr(body)})
	if err != nil {
		return 0, fmt.Errorf("failed to create PR comment: %w", err)
	}
	return comment.GetID(), nil
}

func (c *pullRequestCommenter) EditComment(ctx context.Context, token, owner, repo string, commentID int64, body string) error {
	gh, err := c.client(token)
	if err != nil {
		return err
	}
	_, resp, err := gh.Issues.EditComment(ctx, owner, repo, commentID, &github.IssueComment{Body: github.Ptr(body)})
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return ErrCommentNotFound
		}
		return fmt.Errorf("failed to edit PR comment: %w", err)
	}
	return nil
}
