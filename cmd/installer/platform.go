package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/install"
	k8svalidation "k8s.io/apimachinery/pkg/util/validation"
)

type platformFlags struct {
	baseDomain          *string
	cloudflareTokenFile *string
	acmeEnvironment     *string
}

func registerPlatformFlags(fs *flag.FlagSet) *platformFlags {
	return &platformFlags{
		baseDomain: fs.String(
			"platform-base-domain",
			"",
			"Base domain for automatically allocated platform hostnames",
		),
		cloudflareTokenFile: fs.String(
			"platform-cloudflare-token-file",
			"",
			"File containing the Cloudflare API token for wildcard certificate DNS challenges",
		),
		acmeEnvironment: fs.String(
			"platform-acme-environment",
			"",
			"Let's Encrypt environment: production or staging",
		),
	}
}

// resolvePlatformConfig applies explicitly supplied flags to the configuration
// already stored by the installer. Base domain and ACME environment are stable
// after first configuration; only the Cloudflare token can be rotated.
func (f *platformFlags) resolvePlatformConfig(stored install.PlatformConfig) (install.PlatformConfig, error) {
	resolved := stored
	requestedBaseDomain := strings.TrimSpace(*f.baseDomain)
	requestedACMEEnvironment := strings.TrimSpace(*f.acmeEnvironment)
	if requestedBaseDomain != "" {
		if err := validatePlatformBaseDomain(requestedBaseDomain); err != nil {
			return install.PlatformConfig{}, err
		}
	}

	if stored.Enabled() {
		if requestedBaseDomain != "" && requestedBaseDomain != stored.BaseDomain {
			return install.PlatformConfig{}, fmt.Errorf("platform base domain is already configured as %q", stored.BaseDomain)
		}
		if requestedACMEEnvironment != "" && requestedACMEEnvironment != stored.ACMEEnvironment {
			return install.PlatformConfig{}, fmt.Errorf("platform ACME environment is already configured as %q", stored.ACMEEnvironment)
		}
	} else {
		resolved.BaseDomain = requestedBaseDomain
		resolved.ACMEEnvironment = requestedACMEEnvironment
	}

	if *f.cloudflareTokenFile != "" {
		token, err := os.ReadFile(*f.cloudflareTokenFile)
		if err != nil {
			return install.PlatformConfig{}, fmt.Errorf("reading platform Cloudflare API token: %w", err)
		}
		resolved.CloudflareAPIToken = strings.TrimSpace(string(token))
	}

	if !resolved.Enabled() {
		if resolved.CloudflareAPIToken != "" || resolved.ACMEEnvironment != "" {
			return install.PlatformConfig{}, fmt.Errorf("--platform-base-domain is required with platform TLS configuration")
		}
		return resolved, nil
	}
	if err := validatePlatformBaseDomain(resolved.BaseDomain); err != nil {
		return install.PlatformConfig{}, err
	}
	if resolved.CloudflareAPIToken == "" {
		return install.PlatformConfig{}, fmt.Errorf("--platform-cloudflare-token-file is required when enabling platform TLS")
	}
	if resolved.ACMEEnvironment == "" {
		resolved.ACMEEnvironment = config.ACMEEnvironmentProduction
	}
	if resolved.ACMEEnvironment != config.ACMEEnvironmentProduction && resolved.ACMEEnvironment != config.ACMEEnvironmentStaging {
		return install.PlatformConfig{}, fmt.Errorf("--platform-acme-environment must be production or staging")
	}
	return resolved, nil
}

func validatePlatformBaseDomain(baseDomain string) error {
	if !strings.Contains(baseDomain, ".") {
		return fmt.Errorf("platform base domain %q must contain at least two DNS labels", baseDomain)
	}
	if problems := k8svalidation.IsDNS1123Subdomain(baseDomain); len(problems) != 0 {
		return fmt.Errorf("invalid platform base domain %q: %s", baseDomain, strings.Join(problems, "; "))
	}
	return nil
}

func ensurePlatformClusterCredentials(platform *install.PlatformConfig) error {
	if !platform.Enabled() || (platform.ClusterAPIURL != "" && platform.ClusterCAData != "" && platform.ClusterToken != "") {
		return nil
	}

	clusterURL, caData, token, err := extractClusterCredentials()
	if err != nil {
		return fmt.Errorf("extracting platform cluster credentials: %w", err)
	}
	platform.ClusterAPIURL = clusterURL
	platform.ClusterCAData = caData
	platform.ClusterToken = token
	return nil
}
