package config

import "github.com/Stackdome/stackdome/pkg/models"

var (
	ErrIncompleteClusterConfig         = &ConfigError{"PLATFORM_CLUSTER_API_URL, PLATFORM_CLUSTER_CA_DATA and PLATFORM_CLUSTER_TOKEN must all be set together"}
	ErrClusterDomainMismatch           = &ConfigError{"PLATFORM_CLUSTER_* and PLATFORM_BASE_DOMAIN must be set together"}
	ErrPlatformEmailRequired           = &ConfigError{"PLATFORM_EMAIL is required when a platform cluster is configured"}
	ErrPlatformCloudflareTokenRequired = &ConfigError{"PLATFORM_DNS_CLOUDFLARE_API_TOKEN is required when a platform cluster is configured"}
	ErrPlatformACMEEnvironmentInvalid  = &ConfigError{"PLATFORM_ACME_ENVIRONMENT must be production or staging"}
	ErrPlatformTLSNamespaceRequired    = &ConfigError{"PLATFORM_TLS_NAMESPACE is required when a platform cluster is configured"}
)

const (
	ACMEEnvironmentProduction   = "production"
	ACMEEnvironmentStaging      = "staging"
	DefaultPlatformTLSNamespace = "stackdome-control-plane"
	ACMEProductionDirectoryURL  = "https://acme-v02.api.letsencrypt.org/directory"
	ACMEStagingDirectoryURL     = "https://acme-staging-v02.api.letsencrypt.org/directory"
)

type ConfigError struct {
	msg string
}

func (e *ConfigError) Error() string {
	return e.msg
}

type BootstrapConfig struct {
	Email                 string
	BaseDomain            string
	DNSCloudflareAPIToken string
	ACMEEnvironment       string
	TLSNamespace          string
	OrgRegistry           models.OrgRegistryDefaults
}

func (b *BootstrapConfig) ACMEDirectoryURL() string {
	switch b.ACMEEnvironment {
	case ACMEEnvironmentProduction:
		return ACMEProductionDirectoryURL
	case ACMEEnvironmentStaging:
		return ACMEStagingDirectoryURL
	default:
		return ""
	}
}

func ValidatePlatformProvisioning(cluster *ClusterConfig, bootstrap *BootstrapConfig) error {
	clusterSet := cluster.IsSet()
	if !clusterSet && cluster.AnySet() {
		return ErrIncompleteClusterConfig
	}
	domainSet := bootstrap.BaseDomain != ""
	if clusterSet != domainSet {
		return ErrClusterDomainMismatch
	}
	if !clusterSet {
		return nil
	}
	if bootstrap.Email == "" {
		return ErrPlatformEmailRequired
	}
	if bootstrap.DNSCloudflareAPIToken == "" {
		return ErrPlatformCloudflareTokenRequired
	}
	if bootstrap.ACMEEnvironment != ACMEEnvironmentProduction && bootstrap.ACMEEnvironment != ACMEEnvironmentStaging {
		return ErrPlatformACMEEnvironmentInvalid
	}
	if bootstrap.TLSNamespace == "" {
		return ErrPlatformTLSNamespaceRequired
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

	if val, ok := EnvPlatformDNSCloudflareAPIToken.Lookup(); ok {
		b.DNSCloudflareAPIToken = val
	}

	if val, ok := EnvPlatformACMEEnvironment.Lookup(); ok {
		b.ACMEEnvironment = val
	}

	if val, ok := EnvPlatformTLSNamespace.Lookup(); ok {
		b.TLSNamespace = val
	}

	if val, ok := EnvPlatformOrgRegistryStorageSize.Lookup(); ok {
		b.OrgRegistry.StorageSize = val
	}

	if val, ok := EnvPlatformOrgRegistryStorageClass.Lookup(); ok {
		b.OrgRegistry.StorageClass = val
	}
}
