package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Runtime mode configuration", func() {
	It("defaults to self-hosted", func() {
		Expect(os.Unsetenv(EnvRuntimeMode.Name)).To(Succeed())

		cfg := NewApplicationConfig()
		Expect(cfg.LoadEnvVariables()).To(Succeed())

		Expect(cfg.RuntimeMode).To(Equal(RuntimeModeSelfHosted))
		Expect(cfg.IsStackdomeCloud()).To(BeFalse())
	})

	It("requires a config path in Stackdome Cloud mode", func() {
		GinkgoT().Setenv(EnvRuntimeMode.Name, string(RuntimeModeStackdomeCloud))
		Expect(os.Unsetenv(EnvStackdomeCloudConfig.Name)).To(Succeed())

		cfg := NewApplicationConfig()
		Expect(cfg.LoadEnvVariables()).To(Succeed())

		Expect(cfg.LoadStackdomeCloudConfig()).To(MatchError(ContainSubstring("STACKDOME_CLOUD_CONFIG is required")))
	})

	It("rejects a cloud config path in self-hosted mode", func() {
		cfg := NewApplicationConfig()
		cfg.StackdomeCloudConfigPath = "/etc/stackdome/cloud.yaml"

		Expect(cfg.LoadStackdomeCloudConfig()).To(MatchError(ContainSubstring("STACKDOME_CLOUD_CONFIG requires RUNTIME_MODE=stackdome_cloud")))
	})

	It("rejects a Turnstile secret in self-hosted mode", func() {
		cfg := NewApplicationConfig()
		cfg.TurnstileSecret = "configured-secret"

		Expect(cfg.LoadStackdomeCloudConfig()).To(MatchError(ContainSubstring("TURNSTILE_SECRET requires RUNTIME_MODE=stackdome_cloud")))
	})

	It("loads and validates the typed Stackdome Cloud YAML", func() {
		path := filepath.Join(GinkgoT().TempDir(), "cloud.yaml")
		Expect(os.WriteFile(path, []byte(`
capacity:
  maxActiveTrialAllocations: 200
  allocationTTL: 6h
limits:
  maxStacksPerOrganization: 2
  maxStackResourcesPerOrganization: 6
  replicasPerStackResource: 1
  concurrentBuilds: 1
registry:
  maxActiveRegistries: 200
  storageClass: longhorn
  storageSize: 10Gi
features:
  customDomains: false
  externalPostgresImport: false
  workspaceUsers: false
signup:
  turnstile:
    enabled: false
  throttle:
    maxTrackedKeys: 10000
    ipAttempts: 5
    ipWindow: 1h
    emailAttempts: 3
    emailWindow: 1h
`), 0o600)).To(Succeed())
		GinkgoT().Setenv(EnvRuntimeMode.Name, string(RuntimeModeStackdomeCloud))
		GinkgoT().Setenv(EnvStackdomeCloudConfig.Name, path)

		cfg := NewApplicationConfig()
		Expect(cfg.LoadEnvVariables()).To(Succeed())
		Expect(cfg.LoadStackdomeCloudConfig()).To(Succeed())

		Expect(cfg.IsStackdomeCloud()).To(BeTrue())
		Expect(cfg.StackdomeCloud.Capacity.MaxActiveTrialAllocations).To(Equal(200))
		Expect(cfg.StackdomeCloud.Capacity.AllocationTTL.Duration()).To(Equal(6 * time.Hour))
		Expect(cfg.StackdomeCloud.Limits.MaxStackResourcesPerOrganization).To(Equal(int64(6)))
		Expect(cfg.StackdomeCloud.Registry.MaxActiveRegistries).To(Equal(200))
		Expect(cfg.StackdomeCloud.Registry.StorageClass).To(Equal("longhorn"))
		Expect(cfg.StackdomeCloud.Registry.StorageSize).To(Equal("10Gi"))
	})

	It("rejects unknown cloud configuration fields", func() {
		path := filepath.Join(GinkgoT().TempDir(), "cloud.yaml")
		Expect(os.WriteFile(path, []byte("unexpected: true\n"), 0o600)).To(Succeed())
		_, err := LoadStackdomeCloudConfig(path)
		Expect(err).To(MatchError(ContainSubstring("field unexpected not found")))
	})

	It("rejects trailing YAML documents", func() {
		path := filepath.Join(GinkgoT().TempDir(), "cloud.yaml")
		validConfig, err := os.ReadFile("stackdome_cloud.example.yaml")
		Expect(err).NotTo(HaveOccurred())
		validConfig = append(validConfig, []byte("\n---\nfeatures:\n  workspaceUsers: true\n")...)
		Expect(os.WriteFile(path, validConfig, 0o600)).To(Succeed())

		_, err = LoadStackdomeCloudConfig(path)
		Expect(err).To(MatchError(ContainSubstring("multiple YAML documents are not allowed")))
	})

	It("parses the checked-in Stackdome Cloud example", func() {
		cloudConfig, err := LoadStackdomeCloudConfig("stackdome_cloud.example.yaml")
		Expect(err).NotTo(HaveOccurred())
		Expect(cloudConfig.Capacity.MaxActiveTrialAllocations).To(Equal(200))
		Expect(cloudConfig.Registry.MaxActiveRegistries).To(Equal(200))
		Expect(cloudConfig.Signup.Throttle.IPWindow.Duration()).To(Equal(10 * time.Minute))
		Expect(cloudConfig.Signup.Turnstile.Enabled).To(BeTrue())
	})

	It("requires the Turnstile secret env value while loading enabled cloud signup", func() {
		cfg := NewApplicationConfig()
		cfg.RuntimeMode = RuntimeModeStackdomeCloud
		cfg.StackdomeCloudConfigPath = "stackdome_cloud.example.yaml"

		Expect(cfg.LoadStackdomeCloudConfig()).To(MatchError(ContainSubstring("TURNSTILE_SECRET is required")))
	})

	It("requires Turnstile public fields when Turnstile is enabled", func() {
		cloudConfig := validStackdomeCloudConfigForTest()
		cloudConfig.Signup.Turnstile.Enabled = true

		Expect(cloudConfig.Validate()).To(MatchError(ContainSubstring("signup.turnstile.siteKey is required")))
	})

	It("always requires positive signup throttle limits in Stackdome Cloud", func() {
		cloudConfig := validStackdomeCloudConfigForTest()
		cloudConfig.Signup.Throttle.IPAttempts = 0

		Expect(cloudConfig.Validate()).To(MatchError(ContainSubstring("signup.throttle.ipAttempts must be greater than zero")))
	})

	It("requires a positive registry capacity", func() {
		cloudConfig := validStackdomeCloudConfigForTest()
		cloudConfig.Registry.MaxActiveRegistries = 0

		Expect(cloudConfig.Validate()).To(MatchError(ContainSubstring("registry.maxActiveRegistries must be greater than zero")))
	})

	It("requires registry storage configuration", func() {
		cloudConfig := validStackdomeCloudConfigForTest()
		cloudConfig.Registry.StorageClass = ""

		Expect(cloudConfig.Validate()).To(MatchError(ContainSubstring("registry.storageClass is required")))
	})

	It("rejects an invalid registry storage class", func() {
		cloudConfig := validStackdomeCloudConfigForTest()
		cloudConfig.Registry.StorageClass = "Not A Kubernetes Name"

		Expect(cloudConfig.Validate()).To(MatchError(ContainSubstring("registry.storageClass must be a valid Kubernetes name")))
	})

	It("rejects an invalid registry storage quantity", func() {
		cloudConfig := validStackdomeCloudConfigForTest()
		cloudConfig.Registry.StorageSize = "ten gigabytes"

		Expect(cloudConfig.Validate()).To(MatchError(ContainSubstring("registry.storageSize must be a valid Kubernetes quantity")))
	})

	It("requires a positive registry storage quantity", func() {
		cloudConfig := validStackdomeCloudConfigForTest()
		cloudConfig.Registry.StorageSize = "0"

		Expect(cloudConfig.Validate()).To(MatchError(ContainSubstring("registry.storageSize must be greater than zero")))
	})
})

