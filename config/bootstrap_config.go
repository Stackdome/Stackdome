package config

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
		DefaultUser: &DefaultPlatformAdminConfig{},
	}
}

func (b *BootstrapConfig) LoadEnvVariables() {
	if val, ok := EnvDefaultUserEmail.Lookup(); ok {
		b.DefaultUser.Email = val
	}

	if val, ok := EnvDefaultUserName.Lookup(); ok {
		b.DefaultUser.Name = val
	}

	if val, ok := EnvDefaultUserPassword.Lookup(); ok {
		b.DefaultUser.Password = val
	}
}
