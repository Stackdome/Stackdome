package main

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Installer bootstrap secrets", func() {
	DescribeTable("preserves the stored TLS state during upgrade",
		func(storedTLS *string, expected bool) {
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
			if storedTLS != nil {
				data["platform-tls-enabled"] = encodeSecretValue(*storedTLS)
			}
			installFakeKubectl(data)

			secrets, err := readExistingSecrets()
			Expect(err).NotTo(HaveOccurred())
			flags, err := parsePlatformFlags()
			Expect(err).NotTo(HaveOccurred())
			resolved, err := flags.resolvePlatformConfig(secrets.Platform)
			Expect(err).NotTo(HaveOccurred())
			Expect(resolved.TLSEnabled).To(Equal(expected))
		},
		Entry("legacy secret without TLS key", nil, true),
		Entry("explicitly disabled TLS", stringPtr("false"), false),
	)
})

func encodeSecretValue(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

func installFakeKubectl(data map[string]string) {
	payload, err := json.Marshal(map[string]any{"data": data})
	Expect(err).NotTo(HaveOccurred())

	binDir := GinkgoT().TempDir()
	kubectlPath := filepath.Join(binDir, "kubectl")
	script := "#!/bin/sh\nprintf '%s' '" + string(payload) + "'\n"
	Expect(os.WriteFile(kubectlPath, []byte(script), 0o755)).To(Succeed())
	GinkgoT().Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func stringPtr(value string) *string {
	return &value
}
