package config

var (
	// Application
	EnvJWTSecret     = StringVar("JWT_SECRET", "JWT token signing secret", nil, true)
	EnvLogLevel      = StringVar("LOG_LEVEL", "Logging level (debug, info, warn, error)", ptr("info"), false)
	EnvEncryptionKey = StringVar("ENCRYPTION_KEY", "Master encryption key (64-1024 chars)", nil, true)
	// GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET are used for GitHub OAuth login
	EnvGitHubClientID     = StringVar("GITHUB_CLIENT_ID", "GitHub OAuth app client ID", nil, false)
	EnvGitHubClientSecret = StringVar("GITHUB_CLIENT_SECRET", "GitHub OAuth app client secret", nil, false)
	EnvGitHubRedirectURI  = StringVar("GITHUB_REDIRECT_URI", "GitHub OAuth app redirect URI", nil, false)

	// Server
	EnvServerHostname    = StringVar("SERVER_HOSTNAME", "Server hostname/domain", nil, false)
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

	// Default Cluster
	EnvDefaultClusterName   = StringVar("DEFAULT_CLUSTER_NAME", "Default cluster name", nil, false)
	EnvDefaultClusterAPIURL = StringVar("DEFAULT_CLUSTER_API_URL", "Default cluster API URL", nil, false)
	EnvDefaultClusterCAData = StringVar("DEFAULT_CLUSTER_CA_DATA", "Default cluster CA cert (base64)", nil, false)
	EnvDefaultClusterToken  = StringVar("DEFAULT_CLUSTER_TOKEN", "Default cluster auth token", nil, false)

	// Bootstrap User
	EnvDefaultUserEmail    = StringVar("DEFAULT_USER_EMAIL", "Default platform admin email", nil, false)
	EnvDefaultUserName     = StringVar("DEFAULT_USER_NAME", "Default platform admin name", ptr("Platform Admin"), false)
	EnvDefaultUserPassword = StringVar("DEFAULT_USER_PASSWORD", "Default platform admin password", ptr("welcome@123"), false)

	// Environment
	EnvStackdomeEnv = StringVar("STACKDOME_ENV", "Runtime environment (DEVELOPMENT, PRODUCTION, TESTING)", ptr("DEVELOPMENT"), false)

	// Test Overrides
	EnvTestJWTSecret     = StringVar("TEST_JWT_SECRET", "Override JWT secret in tests", nil, false)
	EnvTestEncryptionKey = StringVar("TEST_ENCRYPTION_KEY", "Override encryption key in tests", nil, false)
	EnvTestLogLevel      = StringVar("TEST_LOG_LEVEL", "Override log level in tests", ptr("info"), false)
)
