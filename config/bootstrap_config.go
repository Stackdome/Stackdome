package config

import (
	"os"
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
	DefaultUser *DefaultPlatformAdminConfig
}

func (b *BootstrapConfig) Validate() error {
	return b.DefaultUser.Validate()
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

func NewBootstrapConfig() *BootstrapConfig {
	return &BootstrapConfig{
		DefaultUser: &DefaultPlatformAdminConfig{
			Name:     "Platform Admin",
			Password: "welcome@123",
		},
	}
}

func (b *BootstrapConfig) LoadEnvVariables() {
	defaultUserEmail, found := os.LookupEnv(DEFAULT_USER_EMAIL)
	if found {
		b.DefaultUser.Email = defaultUserEmail
	}

	defaultUserName, found := os.LookupEnv(DEFAULT_USER_NAME)
	if found {
		b.DefaultUser.Name = defaultUserName
	}

	defaultUserPassword, found := os.LookupEnv(DEFAULT_USER_PASS)
	if found {
		b.DefaultUser.Password = defaultUserPassword
	}
}
