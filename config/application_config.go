package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type SSLMode string

const (
	DBSSLModeDisable SSLMode = "disable"
	DBSSLModeRequire SSLMode = "require"
)

type ApplicationConfig struct {
	Server        *ServerConfig   `json:"server"`
	Database      *DatabaseConfig `json:"database"`
	JwtSecret     string          `json:"jwt_secret"`
	EncryptionKey string          `json:"encryption_key"`
	LogLevel      string          `json:"log_level"`
}

func (c *ApplicationConfig) LoadEnvVariables() {
	c.Server.LoadEnvVariables()
	c.Database.LoadEnvVariables()

	val, found := os.LookupEnv(JWT_SECRET)
	if found {
		c.JwtSecret = strings.TrimSpace(val)
	}

	val, found = os.LookupEnv(LOG_LEVEL)
	if found {
		c.LogLevel = strings.TrimSpace(val)
	}

	val, found = os.LookupEnv(ENCRYPTION_KEY)
	if found {
		c.EncryptionKey = strings.TrimSpace(val)
	}
}

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
	val, found := os.LookupEnv(DEFAULT_CLUSTER_NAME)
	if found {
		c.Name = val
	}

	val, found = os.LookupEnv(DEFAULT_CLUSTER_API_URL)
	if found {
		c.ClusterURL = val
	}

	val, found = os.LookupEnv(DEFAULT_CLUSTER_CA_DATA)
	if found {
		c.ClusterCAData = val
	}

	val, found = os.LookupEnv(DEFAULT_CLUSTER_TOKEN)
	if found {
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
		Server:   NewServerConfig(),
		Database: NewDatabaseConfig(),
		LogLevel: "info",
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
	return &ServerConfig{
		Hostname:    "",
		BindAddress: "0.0.0.0:8000",
	}
}
func (c *ServerConfig) LoadEnvVariables() {
	val, found := os.LookupEnv(SERVER_HOSTNAME)
	if found {
		c.Hostname = val
	}

	val, found = os.LookupEnv(SERVER_BIND_ADDRESS)
	if found {
		c.BindAddress = val
	}
}

func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Dialect:            "postgres",
		SSLMode:            "disable",
		Debug:              false,
		MaxOpenConnections: 50,
	}
}

func (c *DatabaseConfig) LoadEnvVariables() {
	val, found := os.LookupEnv(DB_SSL_MODE)
	if found {
		if val == string(DBSSLModeDisable) || val == string(DBSSLModeRequire) {
			c.SSLMode = SSLMode(val)
		}
	}

	val, found = os.LookupEnv(DB_MAX_CONNECTIONS)
	if found {
		// convert string to int
		maxConns, err := strconv.Atoi(val)
		if err == nil {
			c.MaxOpenConnections = maxConns
		}
		// TODO: log error
	}

	val, found = os.LookupEnv(DB_HOST)
	if found {
		c.Host = val
	}

	val, found = os.LookupEnv(DB_PORT)
	if found {
		// convert string to int
		port, err := strconv.Atoi(val)
		if err == nil {
			c.Port = port
		}
		// TODO: log error
	}

	val, found = os.LookupEnv(DB_NAME)
	if found {
		c.Name = val
	}

	val, found = os.LookupEnv(DB_USERNAME)
	if found {
		c.Username = val
	}

	val, found = os.LookupEnv(DB_PASSWORD)
	if found {
		c.Password = val
	}

	val, found = os.LookupEnv(DB_ROOT_CERT_FILE)
	if found {
		c.RootCertFile = val
	}

	val, found = os.LookupEnv(DB_DEBUG_MODE)
	if found {
		// convert string to bool
		debugMode, err := strconv.ParseBool(val)
		if err == nil {
			c.Debug = debugMode
		}
		// TODO: log error
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
