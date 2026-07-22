package config

var (
	ErrEmptyEmail    = &ConfigError{"empty email"}
	ErrEmptyName     = &ConfigError{"empty name"}
	ErrEmptyPassword = &ConfigError{"empty password"}

	ErrIncompleteClusterConfig = &ConfigError{"PLATFORM_CLUSTER_NAME, PLATFORM_CLUSTER_API_URL, PLATFORM_CLUSTER_CA_DATA and PLATFORM_CLUSTER_TOKEN must all be set together"}
	ErrClusterDomainMismatch   = &ConfigError{"PLATFORM_CLUSTER_* and PLATFORM_BASE_DOMAIN must be set together"}
	ErrBootstrapAdminEmail     = &ConfigError{"PLATFORM_ADMIN_EMAIL is required when a platform cluster is configured"}
)

type ConfigError struct {
	msg string
}

func (e *ConfigError) Error() string {
	return e.msg
}

type BootstrapConfig struct {
	PlatformAdmin        *PlatformAdminConfig
	BaseDomain           string
	RegistryStorageSize  string
	RegistryStorageClass string
	RegistryName         string
}

func ValidatePlatformProvisioning(cluster *ClusterConfig, baseDomain, adminEmail string) error {
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
	return b.PlatformAdmin.Validate()
}

type PlatformAdminConfig struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (d *PlatformAdminConfig) Validate() error {
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
		PlatformAdmin: &PlatformAdminConfig{},
	}
}

func (b *BootstrapConfig) LoadEnvVariables() {
	if val, ok := EnvPlatformAdminEmail.Lookup(); ok {
		b.PlatformAdmin.Email = val
	}

	if val, ok := EnvPlatformAdminName.Lookup(); ok {
		b.PlatformAdmin.Name = val
	}

	if val, ok := EnvPlatformAdminPassword.Lookup(); ok {
		b.PlatformAdmin.Password = val
	}

	if val, ok := EnvPlatformBaseDomain.Lookup(); ok {
		b.BaseDomain = val
	}

	if val, ok := EnvPlatformRegistryStorageSize.Lookup(); ok {
		b.RegistryStorageSize = val
	}

	if val, ok := EnvPlatformRegistryStorageClass.Lookup(); ok {
		b.RegistryStorageClass = val
	}

	if val, ok := EnvPlatformClusterRegistryName.Lookup(); ok {
		b.RegistryName = val
	}
}
