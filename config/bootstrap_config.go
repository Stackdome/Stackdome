package config

var (
	ErrEmptyEmail    = &ConfigError{"empty email"}
	ErrEmptyName     = &ConfigError{"empty name"}
	ErrEmptyPassword = &ConfigError{"empty password"}

	ErrIncompleteClusterConfig = &ConfigError{"DEFAULT_CLUSTER_NAME, DEFAULT_CLUSTER_API_URL, DEFAULT_CLUSTER_CA_DATA and DEFAULT_CLUSTER_TOKEN must all be set together"}
	ErrClusterDomainMismatch   = &ConfigError{"DEFAULT_CLUSTER_* and DEFAULT_BASE_DOMAIN must be set together"}
	ErrBootstrapAdminEmail     = &ConfigError{"DEFAULT_USER_EMAIL is required when a default cluster is configured"}
)

type ConfigError struct {
	msg string
}

func (e *ConfigError) Error() string {
	return e.msg
}

type BootstrapConfig struct {
	DefaultUser          *DefaultPlatformAdminConfig
	PlatformOrgName      string
	BaseDomain           string
	RegistryStorageSize  string
	RegistryStorageClass string
	RegistryName         string
}

func ValidateDefaultProvisioning(cluster *ClusterConfig, baseDomain, adminEmail string) error {
	clusterSet := cluster.IsSet()
	if !clusterSet && cluster.AnySet() {
		return ErrIncompleteClusterConfig
	}
	domainSet := baseDomain != ""
	if clusterSet != domainSet {
		return ErrClusterDomainMismatch
	}
	if !clusterSet {
		return nil
	}
	if adminEmail == "" {
		return ErrBootstrapAdminEmail
	}
	return cluster.Validate()
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

	if val, ok := EnvDefaultOrgName.Lookup(); ok {
		b.PlatformOrgName = val
	}

	if val, ok := EnvDefaultBaseDomain.Lookup(); ok {
		b.BaseDomain = val
	}

	if val, ok := EnvDefaultRegistryStorageSize.Lookup(); ok {
		b.RegistryStorageSize = val
	}

	if val, ok := EnvDefaultRegistryStorageClass.Lookup(); ok {
		b.RegistryStorageClass = val
	}

	if val, ok := EnvDefaultClusterRegistryName.Lookup(); ok {
		b.RegistryName = val
	}
}
