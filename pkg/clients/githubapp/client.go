// Package githubapp is a minimal GitHub App API client covering the manifest
// conversion, installation, and repository discovery flows. It intentionally
// uses plain net/http rather than go-github: only a handful of endpoints are
// needed and the app-JWT auth model differs from the rest of the git client.
package githubapp

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

// DefaultAPIBaseURL is the public GitHub API endpoint.
const DefaultAPIBaseURL = "https://api.github.com"

// CloneUsername is the username GitHub expects alongside an installation
// token for git-over-HTTPS operations.
const CloneUsername = "x-access-token"

// AppCredentials are the app-level credentials produced by the manifest flow.
type AppCredentials struct {
	AppID         int64
	Slug          string
	PEM           string
	WebhookSecret string
	ClientID      string
	ClientSecret  string
}

// Token is a minted installation access token.
type Token struct {
	Value     string
	ExpiresAt time.Time
}

// Installation is a GitHub App installation on a user or org account.
type Installation struct {
	ID                  int64
	AccountLogin        string
	AccountType         string
	RepositorySelection string
}

// Repo is a repository visible to an installation.
type Repo struct {
	FullName      string
	CloneURL      string
	DefaultBranch string
	Private       bool
	PushedAt      *time.Time
	OwnerLogin    string
}

// RepoPage is one page of installation repositories.
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

type ClientSpec struct {
	// BaseURL is optional; it defaults to the public GitHub API.
	BaseURL    string
	HTTPClient *http.Client
}

type client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(spec ClientSpec) Client {
	baseURL := strings.TrimSuffix(spec.BaseURL, "/")
	if baseURL == "" {
		baseURL = DefaultAPIBaseURL
	}
	httpClient := spec.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &client{baseURL: baseURL, httpClient: httpClient}
}

// appJWT builds the RS256 app JWT GitHub expects: iat 60s in the past to
// absorb clock drift, 9 minute expiry, issuer = app ID.
func appJWT(creds *AppCredentials, now time.Time) (string, error) {
	key, err := parseRSAKey(creds.PEM)
	if err != nil {
		return "", err
	}
	claims := jwt.StandardClaims{
		IssuedAt:  now.Add(-60 * time.Second).Unix(),
		ExpiresAt: now.Add(9 * time.Minute).Unix(),
		Issuer:    strconv.FormatInt(creds.AppID, 10),
	}
	return jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
}

func parseRSAKey(pemData string) (*rsa.PrivateKey, error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(pemData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse app private key: %w", err)
	}
	return key, nil
}

func (c *client) doJSON(ctx context.Context, method, path, bearer string, out any) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp, fmt.Errorf("github api %s %s: unexpected status %d: %s", method, path, resp.StatusCode, truncate(body, 300))
	}
	if out != nil {
		if err := json.Unmarshal(body, out); err != nil {
			return resp, fmt.Errorf("github api %s %s: failed to decode response: %w", method, path, err)
		}
	}
	return resp, nil
}

func truncate(b []byte, n int) string {
	if len(b) > n {
		return string(b[:n]) + "..."
	}
	return string(b)
}

func (c *client) ConvertManifestCode(ctx context.Context, code string) (*AppCredentials, error) {
	var payload struct {
		ID            int64  `json:"id"`
		Slug          string `json:"slug"`
		PEM           string `json:"pem"`
		WebhookSecret string `json:"webhook_secret"`
		ClientID      string `json:"client_id"`
		ClientSecret  string `json:"client_secret"`
	}
	if _, err := c.doJSON(ctx, http.MethodPost, "/app-manifests/"+url.PathEscape(code)+"/conversions", "", &payload); err != nil {
		return nil, err
	}
	return &AppCredentials{
		AppID:         payload.ID,
		Slug:          payload.Slug,
		PEM:           payload.PEM,
		WebhookSecret: payload.WebhookSecret,
		ClientID:      payload.ClientID,
		ClientSecret:  payload.ClientSecret,
	}, nil
}

type installationPayload struct {
	ID      int64 `json:"id"`
	Account struct {
		Login string `json:"login"`
		Type  string `json:"type"`
	} `json:"account"`
	RepositorySelection string `json:"repository_selection"`
}

func (p installationPayload) toInstallation() Installation {
	return Installation{
		ID:                  p.ID,
		AccountLogin:        p.Account.Login,
		AccountType:         p.Account.Type,
		RepositorySelection: p.RepositorySelection,
	}
}

