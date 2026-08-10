package config

import (
	"fmt"
	"strings"
)

type SSLMode string

const (
	DBSSLModeDisable SSLMode = "disable"
	DBSSLModeRequire SSLMode = "require"
)

type ApplicationConfig struct {
	Server                   *ServerConfig         `json:"server"`
	Database                 *DatabaseConfig       `json:"database"`
	JwtSecret                string                `json:"jwt_secret"`
	EncryptionKey            string                `json:"encryption_key"`
	LogLevel                 string                `json:"log_level"`
	LogFormat                string                `json:"log_format"`
	RuntimeMode              RuntimeMode           `json:"runtime_mode"`
	StackdomeCloudConfigPath string                `json:"stackdome_cloud_config_path"`
	StackdomeCloud           *StackdomeCloudConfig `json:"stackdome_cloud,omitempty"`
	TurnstileSecret          string                `json:"-"`
	GitHubOAuth              *GitHubOAuthConfig    `json:"github_oauth"`
	// GitHubApp is the platform-wide GitHub App; unset means per-org apps.
	GitHubApp *GitHubAppConfig `json:"github_app"`
	// ServerExternalURL is the externally reachable base URL of the hub,
	// required for the GitHub App manifest flow (browser redirects, webhooks).
	ServerExternalURL string `json:"server_external_url"`
	// GitHubAPIBaseURL overrides the GitHub API endpoint (tests, GHES).
	GitHubAPIBaseURL string `json:"github_api_base_url"`
	// PlatformCluster holds the PLATFORM_CLUSTER_* config seeded at boot.
	PlatformCluster *ClusterConfig `json:"platform_cluster"`
}

func (c *ApplicationConfig) LoadEnvVariables() error {
	c.Server.LoadEnvVariables()
	c.Database.LoadEnvVariables()

	if val, ok := EnvJWTSecret.Lookup(); ok {
		c.JwtSecret = val
	}

	if val, ok := EnvLogLevel.Lookup(); ok {
		c.LogLevel = val
	}

	if val, ok := EnvLogFormat.Lookup(); ok {
		c.LogFormat = val
	}

	if val, ok := EnvEncryptionKey.Lookup(); ok {
		c.EncryptionKey = val
	}

	if val, ok := EnvRuntimeMode.Lookup(); ok {
		c.RuntimeMode = RuntimeMode(val)
	}
	if val, ok := EnvStackdomeCloudConfig.Lookup(); ok {
		c.StackdomeCloudConfigPath = val
	}
	if val, ok := EnvTurnstileSecret.Lookup(); ok {
		c.TurnstileSecret = val
	}

	c.GitHubOAuth.LoadEnvVariables()
	c.GitHubApp.LoadEnvVariables()
	c.PlatformCluster.LoadEnvVariables()

	if val, ok := EnvServerExternalURL.Lookup(); ok {
		c.ServerExternalURL = val
	}
	if val, ok := EnvGitHubAPIBaseURL.Lookup(); ok {
		c.GitHubAPIBaseURL = val
	}

	// The GitHub OAuth redirect is just the hub base plus a fixed callback path,
	// so derive it from SERVER_EXTERNAL_URL when GITHUB_REDIRECT_URI is unset.
	// The explicit env var stays as an override for unusual deployments.
	if c.GitHubOAuth.RedirectURI == "" && c.ServerExternalURL != "" {
		c.GitHubOAuth.RedirectURI = strings.TrimSuffix(c.ServerExternalURL, "/") + gitHubOAuthCallbackPath
	}

	return nil
}

// gitHubOAuthCallbackPath is the SPA route GitHub sends the browser back to
// (frontend/src/App.tsx). That page reads code+state and calls the hub's
// /api/v1/auth/github/callback itself, so the API route is never GitHub's
// redirect target — sending the browser there would render raw JSON and no
// session would be established.
const gitHubOAuthCallbackPath = "/auth/github/callback"

