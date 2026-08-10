package main

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestUpgradePreservesPlatformTLSStateFromBootstrapSecret(t *testing.T) {
	tests := []struct {
		name           string
		storedTLS      *string
		wantTLSEnabled bool
	}{
		{
			name:           "legacy secret without TLS key",
			wantTLSEnabled: true,
		},
		{
			name:           "explicitly disabled TLS",
			storedTLS:      stringPtr("false"),
			wantTLSEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]string{
				"db-password":                    encodeSecretValue("db-password"),
				"jwt-secret":                     encodeSecretValue("jwt-secret"),
				"encryption-key":                 encodeSecretValue("encryption-key"),
				"admin-password":                 encodeSecretValue("admin-password"),
				"platform-base-domain":           encodeSecretValue("apps.example.com"),
				"platform-cloudflare-api-token":  encodeSecretValue("cloudflare-token"),
				"platform-acme-environment":      encodeSecretValue("production"),
				"shared-compute-cluster-api-url": encodeSecretValue("https://10.0.0.1:443"),
				"shared-compute-cluster-ca-data": encodeSecretValue("ca-data"),
				"shared-compute-cluster-token":   encodeSecretValue("cluster-token"),
			}
			if tt.storedTLS != nil {
				data["platform-tls-enabled"] = encodeSecretValue(*tt.storedTLS)
			}
			installFakeKubectl(t, data)

			secrets, err := readExistingSecrets()
			if err != nil {
				t.Fatalf("read existing secrets: %v", err)
			}
			resolved, err := parsePlatformFlags(t).resolvePlatformConfig(secrets.Platform)
			if err != nil {
				t.Fatalf("resolve platform config: %v", err)
			}
			if resolved.TLSEnabled != tt.wantTLSEnabled {
				t.Fatalf("TLS enabled = %t, want %t", resolved.TLSEnabled, tt.wantTLSEnabled)
			}
		})
	}
}

func encodeSecretValue(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

func installFakeKubectl(t *testing.T, data map[string]string) {
	t.Helper()
	payload, err := json.Marshal(map[string]any{"data": data})
	if err != nil {
		t.Fatalf("marshal Secret fixture: %v", err)
	}

	binDir := t.TempDir()
	kubectlPath := filepath.Join(binDir, "kubectl")
	script := "#!/bin/sh\nprintf '%s' '" + string(payload) + "'\n"
	if err := os.WriteFile(kubectlPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake kubectl: %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func stringPtr(value string) *string {
	return &value
}