var _ = Describe("Compute mode configuration", func() {
	It("defaults to bring-your-own compute", func() {
		Expect(os.Unsetenv(EnvComputeMode.Name)).To(Succeed())

		cfg := NewApplicationConfig()
		Expect(cfg.LoadEnvVariables()).To(Succeed())

		Expect(cfg.ComputeMode).To(Equal(ComputeModeBYOC))
		Expect(cfg.UsesSharedCompute()).To(BeFalse())
	})

	DescribeTable("accepts supported compute modes",
		func(mode ComputeMode, usesSharedCompute bool) {
			GinkgoT().Setenv(EnvComputeMode.Name, string(mode))
			GinkgoT().Setenv(EnvRuntimeMode.Name, string(RuntimeModeSelfHosted))

			cfg := validApplicationConfigForTest()
			Expect(cfg.LoadEnvVariables()).To(Succeed())

			Expect(cfg.ComputeMode).To(Equal(mode))
			Expect(cfg.Validate()).To(Succeed())
			Expect(cfg.UsesSharedCompute()).To(Equal(usesSharedCompute))
		},
		Entry("bring-your-own", ComputeModeBYOC, false),
		Entry("shared", ComputeModeShared, true),
	)

	It("rejects an unknown compute mode", func() {
		GinkgoT().Setenv(EnvComputeMode.Name, "dedicated")
		GinkgoT().Setenv(EnvRuntimeMode.Name, string(RuntimeModeSelfHosted))

		cfg := validApplicationConfigForTest()
		Expect(cfg.LoadEnvVariables()).To(Succeed())

		Expect(cfg.Validate()).To(MatchError(
			`compute mode must be "bring_your_own" or "shared"`,
		))
	})

	It("requires shared compute in Stackdome Cloud", func() {
		cloudConfig := validStackdomeCloudConfigForTest()
		cfg := validApplicationConfigForTest()
		cfg.RuntimeMode = RuntimeModeStackdomeCloud
		cfg.ComputeMode = ComputeModeBYOC
		cfg.StackdomeCloud = &cloudConfig

		Expect(cfg.Validate()).To(MatchError(
			`compute mode must be "shared" in "stackdome_cloud" runtime mode`,
		))
	})

	It("validates an injected Stackdome Cloud configuration", func() {
		cfg := validApplicationConfigForTest()
		cfg.RuntimeMode = RuntimeModeStackdomeCloud
		cfg.ComputeMode = ComputeModeShared
		cfg.StackdomeCloud = &StackdomeCloudConfig{}

		Expect(cfg.Validate()).To(MatchError(
			"validate Stackdome Cloud config: capacity.maxActiveTrialAllocations must be greater than zero",
		))
	})
})