func (c *ApplicationConfig) Validate() error {
	validateFuncs := []func() error{
		c.Server.Validate,
		c.Database.Validate,
		c.validateRuntimeMode,
		func() error {
			if c.JwtSecret == "" {
				return fmt.Errorf("jwt secret is required")
			}
			return nil
		},
		func() error {
			if c.LogLevel == "" {
				return fmt.Errorf("log level is required")
			}
			return nil
		},
		func() error {
			if c.EncryptionKey == "" {
				return fmt.Errorf("encryption key is required")
			}
			if len(c.EncryptionKey) < 64 || len(c.EncryptionKey) > 1024 {
				return fmt.Errorf("encryption key must be at least 64 characters and at most 1024 characters for security")
			}
			return nil
		},
	}

	for _, f := range validateFuncs {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func (c *ApplicationConfig) validateRuntimeMode() error {
	if c.RuntimeMode != RuntimeModeSelfHosted && c.RuntimeMode != RuntimeModeStackdomeCloud {
		return fmt.Errorf("runtime mode must be %q or %q", RuntimeModeSelfHosted, RuntimeModeStackdomeCloud)
	}
	if c.IsStackdomeCloud() && c.StackdomeCloud == nil {
		return fmt.Errorf("stackdome Cloud config is required in %q runtime mode", RuntimeModeStackdomeCloud)
	}
	return nil
}

type ClusterConfig struct {
	ClusterURL    string `yaml:"cluster_url"`
	ClusterCAData string `yaml:"cluster_ca_data"`
	Token         string `yaml:"token"`
}

func (c *ClusterConfig) Validate() error {
	if c.ClusterURL == "" {
		return fmt.Errorf("cluster url is required")
	}

	if c.ClusterCAData == "" {
		return fmt.Errorf("cluster ca data is required")
	}

	if c.Token == "" {
		return fmt.Errorf("cluster token is required")
	}

	return nil
}

func (c *ClusterConfig) LoadEnvVariables() {
	if val, ok := EnvPlatformClusterAPIURL.Lookup(); ok {
		c.ClusterURL = val
	}

	if val, ok := EnvPlatformClusterCAData.Lookup(); ok {
		c.ClusterCAData = val
	}

	if val, ok := EnvPlatformClusterToken.Lookup(); ok {
		c.Token = val
	}
}

func (c *ClusterConfig) IsSet() bool {
	return c.ClusterURL != "" && c.ClusterCAData != "" && c.Token != ""
}

func (c *ClusterConfig) AnySet() bool {
	return c.ClusterURL != "" || c.ClusterCAData != "" || c.Token != ""
}

type DatabaseConfig struct {
	Dialect            string  `json:"dialect"`
	SSLMode            SSLMode `json:"sslmode"`
	RootCertFile       string
	Debug              bool `json:"debug"`
	MaxOpenConnections int  `json:"max_connections"`
	DBConnectionConfig
}

func (c *DatabaseConfig) Validate() error {
	if c.Dialect == "" {
		return fmt.Errorf("dialect is required")
	}

	if c.SSLMode == "" {
		return fmt.Errorf("sslmode is required")
	}

	if c.SSLMode == DBSSLModeRequire && c.RootCertFile == "" {
		return fmt.Errorf("root_cert_file is required when sslmode is require")
	}

	if c.MaxOpenConnections == 0 {
		return fmt.Errorf("max_connections is required")
	}

	if c.Host == "" {
		return fmt.Errorf("host is required")
	}

	if c.Port == 0 {
		return fmt.Errorf("port is required")
	}

	if c.Name == "" {
		return fmt.Errorf("name is required")
	}

	if c.Username == "" {
		return fmt.Errorf("username is required")
	}

	if c.Password == "" {
		return fmt.Errorf("password is required")
	}

	return nil
}

type DBConnectionConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Name     string `json:"name" yaml:"name"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

func NewApplicationConfig() *ApplicationConfig {
	return &ApplicationConfig{
		Server:          NewServerConfig(),
		Database:        NewDatabaseConfig(),
		LogLevel:        "info",
		LogFormat:       "json",
		GitHubOAuth:     NewGitHubOAuthConfig(),
		GitHubApp:       &GitHubAppConfig{},
		PlatformCluster: &ClusterConfig{},
		RuntimeMode:     RuntimeModeSelfHosted,
	}
}

func (c *ApplicationConfig) IsStackdomeCloud() bool {
	return c.RuntimeMode == RuntimeModeStackdomeCloud
}

func (c *ApplicationConfig) WorkspaceUsersEnabled() bool {
	return !c.IsStackdomeCloud() || (c.StackdomeCloud != nil && c.StackdomeCloud.Features.WorkspaceUsers)
}

func (c *ApplicationConfig) CustomDomainsEnabled() bool {
	return !c.IsStackdomeCloud() || (c.StackdomeCloud != nil && c.StackdomeCloud.Features.CustomDomains)
}

func (c *ApplicationConfig) ExternalPostgresImportEnabled() bool {
	return !c.IsStackdomeCloud() || (c.StackdomeCloud != nil && c.StackdomeCloud.Features.ExternalPostgresImport)
}

func (c *ApplicationConfig) LoadStackdomeCloudConfig() error {
	if !c.IsStackdomeCloud() {
		if c.StackdomeCloudConfigPath != "" {
			return fmt.Errorf("STACKDOME_CLOUD_CONFIG requires RUNTIME_MODE=%s", RuntimeModeStackdomeCloud)
		}
		if c.TurnstileSecret != "" {
			return fmt.Errorf("TURNSTILE_SECRET requires RUNTIME_MODE=%s", RuntimeModeStackdomeCloud)
		}
		c.StackdomeCloud = nil
		return nil
	}
	if c.StackdomeCloudConfigPath == "" {
		return fmt.Errorf("STACKDOME_CLOUD_CONFIG is required in %q runtime mode", RuntimeModeStackdomeCloud)
	}
	cloudConfig, err := LoadStackdomeCloudConfig(c.StackdomeCloudConfigPath)
	if err != nil {
		return err
	}
	if cloudConfig.Signup.Turnstile.Enabled && c.TurnstileSecret == "" {
		return fmt.Errorf("TURNSTILE_SECRET is required when signup Turnstile is enabled")
	}
	c.StackdomeCloud = cloudConfig
	return nil
}

func NewGitHubOAuthConfig() *GitHubOAuthConfig {
	return &GitHubOAuthConfig{}
}

type GitHubOAuthConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
}

func (c *GitHubOAuthConfig) Enabled() bool {
	return c.ClientID != "" && c.ClientSecret != ""
}

// GitHubAppConfig is the platform-wide GitHub App every org installs. Not
// configured: each org creates its own app through the manifest flow. Login
// uses the same app's client id/secret via the GITHUB_CLIENT_* vars.
type GitHubAppConfig struct {
	AppID         int64  `json:"app_id"`
	Slug          string `json:"slug"`
	PrivateKey    string `json:"private_key"`
	WebhookSecret string `json:"webhook_secret"`
}

func (c *GitHubAppConfig) Configured() bool {
	return c.AppID != 0 && c.Slug != "" && c.PrivateKey != "" && c.WebhookSecret != ""
}

func (c *GitHubAppConfig) LoadEnvVariables() {
	if val, ok := EnvGitHubAppID.Lookup(); ok {
		c.AppID = int64(val)
	}
	if val, ok := EnvGitHubAppSlug.Lookup(); ok {
		c.Slug = val
	}
	if val, ok := EnvGitHubAppPrivateKey.Lookup(); ok {
		c.PrivateKey = val
	}
	if val, ok := EnvGitHubAppWebhookSecret.Lookup(); ok {
		c.WebhookSecret = val
	}
}

func (c *GitHubOAuthConfig) LoadEnvVariables() {
	if val, ok := EnvGitHubClientID.Lookup(); ok {
		c.ClientID = val
	}

	if val, ok := EnvGitHubClientSecret.Lookup(); ok {
		c.ClientSecret = val
	}

	if val, ok := EnvGitHubRedirectURI.Lookup(); ok {
		c.RedirectURI = val
	}
}

type ServerConfig struct {
	Hostname           string `json:"hostname"`
	BindAddress        string `json:"bind_address"`
	MetricsBindAddress string `json:"metrics_bind_address"`
}

func (c *ServerConfig) Validate() error {
	if c.BindAddress == "" {
		return fmt.Errorf("bind_address is required")
	}
	return nil
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{}
}
func (c *ServerConfig) LoadEnvVariables() {
	if val, ok := EnvServerHostname.Lookup(); ok {
		c.Hostname = val
	}

	if val, ok := EnvServerBindAddress.Lookup(); ok {
		c.BindAddress = val
	}
	if val, ok := EnvMetricsBindAddress.Lookup(); ok {
		c.MetricsBindAddress = val
	}
}

func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Dialect: "postgres",
	}
}

func (c *DatabaseConfig) LoadEnvVariables() {
	if val, ok := EnvDBSSLMode.Lookup(); ok {
		if val == string(DBSSLModeDisable) || val == string(DBSSLModeRequire) {
			c.SSLMode = SSLMode(val)
		}
	}

	if val, ok := EnvDBMaxConnections.Lookup(); ok {
		c.MaxOpenConnections = val
	}

	if val, ok := EnvDBHost.Lookup(); ok {
		c.Host = val
	}

	if val, ok := EnvDBPort.Lookup(); ok {
		c.Port = val
	}

	if val, ok := EnvDBName.Lookup(); ok {
		c.Name = val
	}

	if val, ok := EnvDBUsername.Lookup(); ok {
		c.Username = val
	}

	if val, ok := EnvDBPassword.Lookup(); ok {
		c.Password = val
	}

	if val, ok := EnvDBRootCertFile.Lookup(); ok {
		c.RootCertFile = val
	}

	if val, ok := EnvDBDebugMode.Lookup(); ok {
		c.Debug = val
	}
}

func (c *DatabaseConfig) ConnectionString() string {
	if c.SSLMode == DBSSLModeRequire {
		return c.ConnectionStringWithName(c.Name, true)
	}
	return c.ConnectionStringWithName(c.Name, false)
}

func (c *DatabaseConfig) ConnectionStringWithName(name string, withSSL bool) string {
	var cmd string
	if withSSL {
		cmd = fmt.Sprintf(
			"host=%s port=%d user=%s password='%s' dbname=%s sslmode=%s sslrootcert=%s",
			c.Host, c.Port, c.Username, c.Password, name, c.SSLMode, c.RootCertFile,
		)
	} else {
		cmd = fmt.Sprintf(
			"host=%s port=%d user=%s password='%s' dbname=%s sslmode=disable",
			c.Host, c.Port, c.Username, c.Password, name,
		)
	}

	return cmd
}

func (c *DatabaseConfig) LogSafeConnectionString(withSSL bool) string {
	return c.LogSafeConnectionStringWithName(c.Name, withSSL)
}

func (c *DatabaseConfig) LogSafeConnectionStringWithName(name string, withSSL bool) string {
	if withSSL {
		return fmt.Sprintf(
			"host=%s port=%d user=%s password='<REDACTED>' dbname=%s sslmode=%s sslrootcert='<REDACTED>'",
			c.Host, c.Port, c.Username, name, c.SSLMode,
		)
	} else {
		return fmt.Sprintf(
			"host=%s port=%d user=%s password='<REDACTED>' dbname=%s",
			c.Host, c.Port, c.Username, name,
		)
	}
}
