package config

import "github.com/Stackdome/stackdome/pkg/models"

var (
	ErrIncompleteSharedComputeClusterConfig = &ConfigError{"SHARED_COMPUTE_CLUSTER_API_URL, SHARED_COMPUTE_CLUSTER_CA_DATA and SHARED_COMPUTE_CLUSTER_TOKEN must all be set together"}
	ErrUnsupportedComputeMode               = &ConfigError{"COMPUTE_MODE must be bring_your_own or shared"}
	ErrSharedComputeProvisioningRequired    = &ConfigError{"shared compute provisioning is required in shared compute mode"}
	ErrSharedComputeProvisioningNotAllowed  = &ConfigError{"shared compute provisioning is not allowed in bring_your_own compute mode"}
	ErrPlatformRoutingNotAllowed            = &ConfigError{"platform routing is not allowed in bring_your_own compute mode"}
	ErrPlatformBaseDomainRequired           = &ConfigError{"PLATFORM_BASE_DOMAIN is required in shared compute mode"}
	ErrPlatformTLSRequired                  = &ConfigError{"PLATFORM_TLS_ENABLED is required in stackdome_cloud runtime mode"}
	ErrPlatformTLSConfigNotAllowed          = &ConfigError{"platform TLS configuration requires PLATFORM_TLS_ENABLED=true"}
	ErrPlatformEmailRequired                = &ConfigError{"PLATFORM_EMAIL is required when platform TLS is enabled"}
	ErrPlatformCloudflareTokenRequired      = &ConfigError{"PLATFORM_DNS_CLOUDFLARE_API_TOKEN is required when platform TLS is enabled"}
	ErrPlatformACMEEnvironmentInvalid       = &ConfigError{"PLATFORM_ACME_ENVIRONMENT must be production or staging"}
	ErrPlatformTLSNamespaceRequired         = &ConfigError{"PLATFORM_TLS_NAMESPACE is required when platform TLS is enabled"}
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
	PlatformTLSEnabled    bool
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

func ValidateSharedComputeProvisioning(mode ComputeMode, cluster *ClusterConfig) error {
	if !cluster.IsSet() && cluster.AnySet() {
		return ErrIncompleteSharedComputeClusterConfig
	}
	switch mode {
	case ComputeModeBYOC:
		if cluster.AnySet() {
			return ErrSharedComputeProvisioningNotAllowed
		}
		return nil
	case ComputeModeShared:
		if !cluster.IsSet() {
			return ErrSharedComputeProvisioningRequired
		}
		return cluster.Validate()
	default:
		return ErrUnsupportedComputeMode
	}
}

func (b *BootstrapConfig) anyPlatformTLSConfigSet() bool {
	return b.Email != "" || b.DNSCloudflareAPIToken != "" || b.ACMEEnvironment != "" || b.TLSNamespace != ""
}

func ValidatePlatformRouting(runtime RuntimeMode, mode ComputeMode, bootstrap *BootstrapConfig) error {
	switch mode {
	case ComputeModeBYOC:
		if bootstrap.BaseDomain != "" || bootstrap.PlatformTLSEnabled || bootstrap.anyPlatformTLSConfigSet() {
			return ErrPlatformRoutingNotAllowed
		}
	case ComputeModeShared:
		if bootstrap.BaseDomain == "" {
			return ErrPlatformBaseDomainRequired
		}
	default:
		return ErrUnsupportedComputeMode
	}
	if runtime == RuntimeModeStackdomeCloud && !bootstrap.PlatformTLSEnabled {
		return ErrPlatformTLSRequired
	}
	if !bootstrap.PlatformTLSEnabled {
		if bootstrap.anyPlatformTLSConfigSet() {
			return ErrPlatformTLSConfigNotAllowed
		}
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
	return nil
}

func NewBootstrapConfig() *BootstrapConfig {
	return &BootstrapConfig{}
}

func (b *BootstrapConfig) LoadEnvVariables() error {
	if val, ok := EnvPlatformEmail.Lookup(); ok {
		b.Email = val
	}

	if val, ok := EnvPlatformBaseDomain.Lookup(); ok {
		b.BaseDomain = val
	}

	if val, ok := EnvPlatformDNSCloudflareAPIToken.Lookup(); ok {
		b.DNSCloudflareAPIToken = val
	}
	if val, ok := EnvPlatformTLSEnabled.Lookup(); ok {
		b.PlatformTLSEnabled = val
	}

	if val, ok := EnvPlatformACMEEnvironment.Lookup(); ok {
		b.ACMEEnvironment = val
	} else if b.PlatformTLSEnabled {
		b.ACMEEnvironment = ACMEEnvironmentProduction
	}

	if val, ok := EnvPlatformTLSNamespace.Lookup(); ok {
		b.TLSNamespace = val
	} else if b.PlatformTLSEnabled {
		b.TLSNamespace = DefaultPlatformTLSNamespace
	}

	if val, ok := EnvPlatformOrgRegistryStorageSize.Lookup(); ok {
		b.OrgRegistry.StorageSize = val
	}

	if val, ok := EnvPlatformOrgRegistryStorageClass.Lookup(); ok {
		b.OrgRegistry.StorageClass = val
	}

	return nil
}
