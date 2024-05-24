package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

const (
	JWT_SECRET_ENV string = "JWT_KEY"
)

type ApplicationConfig struct {
	Server   *ServerConfig   `json:"server"`
	Database *DatabaseConfig `json:"database"`
}

type DatabaseConfig struct {
	Dialect            string `json:"dialect"`
	SSLMode            string `json:"sslmode"`
	RootCertFile       string
	Debug              bool `json:"debug"`
	MaxOpenConnections int  `json:"max_connections"`

	DatabaseConfigFile string `json:"database_config_file"`
	DBConnectionConfig
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
	}
}

func (c *ApplicationConfig) AddFlags(flagset *pflag.FlagSet) {
	flagset.AddGoFlagSet(flag.CommandLine)
	c.Server.AddFlags(flagset)
	c.Database.AddFlags(flagset)
}

type ServerConfig struct {
	Hostname          string `json:"hostname"`
	BindAddress       string `json:"bind_address"`
	JwtSecret         string `json:"jwt_secret"`
	JwtSecretFilePath string `json:"jwt_secret_file_path"`
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		Hostname:    "",
		BindAddress: "localhost:8000",
	}
}

func (s *ServerConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.BindAddress, "api-server-bindaddress", s.BindAddress, "API server bind adddress")
	fs.StringVar(&s.Hostname, "api-server-hostname", s.Hostname, "Server's public hostname")
	fs.StringVar(&s.JwtSecretFilePath, "jwt-secret-file", s.JwtSecretFilePath, "File containing the jwt secret")
}

func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Dialect:            "postgres",
		SSLMode:            "disable",
		Debug:              false,
		MaxOpenConnections: 50,
	}
}

func (c *DatabaseConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.DatabaseConfigFile, "db-config-file", c.DatabaseConfigFile, "Database root certificate file")
	fs.StringVar(&c.SSLMode, "db-sslmode", c.SSLMode, "Database ssl mode (disable | require | verify-ca | verify-full)")
	fs.BoolVar(&c.Debug, "enable-db-debug", c.Debug, " framework's debug mode")
	fs.IntVar(&c.MaxOpenConnections, "db-max-open-connections", c.MaxOpenConnections, "Maximum open DB connections for this instance")
}

func (c *DatabaseConfig) ConnectionString(withSSL bool) string {
	return c.ConnectionStringWithName(c.Name, withSSL)
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

func (a *ApplicationConfig) ReadConfigFiles() error {
	data, err := os.ReadFile(a.Database.DatabaseConfigFile)
	if err != nil {
		return fmt.Errorf("error reading YAML file: %v", err)
	}

	var config DBConnectionConfig
	err = yaml.UnmarshalStrict(data, &config)
	if err != nil {
		return fmt.Errorf("error unmarshalling DB config YAML: %v", err)
	}
	a.Database.DBConnectionConfig = config

	if len(a.Server.JwtSecret) == 0 {
		data, err := os.ReadFile(a.Server.JwtSecretFilePath)
		if err != nil {
			return fmt.Errorf("failed to read jwt secret from file '%s'. Error: %w", a.Server.JwtSecretFilePath, err)
		}
		a.Server.JwtSecret = string(data)
	}
	return nil
}

func (a *ApplicationConfig) ReadEnvironmentVariables() {
	secret, found := os.LookupEnv(JWT_SECRET_ENV)
	if found {
		a.Server.JwtSecret = secret
	}
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
