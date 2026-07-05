package models

import (
	"encoding/json"
	"time"
)

type GitIntegrationType string

const (
	GitIntegrationTypeGitCredentials GitIntegrationType = "git_credentials"
	GitIntegrationTypeGitHubApp      GitIntegrationType = "github_app"
)

func (t GitIntegrationType) Valid() bool {
	switch t {
	case GitIntegrationTypeGitCredentials, GitIntegrationTypeGitHubApp:
		return true
	default:
		return false
	}
}

type GitIntegrationStatus string

const (
	GitIntegrationStatusActive         GitIntegrationStatus = "active"
	GitIntegrationStatusPendingInstall GitIntegrationStatus = "pending_install"
	GitIntegrationStatusInstalled      GitIntegrationStatus = "installed"
)

// Annotations stamped on cluster secrets synced from minted GitHub App
// installation tokens, driving hub-side refresh.
const (
	GitTokenMintedAtAnnotation  = "stackdome.io/git-token-minted-at"
	GitTokenExpiresAtAnnotation = "stackdome.io/git-token-expires-at"
	GitIntegrationIDAnnotation  = "stackdome.io/git-integration-id"
)

// OAuthProviderGitHubAppManifest scopes manifest-flow state records.
const OAuthProviderGitHubAppManifest = "github_app_manifest"

// Condition recorded on image builds when a GitHub App token refresh fails.
const (
	GitTokenRefreshFailedConditionType   = "GitTokenRefreshFailed"
	GitTokenRefreshFailedConditionReason = "MintFailed"
)

// GitHubAppCredentials are the app-level credentials returned by the GitHub
// App manifest conversion; stored encrypted in the integration's auth blob.
type GitHubAppCredentials struct {
	AppID         int64  `json:"app_id"`
	Slug          string `json:"slug"`
	PEM           string `json:"pem"`
	WebhookSecret string `json:"webhook_secret"`
	ClientID      string `json:"client_id"`
	ClientSecret  string `json:"client_secret"`
}

// GitIntegrationBasicAuth is username/password clone auth.
type GitIntegrationBasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// GitIntegrationAuth is the decrypted auth blob for a git integration.
// Exactly one of Basic or Token is set for git_credentials integrations;
// github_app integrations carry app credentials instead (added with the
// GitHub App flow). SSH auth is deferred.
type GitIntegrationAuth struct {
	Basic     *GitIntegrationBasicAuth `json:"basic,omitempty"`
	Token     string                   `json:"token,omitempty"`
	GitHubApp *GitHubAppCredentials    `json:"github_app,omitempty"`
}

func (a GitIntegrationAuth) Empty() bool {
	return a.Basic == nil && a.Token == "" && a.GitHubApp == nil
}

func (a GitIntegrationAuth) Marshal() ([]byte, error) {
	return json.Marshal(a)
}

// GitIntegration is an org-level git auth integration, auto-attached to
// repositories by host match when no explicit override is set.
type GitIntegration struct {
	ID             string             `gorm:"primary_key;default:gen_random_uuid()"`
	OrganisationID string             `gorm:"not null;index"`
	Type           GitIntegrationType `gorm:"not null"`
	// Host is the git host this integration covers (required for
	// git_credentials; github.com for github_app).
	Host          string `gorm:"index"`
	EncryptedAuth string `gorm:"type:text"`
	Status        GitIntegrationStatus
	DataHash      string `gorm:"not null"` // for change detection / release freezing
	CreatedAt     time.Time
	UpdatedAt     time.Time

	// Transient decrypted auth; never stored or serialized.
	Auth *GitIntegrationAuth `gorm:"-" json:"-"`
	// Transient GitHub install URL for github_app integrations.
	InstallURL string `gorm:"-" json:"-"`
}

// GitHubAppManifestFlow is what the UI needs to send the user to GitHub's
// app-creation page.
type GitHubAppManifestFlow struct {
	Manifest  map[string]any
	GitHubURL string
	State     string
}

// GitHubAppMintResult is a freshly minted installation token plus the
// identity of the integration it came from.
type GitHubAppMintResult struct {
	IntegrationID string
	Host          string
	Token         string
	MintedAt      time.Time
	ExpiresAt     time.Time
	DataHash      string
}
