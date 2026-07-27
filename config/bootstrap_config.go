package config

import "github.com/Stackdome/stackdome/pkg/models"

var (
	ErrIncompleteClusterConfig = &ConfigError{"PLATFORM_CLUSTER_API_URL, PLATFORM_CLUSTER_CA_DATA and PLATFORM_CLUSTER_TOKEN must all be set together"}
	ErrClusterDomainMismatch   = &ConfigError{"PLATFORM_CLUSTER_* and PLATFORM_BASE_DOMAIN must be set together"}
	ErrPlatformEmailRequired   = &ConfigError{"PLATFORM_EMAIL is required when a platform cluster is configured"}
)

type ConfigError struct {
	msg string
}

func (e *ConfigError) Error() string {
	return e.msg
}

type BootstrapConfig struct {
	Email       string
	BaseDomain  string
	OrgRegistry models.OrgRegistryDefaults
}

func ValidatePlatformProvisioning(cluster *ClusterConfig, baseDomain, email string) error {
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
	if email == "" {
		return ErrPlatformEmailRequired
	}
	return cluster.Validate()
}

func NewBootstrapConfig() *BootstrapConfig {
	return &BootstrapConfig{}
}

func (b *BootstrapConfig) LoadEnvVariables() {
	if val, ok := EnvPlatformEmail.Lookup(); ok {
		b.Email = val
	}

	if val, ok := EnvPlatformBaseDomain.Lookup(); ok {
		b.BaseDomain = val
	}

	if val, ok := EnvPlatformOrgRegistryStorageSize.Lookup(); ok {
		b.OrgRegistry.StorageSize = val
	}

	if val, ok := EnvPlatformOrgRegistryStorageClass.Lookup(); ok {
		b.OrgRegistry.StorageClass = val
	}
}
