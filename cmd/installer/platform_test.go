package main

import (
	"flag"
	"strings"
	"testing"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/install"
)

func parsePlatformFlags(t *testing.T, args ...string) *platformFlags {
	t.Helper()
	fs := flag.NewFlagSet("platform-test", flag.ContinueOnError)
	flags := registerPlatformFlags(fs)
	if err := fs.Parse(args); err != nil {
		t.Fatalf("parse platform flags: %v", err)
	}
	return flags
}

func TestResolvePlatformConfigPreservesStoredTLSWhenFlagOmitted(t *testing.T) {
	stored := install.PlatformConfig{
		BaseDomain:         "apps.example.com",
		TLSEnabled:         true,
		CloudflareAPIToken: "cloudflare-token",
		ACMEEnvironment:    config.ACMEEnvironmentStaging,
	}

	got, err := parsePlatformFlags(t).resolvePlatformConfig(stored)
	if err != nil {
		t.Fatalf("resolve platform config: %v", err)
	}
	if got != stored {
		t.Fatalf("resolved config = %#v, want %#v", got, stored)
	}
}

func TestResolvePlatformConfigAllowsSharedComputeWithoutTLS(t *testing.T) {
	got, err := parsePlatformFlags(t,
		"--platform-base-domain", "apps.example.com",
	).resolvePlatformConfig(install.PlatformConfig{})
	if err != nil {
		t.Fatalf("resolve platform config: %v", err)
	}
	if !got.Enabled() {
		t.Fatal("platform routing is disabled")
	}
	if got.TLSEnabled {
		t.Fatal("platform TLS is enabled")
	}
	if got.CloudflareAPIToken != "" || got.ACMEEnvironment != "" {
		t.Fatalf("TLS-only config was populated: %#v", got)
	}
}

func TestResolvePlatformConfigRequiresCloudflareTokenWhenTLSEnabled(t *testing.T) {
	_, err := parsePlatformFlags(t,
		"--platform-base-domain", "apps.example.com",
		"--platform-tls",
	).resolvePlatformConfig(install.PlatformConfig{})
	if err == nil || !strings.Contains(err.Error(), "--platform-cloudflare-token-file is required") {
		t.Fatalf("error = %v, want missing Cloudflare token error", err)
	}
}

func TestResolvePlatformConfigRejectsTLSOnlyFlagsWhenDisabled(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "Cloudflare token",
			args: []string{"--platform-base-domain", "apps.example.com", "--platform-cloudflare-token-file", "token.txt"},
		},
		{
			name: "ACME environment",
			args: []string{"--platform-base-domain", "apps.example.com", "--platform-acme-environment", "staging"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parsePlatformFlags(t, tt.args...).resolvePlatformConfig(install.PlatformConfig{})
			if err == nil || !strings.Contains(err.Error(), "require --platform-tls=true") {
				t.Fatalf("error = %v, want TLS-disabled validation error", err)
			}
		})
	}
}
