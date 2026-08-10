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

type PlatformConfig struct {
	Email                 string
	BaseDomain            string
	DNSCloudflareAPIToken string
	PlatformTLSEnabled    bool
	ACMEEnvironment       string
	TLSNamespace          string
	OrgRegistry           models.OrgRegistryDefaults
}

func NewPlatformConfig() *PlatformConfig {
	return &PlatformConfig{}
}

func (p *PlatformConfig) LoadEnvVariables() error {
	if val, ok := EnvPlatformEmail.Lookup(); ok {
		p.Email = val
	}
	if val, ok := EnvPlatformBaseDomain.Lookup(); ok {
		p.BaseDomain = val
	}
	if val, ok := EnvPlatformDNSCloudflareAPIToken.Lookup(); ok {
		p.DNSCloudflareAPIToken = val
	}
	if val, ok := EnvPlatformTLSEnabled.Lookup(); ok {
		p.PlatformTLSEnabled = val
	}
	if val, ok := EnvPlatformACMEEnvironment.Lookup(); ok {
		p.ACMEEnvironment = val
	}
	if val, ok := EnvPlatformTLSNamespace.Lookup(); ok {
		p.TLSNamespace = val
	}
	if val, ok := EnvPlatformOrgRegistryStorageSize.Lookup(); ok {
		p.OrgRegistry.StorageSize = val
	}
	if val, ok := EnvPlatformOrgRegistryStorageClass.Lookup(); ok {
		p.OrgRegistry.StorageClass = val
	}

	p.applyTLSDefaults()
	return nil
}

func (p *PlatformConfig) applyTLSDefaults() {
	if !p.PlatformTLSEnabled {
		return
	}
	if p.ACMEEnvironment == "" {
		p.ACMEEnvironment = ACMEEnvironmentProduction
	}
	if p.TLSNamespace == "" {
		p.TLSNamespace = DefaultPlatformTLSNamespace
	}
}

func (p *PlatformConfig) ACMEDirectoryURL() string {
	switch p.ACMEEnvironment {
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

func ValidatePlatformRouting(runtime RuntimeMode, mode ComputeMode, platform *PlatformConfig) error {
	if err := validateRoutingForComputeMode(mode, platform); err != nil {
		return err
	}
	return validatePlatformTLS(runtime, platform)
}

func validateRoutingForComputeMode(mode ComputeMode, platform *PlatformConfig) error {
	switch mode {
	case ComputeModeBYOC:
		if platform.hasRoutingConfig() {
			return ErrPlatformRoutingNotAllowed
		}
		return nil
	case ComputeModeShared:
		if platform.BaseDomain == "" {
			return ErrPlatformBaseDomainRequired
		}
		return nil
	default:
		return ErrUnsupportedComputeMode
	}
}

func validatePlatformTLS(runtime RuntimeMode, platform *PlatformConfig) error {
	if runtime == RuntimeModeStackdomeCloud && !platform.PlatformTLSEnabled {
		return ErrPlatformTLSRequired
	}
	if !platform.PlatformTLSEnabled {
		if platform.anyTLSConfigSet() {
			return ErrPlatformTLSConfigNotAllowed
		}
		return nil
	}
	return platform.validateEnabledTLS()
}

func (p *PlatformConfig) hasRoutingConfig() bool {
	return p.BaseDomain != "" || p.PlatformTLSEnabled || p.anyTLSConfigSet()
}

func (p *PlatformConfig) anyTLSConfigSet() bool {
	return p.Email != "" ||
		p.DNSCloudflareAPIToken != "" ||
		p.ACMEEnvironment != "" ||
		p.TLSNamespace != ""
}

func (p *PlatformConfig) validateEnabledTLS() error {
	if p.Email == "" {
		return ErrPlatformEmailRequired
	}
	if p.DNSCloudflareAPIToken == "" {
		return ErrPlatformCloudflareTokenRequired
	}
	if !isValidACMEEnvironment(p.ACMEEnvironment) {
		return ErrPlatformACMEEnvironmentInvalid
	}
	if p.TLSNamespace == "" {
		return ErrPlatformTLSNamespaceRequired
	}
	return nil
}

func isValidACMEEnvironment(env string) bool {
	return env == ACMEEnvironmentProduction || env == ACMEEnvironmentStaging
}