func validStackdomeCloudConfigForTest() StackdomeCloudConfig {
	return StackdomeCloudConfig{
		Capacity: StackdomeCloudCapacityConfig{
			MaxActiveTrialAllocations: 200,
			AllocationTTL:             ConfigDuration(6 * time.Hour),
		},
		Limits: StackdomeCloudLimitsConfig{
			MaxStacksPerOrganization:         2,
			MaxStackResourcesPerOrganization: 6,
			ReplicasPerStackResource:         1,
			ConcurrentBuilds:                 1,
		},
		Registry: StackdomeCloudRegistryConfig{
			MaxActiveRegistries: 200,
			StorageClass:        "longhorn",
			StorageSize:         "10Gi",
		},
		Signup: StackdomeCloudSignupConfig{
			Throttle: StackdomeCloudThrottleConfig{
				MaxTrackedKeys: 10_000,
				IPAttempts:     5,
				IPWindow:       ConfigDuration(time.Hour),
				EmailAttempts:  3,
				EmailWindow:    ConfigDuration(time.Hour),
			},
		},
	}
}

func validApplicationConfigForTest() *ApplicationConfig {
	cfg := NewApplicationConfig()
	cfg.Server.BindAddress = "127.0.0.1:8000"
	cfg.Database.SSLMode = DBSSLModeDisable
	cfg.Database.MaxOpenConnections = 10
	cfg.Database.Host = "localhost"
	cfg.Database.Port = 5432
	cfg.Database.Name = "stackdome"
	cfg.Database.Username = "stackdome"
	cfg.Database.Password = "password"
	cfg.JwtSecret = "jwt-secret"
	cfg.EncryptionKey = "1234567890123456789012345678901234567890123456789012345678901234"
	return cfg
}

func TestGitHubRedirectURIDerivedFromExternalURL(t *testing.T) {
	t.Setenv("SERVER_EXTERNAL_URL", "https://hub.example.com")
	t.Setenv("GITHUB_REDIRECT_URI", "")

	c := NewApplicationConfig()
	if err := c.LoadEnvVariables(); err != nil {
		t.Fatalf("load environment variables: %v", err)
	}

	want := "https://hub.example.com/auth/github/callback"
	if c.GitHubOAuth.RedirectURI != want {
		t.Fatalf("expected derived redirect %q, got %q", want, c.GitHubOAuth.RedirectURI)
	}
}

func TestGitHubRedirectURITrimsTrailingSlash(t *testing.T) {
	t.Setenv("SERVER_EXTERNAL_URL", "https://hub.example.com/")
	t.Setenv("GITHUB_REDIRECT_URI", "")

	c := NewApplicationConfig()
	if err := c.LoadEnvVariables(); err != nil {
		t.Fatalf("load environment variables: %v", err)
	}

	want := "https://hub.example.com/auth/github/callback"
	if c.GitHubOAuth.RedirectURI != want {
		t.Fatalf("expected no double slash, got %q", c.GitHubOAuth.RedirectURI)
	}
}

func TestGitHubRedirectURIExplicitOverrideWins(t *testing.T) {
	t.Setenv("SERVER_EXTERNAL_URL", "https://hub.example.com")
	t.Setenv("GITHUB_REDIRECT_URI", "https://custom.example.com/callback")

	c := NewApplicationConfig()
	if err := c.LoadEnvVariables(); err != nil {
		t.Fatalf("load environment variables: %v", err)
	}

	if c.GitHubOAuth.RedirectURI != "https://custom.example.com/callback" {
		t.Fatalf("explicit override should win, got %q", c.GitHubOAuth.RedirectURI)
	}
}
