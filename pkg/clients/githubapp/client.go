// Package githubapp is a thin GitHub App API client covering the manifest
// conversion, installation, and repository discovery flows. It wraps
// go-github for the REST calls and ghinstallation for the App-JWT /
// installation-token auth model, which differs from the rest of the git
// client (that uses a static token or basic auth).
package githubapp

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v88/github"
)

// DefaultAPIBaseURL is the public GitHub API endpoint.
const DefaultAPIBaseURL = "https://api.github.com"

// CloneUsername is the username GitHub expects alongside an installation
// token for git-over-HTTPS operations.
const CloneUsername = "x-access-token"

// AppCredentials are the app-level credentials produced by the manifest flow.
// Every field is a secret; the PEM in particular is the app's RSA private key
// used to sign app JWTs.
type AppCredentials struct {
	AppID         int64  // numeric GitHub App ID
	Slug          string // URL slug, e.g. "stackdome-ci"
	PEM           string // RSA private key (PEM) used to sign app JWTs
	WebhookSecret string // secret for verifying webhook payload signatures
	ClientID      string // OAuth client ID
	ClientSecret  string // OAuth client secret
}

// Token is a minted installation access token. Installation tokens are
// short-lived (~1h), so ExpiresAt is what drives refresh scheduling.
type Token struct {
	Value     string
	ExpiresAt time.Time
}

// Installation is a GitHub App installation on a user or org account.
type Installation struct {
	ID                  int64
	AccountLogin        string // account (user or org) the app is installed on
	AccountType         string // "User" or "Organization"
	RepositorySelection string // "all" or "selected"
}

// Repo is a repository visible to an installation.
type Repo struct {
	FullName      string // "owner/name"
	CloneURL      string // HTTPS clone URL
	DefaultBranch string
	Private       bool
	PushedAt      *time.Time // last push time; nil if never pushed
	OwnerLogin    string
}

// RepoPage is one page of installation repositories. TotalCount is GitHub's
// unfiltered total for the installation, so when a query filter is applied
// len(Repos) may be smaller — even zero — while HasNext is still true.
type RepoPage struct {
	Repos      []Repo
	Page       int
	TotalCount int
	HasNext    bool
}

//go:generate mockgen -destination=../../mocks/mock_githubapp_client.go -package=mocks -mock_names Client=MockGitHubAppClient github.com/ashishmax31/stackdome-api-server/pkg/clients/githubapp Client
type Client interface {
	// ConvertManifestCode exchanges the temporary code from the manifest
	// redirect for the newly created app's credentials.
	ConvertManifestCode(ctx context.Context, code string) (*AppCredentials, error)
	// ListInstallations lists every installation of the app (app JWT auth),
	// following pagination.
	ListInstallations(ctx context.Context, creds *AppCredentials) ([]Installation, error)
	// MintInstallationToken creates a short-lived installation access token.
	MintInstallationToken(ctx context.Context, creds *AppCredentials, installationID int64) (*Token, error)
	// ListInstallationRepos lists one page of repositories the installation
	// can access, optionally filtered by a substring query.
	ListInstallationRepos(ctx context.Context, creds *AppCredentials, installationID int64, query string, page int) (*RepoPage, error)
	// GetRepo fetches repository details through the installation.
	GetRepo(ctx context.Context, creds *AppCredentials, installationID int64, owner, repo string) (*Repo, error)
	// ListBranches lists branch names through the installation.
	ListBranches(ctx context.Context, creds *AppCredentials, installationID int64, owner, repo string) ([]string, error)
}

// ClientSpec configures a Client.
type ClientSpec struct {
	// BaseURL is optional; it defaults to the public GitHub API. Set it to a
	// GitHub Enterprise or test-server URL to point the client elsewhere.
	BaseURL string
	// HTTPClient is optional. Its Transport is used as the base round-tripper
	// beneath the app-JWT / installation-token auth transports, and its Timeout
	// (when set) overrides the 30s default.
	HTTPClient *http.Client
}

// client is the default Client. It holds no per-request auth state: each call
// builds a short-lived go-github client from the AppCredentials passed in.
type client struct {
	baseURL       string
	baseTransport http.RoundTripper
	timeout       time.Duration
}

// NewClient returns a GitHub App client. Because credentials are supplied per
// call rather than stored, a single instance is safe to share across orgs and
// goroutines.
func NewClient(spec ClientSpec) Client {
	baseURL := strings.TrimSuffix(spec.BaseURL, "/")
	if baseURL == "" {
		baseURL = DefaultAPIBaseURL
	}
	baseTransport := http.DefaultTransport
	timeout := 30 * time.Second
	if spec.HTTPClient != nil {
		if spec.HTTPClient.Transport != nil {
			baseTransport = spec.HTTPClient.Transport
		}
		if spec.HTTPClient.Timeout != 0 {
			timeout = spec.HTTPClient.Timeout
		}
	}
	return &client{baseURL: baseURL, baseTransport: baseTransport, timeout: timeout}
}

// newGitHubClient builds a go-github client with the given auth transport (nil
// for unauthenticated) pointed at c.baseURL.
func (c *client) newGitHubClient(transport http.RoundTripper) (*github.Client, error) {
	opts := []github.ClientOptionsFunc{github.WithTimeout(c.timeout)}
	if transport != nil {
		opts = append(opts, github.WithTransport(transport))
	}
	if c.baseURL != DefaultAPIBaseURL {
		base := c.baseURL + "/"
		opts = append(opts, github.WithURLs(&base, &base))
	}
	return github.NewClient(opts...)
}

