package config

import "github.com/Stackdome/stackdome/pkg/models"

var (
	// Application
	EnvJWTSecret     = StringVar("JWT_SECRET", "JWT token signing secret", nil, true)
	EnvLogLevel      = StringVar("LOG_LEVEL", "Logging level (debug, info, warn, error)", ptr("info"), false)
	EnvLogFormat     = StringVar("LOG_FORMAT", "Log output format (json or text)", ptr("json"), false)
	EnvEncryptionKey = StringVar("ENCRYPTION_KEY", "Master encryption key (64-1024 chars)", nil, true)
	// GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET are used for GitHub OAuth login
	EnvGitHubClientID     = StringVar("GITHUB_CLIENT_ID", "GitHub OAuth app client ID", nil, false)
	EnvGitHubClientSecret = StringVar("GITHUB_CLIENT_SECRET", "GitHub OAuth app client secret", nil, false)
	EnvGitHubRedirectURI  = StringVar("GITHUB_REDIRECT_URI", "GitHub OAuth app redirect URI (optional; derived from SERVER_EXTERNAL_URL when unset)", nil, false)
	// The GITHUB_APP_* vars configure one platform-wide GitHub App that every
	// org installs; unset leaves each org creating its own app via the
	// manifest flow.
	EnvGitHubAppID            = IntVar("GITHUB_APP_ID", "Platform GitHub App numeric ID", nil, false)
	EnvGitHubAppSlug          = StringVar("GITHUB_APP_SLUG", "Platform GitHub App URL slug", nil, false)
	EnvGitHubAppPrivateKey    = StringVar("GITHUB_APP_PRIVATE_KEY", "Platform GitHub App private key (PEM)", nil, false)
	EnvGitHubAppWebhookSecret = StringVar("GITHUB_APP_WEBHOOK_SECRET", "Platform GitHub App webhook secret", nil, false)

	// Server
	EnvServerHostname    = StringVar("SERVER_HOSTNAME", "Server hostname/domain", nil, false)
	EnvServerExternalURL = StringVar("SERVER_EXTERNAL_URL", "Externally reachable base URL of the hub (used for GitHub App callbacks and webhooks)", nil, false)
	// GitHub API base URL override (httptest stubs / GHES)
	EnvGitHubAPIBaseURL  = StringVar("GITHUB_API_BASE_URL", "GitHub API base URL", ptr("https://api.github.com"), false)
	EnvServerBindAddress = StringVar("SERVER_BIND_ADDRESS", "Server bind address and port", ptr("0.0.0.0:8000"), false)

	// Database
	EnvDBHost           = StringVar("DB_HOST", "PostgreSQL host", nil, true)
	EnvDBPort           = IntVar("DB_PORT", "PostgreSQL port", nil, true)
	EnvDBName           = StringVar("DB_NAME", "Database name", nil, true)
	EnvDBUsername       = StringVar("DB_USERNAME", "Database user", nil, true)
	EnvDBPassword       = StringVar("DB_PASSWORD", "Database password", nil, true)
	EnvDBSSLMode        = StringVar("DB_SSL_MODE", "SSL mode (disable or require)", ptr("disable"), false)
	EnvDBMaxConnections = IntVar("DB_MAX_CONNECTIONS", "Max open DB connections", ptr(50), false)
	EnvDBRootCertFile   = StringVar("DB_ROOT_CERT_FILE", "Root CA cert path for SSL", nil, false)
	EnvDBDebugMode      = BoolVar("DB_DEBUG_MODE", "Enable DB query debug logging", ptr(false), false)

	// Platform Cluster
	EnvPlatformClusterAPIURL = StringVar("PLATFORM_CLUSTER_API_URL", "Platform cluster API URL", nil, false)
	EnvPlatformClusterCAData = StringVar("PLATFORM_CLUSTER_CA_DATA", "Platform cluster CA cert (base64)", nil, false)
	EnvPlatformClusterToken  = StringVar("PLATFORM_CLUSTER_TOKEN", "Platform cluster auth token", nil, false)

	// Platform Provisioning
	EnvPlatformEmail                   = StringVar("PLATFORM_EMAIL", "Operator contact email; ACME contact for the platform cluster issuer", nil, false)
	EnvPlatformBaseDomain              = StringVar("PLATFORM_BASE_DOMAIN", "Base domain for the platform org and per-org subdomains", nil, false)
	EnvPlatformDNSCloudflareAPIToken   = StringVar("PLATFORM_DNS_CLOUDFLARE_API_TOKEN", "Cloudflare API token used for platform wildcard DNS challenges", nil, false)
	EnvPlatformACMEEnvironment         = StringVar("PLATFORM_ACME_ENVIRONMENT", "Let's Encrypt ACME environment (production or staging)", ptr(ACMEEnvironmentProduction), false)
	EnvPlatformTLSNamespace            = StringVar("PLATFORM_TLS_NAMESPACE", "Namespace for platform TLS resources", ptr(DefaultPlatformTLSNamespace), false)
	EnvPlatformOrgRegistryStorageSize  = StringVar("PLATFORM_ORG_REGISTRY_STORAGE_SIZE", "Default storage size for each org's registry seeded at signup", ptr(models.DefaultPlatformOrgRegistryStorageSize), false)
	EnvPlatformOrgRegistryStorageClass = StringVar("PLATFORM_ORG_REGISTRY_STORAGE_CLASS", "Default storage class for each org's registry seeded at signup", nil, false)

	// Environment
	EnvStackdomeEnv = StringVar("STACKDOME_ENV", "Runtime environment (DEVELOPMENT, PRODUCTION)", ptr("DEVELOPMENT"), false)

	// Test Overrides
	EnvTestJWTSecret     = StringVar("TEST_JWT_SECRET", "Override JWT secret in tests", nil, false)
	EnvTestEncryptionKey = StringVar("TEST_ENCRYPTION_KEY", "Override encryption key in tests", nil, false)
	EnvTestLogLevel      = StringVar("TEST_LOG_LEVEL", "Override log level in tests", ptr("info"), false)
)