func (c *client) ListInstallations(ctx context.Context, creds *AppCredentials) ([]Installation, error) {
	token, err := appJWT(creds, time.Now())
	if err != nil {
		return nil, err
	}

	var installations []Installation
	for page := 1; ; page++ {
		var payload []installationPayload
		path := fmt.Sprintf("/app/installations?per_page=100&page=%d", page)
		if _, err := c.doJSON(ctx, http.MethodGet, path, token, &payload); err != nil {
			return nil, err
		}
		for _, p := range payload {
			installations = append(installations, p.toInstallation())
		}
		if len(payload) < 100 {
			return installations, nil
		}
	}
}

func (c *client) MintInstallationToken(ctx context.Context, creds *AppCredentials, installationID int64) (*Token, error) {
	token, err := appJWT(creds, time.Now())
	if err != nil {
		return nil, err
	}
	var payload struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	path := fmt.Sprintf("/app/installations/%d/access_tokens", installationID)
	if _, err := c.doJSON(ctx, http.MethodPost, path, token, &payload); err != nil {
		return nil, err
	}
	return &Token{Value: payload.Token, ExpiresAt: payload.ExpiresAt}, nil
}

type repoPayload struct {
	FullName      string     `json:"full_name"`
	CloneURL      string     `json:"clone_url"`
	DefaultBranch string     `json:"default_branch"`
	Private       bool       `json:"private"`
	PushedAt      *time.Time `json:"pushed_at"`
	Owner         struct {
		Login string `json:"login"`
	} `json:"owner"`
}

func (p repoPayload) toRepo() Repo {
	return Repo{
		FullName:      p.FullName,
		CloneURL:      p.CloneURL,
		DefaultBranch: p.DefaultBranch,
		Private:       p.Private,
		PushedAt:      p.PushedAt,
		OwnerLogin:    p.Owner.Login,
	}
}

func (c *client) ListInstallationRepos(ctx context.Context, creds *AppCredentials, installationID int64, query string, page int) (*RepoPage, error) {
	installationToken, err := c.MintInstallationToken(ctx, creds, installationID)
	if err != nil {
		return nil, err
	}
	if page <= 0 {
		page = 1
	}

	var payload struct {
		TotalCount   int           `json:"total_count"`
		Repositories []repoPayload `json:"repositories"`
	}
	path := fmt.Sprintf("/installation/repositories?per_page=100&page=%d", page)
	if _, err := c.doJSON(ctx, http.MethodGet, path, installationToken.Value, &payload); err != nil {
		return nil, err
	}

	result := &RepoPage{
		Page:       page,
		TotalCount: payload.TotalCount,
		HasNext:    page*100 < payload.TotalCount,
	}
	query = strings.ToLower(strings.TrimSpace(query))
	for _, p := range payload.Repositories {
		if query != "" && !strings.Contains(strings.ToLower(p.FullName), query) {
			continue
		}
		result.Repos = append(result.Repos, p.toRepo())
	}
	return result, nil
}

func (c *client) GetRepo(ctx context.Context, creds *AppCredentials, installationID int64, owner, repo string) (*Repo, error) {
	installationToken, err := c.MintInstallationToken(ctx, creds, installationID)
	if err != nil {
		return nil, err
	}
	var payload repoPayload
	path := fmt.Sprintf("/repos/%s/%s", url.PathEscape(owner), url.PathEscape(repo))
	if _, err := c.doJSON(ctx, http.MethodGet, path, installationToken.Value, &payload); err != nil {
		return nil, err
	}
	result := payload.toRepo()
	return &result, nil
}

func (c *client) ListBranches(ctx context.Context, creds *AppCredentials, installationID int64, owner, repo string) ([]string, error) {
	installationToken, err := c.MintInstallationToken(ctx, creds, installationID)
	if err != nil {
		return nil, err
	}

	var branches []string
	for page := 1; ; page++ {
		var payload []struct {
			Name string `json:"name"`
		}
		path := fmt.Sprintf("/repos/%s/%s/branches?per_page=100&page=%d", url.PathEscape(owner), url.PathEscape(repo), page)
		if _, err := c.doJSON(ctx, http.MethodGet, path, installationToken.Value, &payload); err != nil {
			return nil, err
		}
		for _, b := range payload {
			branches = append(branches, b.Name)
		}
		if len(payload) < 100 {
			return branches, nil
		}
	}
}
