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
	Server        *ServerConfig      `json:"server"`
	Database      *DatabaseConfig    `json:"database"`
	JwtSecret     string             `json:"jwt_secret"`
	EncryptionKey string             `json:"encryption_key"`
	LogLevel      string             `json:"log_level"`
	LogFormat     string             `json:"log_format"`
	GitHubOAuth   *GitHubOAuthConfig `json:"github_oauth"`
	// ServerExternalURL is the externally reachable base URL of the hub,
	// required for the GitHub App manifest flow (browser redirects, webhooks).
	ServerExternalURL string `json:"server_external_url"`
	// GitHubAPIBaseURL overrides the GitHub API endpoint (tests, GHES).
	GitHubAPIBaseURL string `json:"github_api_base_url"`
}

func (c *ApplicationConfig) LoadEnvVariables() {
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

	c.GitHubOAuth.LoadEnvVariables()

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
}

// gitHubOAuthCallbackPath is the route the GitHub OAuth handler is mounted on;
// it must match the "/github/callback" route under /api/v1/auth in
// cmd/server/routes.go.
const gitHubOAuthCallbackPath = "/api/v1/auth/github/callback"

func (c *ApplicationConfig) Validate() error {
	validateFuncs := []func() error{
		c.Server.Validate,
		c.Database.Validate,
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

type ClusterConfig struct {
	Name          string `yaml:"name"`
	ClusterURL    string `yaml:"cluster_url"`
	ClusterCAData string `yaml:"cluster_ca_data"`
	Token         string `yaml:"token"`
}

func (c *ClusterConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("cluster name is required")
	}

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
	if val, ok := EnvDefaultClusterName.Lookup(); ok {
		c.Name = val
	}

	if val, ok := EnvDefaultClusterAPIURL.Lookup(); ok {
		c.ClusterURL = val
	}

	if val, ok := EnvDefaultClusterCAData.Lookup(); ok {
		c.ClusterCAData = val
	}

	if val, ok := EnvDefaultClusterToken.Lookup(); ok {
		c.Token = val
	}
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
		Server:      NewServerConfig(),
		Database:    NewDatabaseConfig(),
		LogLevel:    "info",
		LogFormat:   "json",
		GitHubOAuth: NewGitHubOAuthConfig(),
	}
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
	Hostname    string `json:"hostname"`
	BindAddress string `json:"bind_address"`
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
