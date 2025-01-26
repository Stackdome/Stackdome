package config

import (
	"flag"
	"os"

	"github.com/spf13/pflag"
)

const (
	DefaultUserEmailEnv    = "DEFAULT_USER_EMAIL"
	DefaultUserNameEnv     = "DEFAULT_USER_NAME"
	DefaultUserPasswordEnv = "DEFAULT_USER_PASSWORD"
)

var (
	ErrEmptyEmail    = &ConfigError{"empty email"}
	ErrEmptyName     = &ConfigError{"empty name"}
	ErrEmptyPassword = &ConfigError{"empty password"}
)

type ConfigError struct {
	msg string
}

func (e *ConfigError) Error() string {
	return e.msg
}

type BootstrapConfig struct {
	// Created during migration.
	// DefaultOrganisation *DefaultOrganisationConfig
	DefaultUser *DefaultPlatformAdminConfig
}

type DefaultPlatformAdminConfig struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (d *DefaultPlatformAdminConfig) Validate() error {
	if d.Email == "" {
		return ErrEmptyEmail
	}
	if d.Name == "" {
		return ErrEmptyName
	}
	if d.Password == "" {
		return ErrEmptyPassword
	}
	return nil
}

type DefaultOrganisationConfig struct {
	Name       string
	DomainName string
}

func NewBootstrapConfig() *BootstrapConfig {
	return &BootstrapConfig{
		DefaultUser: &DefaultPlatformAdminConfig{
			Name:     "Platform Admin",
			Password: "welcome@123",
		},
	}
}

func (b *BootstrapConfig) AddFlags(fs *pflag.FlagSet) {
	fs.AddGoFlagSet(flag.CommandLine)
	fs.StringVar(&b.DefaultUser.Email, "default-user-email", b.DefaultUser.Email, "Default user email")
	fs.StringVar(&b.DefaultUser.Name, "default-user-name", b.DefaultUser.Name, "Default user name")
	fs.StringVar(&b.DefaultUser.Password, "default-user-password", b.DefaultUser.Password, "Default user password")
}

func (b *BootstrapConfig) ReadEnvironmentVariables() {
	defaultUserEmail, found := os.LookupEnv(DefaultUserEmailEnv)
	if found {
		b.DefaultUser.Email = defaultUserEmail
	}

	defaultUserName, found := os.LookupEnv(DefaultUserNameEnv)
	if found {
		b.DefaultUser.Name = defaultUserName
	}

	defaultUserPassword, found := os.LookupEnv(DefaultUserPasswordEnv)
	if found {
		b.DefaultUser.Password = defaultUserPassword
	}
}
