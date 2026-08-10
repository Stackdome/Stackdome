package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation"
)

type RuntimeMode string

const (
	RuntimeModeSelfHosted     RuntimeMode = "self_hosted"
	RuntimeModeStackdomeCloud RuntimeMode = "stackdome_cloud"
)

type ConfigDuration time.Duration

func (d *ConfigDuration) UnmarshalYAML(value *yaml.Node) error {
	parsed, err := time.ParseDuration(value.Value)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", value.Value, err)
	}
	*d = ConfigDuration(parsed)
	return nil
}

func (d ConfigDuration) Duration() time.Duration {
	return time.Duration(d)
}

type StackdomeCloudConfig struct {
	Capacity StackdomeCloudCapacityConfig `yaml:"capacity" json:"capacity"`
	Limits   StackdomeCloudLimitsConfig   `yaml:"limits" json:"limits"`
	Registry StackdomeCloudRegistryConfig `yaml:"registry" json:"registry"`
	Features StackdomeCloudFeaturesConfig `yaml:"features" json:"features"`
	Signup   StackdomeCloudSignupConfig   `yaml:"signup" json:"signup"`
}

type StackdomeCloudCapacityConfig struct {
	MaxActiveTrialAllocations int            `yaml:"maxActiveTrialAllocations" json:"max_active_trial_allocations"`
	AllocationTTL             ConfigDuration `yaml:"allocationTTL" json:"allocation_ttl"`
}

type StackdomeCloudLimitsConfig struct {
	MaxStacksPerOrganization         int64 `yaml:"maxStacksPerOrganization" json:"max_stacks_per_organization"`
	MaxStackResourcesPerOrganization int64 `yaml:"maxStackResourcesPerOrganization" json:"max_stack_resources_per_organization"`
	ReplicasPerStackResource         int32 `yaml:"replicasPerStackResource" json:"replicas_per_stack_resource"`
	ConcurrentBuilds                 int   `yaml:"concurrentBuilds" json:"concurrent_builds"`
}

type StackdomeCloudRegistryConfig struct {
	MaxActiveRegistries int    `yaml:"maxActiveRegistries" json:"max_active_registries"`
	StorageClass        string `yaml:"storageClass" json:"storage_class"`
	StorageSize         string `yaml:"storageSize" json:"storage_size"`
}

type StackdomeCloudFeaturesConfig struct {
	CustomDomains          bool `yaml:"customDomains" json:"custom_domains"`
	ExternalPostgresImport bool `yaml:"externalPostgresImport" json:"external_postgres_import"`
	// WorkspaceUsers is accepted only as false for compatibility with the first cloud config.
	WorkspaceUsers bool `yaml:"workspaceUsers" json:"workspace_users"`
}

type StackdomeCloudClientIPSource string

const (
	StackdomeCloudClientIPSourceCloudflare StackdomeCloudClientIPSource = "cloudflare"
	StackdomeCloudClientIPSourceRemoteAddr StackdomeCloudClientIPSource = "remote_addr"
)

type StackdomeCloudSignupConfig struct {
	ClientIPSource StackdomeCloudClientIPSource  `yaml:"clientIPSource" json:"client_ip_source"`
	Turnstile      StackdomeCloudTurnstileConfig `yaml:"turnstile" json:"turnstile"`
	// Throttle is always enforced in the Stackdome Cloud runtime.
	Throttle StackdomeCloudThrottleConfig `yaml:"throttle" json:"throttle"`
}

type StackdomeCloudTurnstileConfig struct {
	Enabled             bool           `yaml:"enabled" json:"enabled"`
	SiteKey             string         `yaml:"siteKey" json:"site_key"`
	ExpectedHostname    string         `yaml:"expectedHostname" json:"expected_hostname"`
	ExpectedAction      string         `yaml:"expectedAction" json:"expected_action"`
	VerificationTimeout ConfigDuration `yaml:"verificationTimeout" json:"verification_timeout"`
}

type StackdomeCloudThrottleConfig struct {
	IP    StackdomeCloudIPThrottleConfig    `yaml:"ip" json:"ip"`
	Email StackdomeCloudEmailThrottleConfig `yaml:"email" json:"email"`
}

type StackdomeCloudIPThrottleConfig struct {
	MaxTrackedClients int            `yaml:"maxTrackedClients" json:"max_tracked_clients"`
	MaxAttempts       int            `yaml:"maxAttempts" json:"max_attempts"`
	Window            ConfigDuration `yaml:"window" json:"window"`
}

type StackdomeCloudEmailThrottleConfig struct {
	MaxTrackedAddresses int            `yaml:"maxTrackedAddresses" json:"max_tracked_addresses"`
	MaxAttempts         int            `yaml:"maxAttempts" json:"max_attempts"`
	Window              ConfigDuration `yaml:"window" json:"window"`
}