// anonClient is unauthenticated — used for the manifest conversion, which is
// authorized by the one-time code in the path.
func (c *client) anonClient() (*github.Client, error) {
	return c.newGitHubClient(nil)
}

// appClient authenticates as the app itself (signed app JWT) — used to list
// installations and mint installation tokens.
func (c *client) appClient(creds *AppCredentials) (*github.Client, error) {
	atr, err := ghinstallation.NewAppsTransport(c.baseTransport, creds.AppID, []byte(creds.PEM))
	if err != nil {
		return nil, fmt.Errorf("failed to build app transport: %w", err)
	}
	atr.BaseURL = c.baseURL
	return c.newGitHubClient(atr)
}

// installationClient authenticates as a specific installation. The transport
// mints and caches the installation token on demand.
func (c *client) installationClient(creds *AppCredentials, installationID int64) (*github.Client, error) {
	itr, err := ghinstallation.New(c.baseTransport, creds.AppID, installationID, []byte(creds.PEM))
	if err != nil {
		return nil, fmt.Errorf("failed to build installation transport: %w", err)
	}
	itr.BaseURL = c.baseURL
	return c.newGitHubClient(itr)
}

func (c *client) ConvertManifestCode(ctx context.Context, code string) (*AppCredentials, error) {
	gh, err := c.anonClient()
	if err != nil {
		return nil, err
	}
	cfg, _, err := gh.Apps.CompleteAppManifest(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to convert app manifest: %w", err)
	}
	return &AppCredentials{
		AppID:         cfg.GetID(),
		Slug:          cfg.GetSlug(),
		PEM:           cfg.GetPEM(),
		WebhookSecret: cfg.GetWebhookSecret(),
		ClientID:      cfg.GetClientID(),
		ClientSecret:  cfg.GetClientSecret(),
	}, nil
}

func (c *client) ListInstallations(ctx context.Context, creds *AppCredentials) ([]Installation, error) {
	gh, err := c.appClient(creds)
	if err != nil {
		return nil, err
	}
	var installations []Installation
	opts := &github.ListOptions{PerPage: 100}
	for {
		page, resp, err := gh.Apps.ListInstallations(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list installations: %w", err)
		}
		for _, in := range page {
			installations = append(installations, Installation{
				ID:                  in.GetID(),
				AccountLogin:        in.GetAccount().GetLogin(),
				AccountType:         in.GetAccount().GetType(),
				RepositorySelection: in.GetRepositorySelection(),
			})
		}
		if resp.NextPage == 0 {
			return installations, nil
		}
		opts.Page = resp.NextPage
	}
}

func (c *client) MintInstallationToken(ctx context.Context, creds *AppCredentials, installationID int64) (*Token, error) {
	gh, err := c.appClient(creds)
	if err != nil {
		return nil, err
	}
	tok, _, err := gh.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to mint installation token: %w", err)
	}
	return &Token{Value: tok.GetToken(), ExpiresAt: tok.GetExpiresAt().Time}, nil
}

func (c *client) ListInstallationRepos(ctx context.Context, creds *AppCredentials, installationID int64, query string, page int) (*RepoPage, error) {
	gh, err := c.installationClient(creds, installationID)
	if err != nil {
		return nil, err
	}
	if page <= 0 {
		page = 1
	}
	list, resp, err := gh.Apps.ListRepos(ctx, &github.ListOptions{Page: page, PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to list installation repositories: %w", err)
	}
	result := &RepoPage{
		Page:       page,
		TotalCount: list.GetTotalCount(),
		HasNext:    resp.NextPage != 0,
	}
	// The installation-repositories endpoint has no server-side search, so the
	// query is applied client-side over the fetched page; TotalCount and HasNext
	// still reflect the unfiltered listing.
	query = strings.ToLower(strings.TrimSpace(query))
	for _, r := range list.Repositories {
		if query != "" && !strings.Contains(strings.ToLower(r.GetFullName()), query) {
			continue
		}
		result.Repos = append(result.Repos, repoFrom(r))
	}
	return result, nil
}

func (c *client) GetRepo(ctx context.Context, creds *AppCredentials, installationID int64, owner, repo string) (*Repo, error) {
	gh, err := c.installationClient(creds, installationID)
	if err != nil {
		return nil, err
	}
	r, _, err := gh.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository %s/%s: %w", owner, repo, err)
	}
	result := repoFrom(r)
	return &result, nil
}

func (c *client) ListBranches(ctx context.Context, creds *AppCredentials, installationID int64, owner, repo string) ([]string, error) {
	gh, err := c.installationClient(creds, installationID)
	if err != nil {
		return nil, err
	}
	var branches []string
	opts := &github.BranchListOptions{ListOptions: github.ListOptions{PerPage: 100}}
	for {
		page, resp, err := gh.Repositories.ListBranches(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list branches for %s/%s: %w", owner, repo, err)
		}
		for _, b := range page {
			branches = append(branches, b.GetName())
		}
		if resp.NextPage == 0 {
			return branches, nil
		}
		opts.Page = resp.NextPage
	}
}

// repoFrom projects a go-github repository onto the client's Repo type.
func repoFrom(r *github.Repository) Repo {
	var pushedAt *time.Time
	if ts := r.GetPushedAt(); !ts.IsZero() {
		t := ts.Time
		pushedAt = &t
	}
	return Repo{
		FullName:      r.GetFullName(),
		CloneURL:      r.GetCloneURL(),
		DefaultBranch: r.GetDefaultBranch(),
		Private:       r.GetPrivate(),
		PushedAt:      pushedAt,
		OwnerLogin:    r.GetOwner().GetLogin(),
	}
}