func (c *StackdomeCloudConfig) Validate() error {
	if c.Capacity.MaxActiveTrialAllocations <= 0 {
		return fmt.Errorf("capacity.maxActiveTrialAllocations must be greater than zero")
	}
	if c.Capacity.AllocationTTL.Duration() <= 0 {
		return fmt.Errorf("capacity.allocationTTL must be greater than zero")
	}
	if c.Limits.MaxStacksPerOrganization <= 0 {
		return fmt.Errorf("limits.maxStacksPerOrganization must be greater than zero")
	}
	if c.Limits.MaxStackResourcesPerOrganization <= 0 {
		return fmt.Errorf("limits.maxStackResourcesPerOrganization must be greater than zero")
	}
	if c.Limits.ReplicasPerStackResource <= 0 {
		return fmt.Errorf("limits.replicasPerStackResource must be greater than zero")
	}
	if c.Limits.ConcurrentBuilds <= 0 {
		return fmt.Errorf("limits.concurrentBuilds must be greater than zero")
	}
	if c.Registry.MaxActiveRegistries <= 0 {
		return fmt.Errorf("registry.maxActiveRegistries must be greater than zero")
	}
	if c.Registry.StorageClass == "" {
		return fmt.Errorf("registry.storageClass is required")
	}
	if problems := validation.IsDNS1123Subdomain(c.Registry.StorageClass); len(problems) > 0 {
		return fmt.Errorf("registry.storageClass must be a valid Kubernetes name: %s", problems[0])
	}
	if c.Registry.StorageSize == "" {
		return fmt.Errorf("registry.storageSize is required")
	}
	if c.Features.WorkspaceUsers {
		return fmt.Errorf("features.workspaceUsers has been removed and must be false")
	}
	storageSize, err := resource.ParseQuantity(c.Registry.StorageSize)
	if err != nil {
		return fmt.Errorf("registry.storageSize must be a valid Kubernetes quantity: %w", err)
	}
	if storageSize.Sign() <= 0 {
		return fmt.Errorf("registry.storageSize must be greater than zero")
	}
	switch c.Signup.ClientIPSource {
	case StackdomeCloudClientIPSourceCloudflare, StackdomeCloudClientIPSourceRemoteAddr:
	default:
		return fmt.Errorf("signup.clientIPSource must be %q or %q", StackdomeCloudClientIPSourceCloudflare, StackdomeCloudClientIPSourceRemoteAddr)
	}
	if !c.Signup.Turnstile.Enabled {
		return fmt.Errorf("signup.turnstile.enabled must be true")
	}
	if c.Signup.Turnstile.SiteKey == "" {
		return fmt.Errorf("signup.turnstile.siteKey is required")
	}
	if c.Signup.Turnstile.ExpectedHostname == "" {
		return fmt.Errorf("signup.turnstile.expectedHostname is required")
	}
	if c.Signup.Turnstile.ExpectedAction == "" {
		return fmt.Errorf("signup.turnstile.expectedAction is required")
	}
	if c.Signup.Turnstile.VerificationTimeout.Duration() <= 0 {
		return fmt.Errorf("signup.turnstile.verificationTimeout must be greater than zero")
	}
	if c.Signup.Throttle.IP.MaxTrackedClients <= 0 {
		return fmt.Errorf("signup.throttle.ip.maxTrackedClients must be greater than zero")
	}
	if c.Signup.Throttle.IP.MaxAttempts <= 0 {
		return fmt.Errorf("signup.throttle.ip.maxAttempts must be greater than zero")
	}
	if c.Signup.Throttle.IP.Window.Duration() <= 0 {
		return fmt.Errorf("signup.throttle.ip.window must be greater than zero")
	}
	if c.Signup.Throttle.Email.MaxTrackedAddresses <= 0 {
		return fmt.Errorf("signup.throttle.email.maxTrackedAddresses must be greater than zero")
	}
	if c.Signup.Throttle.Email.MaxAttempts <= 0 {
		return fmt.Errorf("signup.throttle.email.maxAttempts must be greater than zero")
	}
	if c.Signup.Throttle.Email.Window.Duration() <= 0 {
		return fmt.Errorf("signup.throttle.email.window must be greater than zero")
	}
	return nil
}

func LoadStackdomeCloudConfig(path string) (*StackdomeCloudConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("open Stackdome Cloud config: %w", err)
	}

	var cloudConfig StackdomeCloudConfig
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cloudConfig); err != nil {
		return nil, fmt.Errorf("decode Stackdome Cloud config: %w", err)
	}
	var trailingDocument any
	if err := decoder.Decode(&trailingDocument); !errors.Is(err, io.EOF) {
		if err != nil {
			return nil, fmt.Errorf("decode trailing Stackdome Cloud config: %w", err)
		}
		return nil, fmt.Errorf("multiple YAML documents are not allowed")
	}
	if err := cloudConfig.Validate(); err != nil {
		return nil, fmt.Errorf("validate Stackdome Cloud config: %w", err)
	}
	return &cloudConfig, nil
}
