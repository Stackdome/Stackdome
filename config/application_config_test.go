package config

import (
	"os"
	"path/filepath"
	"strings"
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
  clientIPSource: cloudflare
  turnstile:
    enabled: true
    siteKey: test-site-key
    expectedHostname: stackdome.example.com
    expectedAction: turnstile-spin-v2
    verificationTimeout: 10s
  throttle:
    ip:
      maxTrackedClients: 10000
      maxAttempts: 5
      window: 1h
    email:
      maxTrackedAddresses: 10000
      maxAttempts: 3
      window: 1h
`), 0o600)).To(Succeed())
		GinkgoT().Setenv(EnvRuntimeMode.Name, string(RuntimeModeStackdomeCloud))
		GinkgoT().Setenv(EnvStackdomeCloudConfig.Name, path)
		GinkgoT().Setenv(EnvTurnstileSecret.Name, "test-secret")

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
		Expect(cloudConfig.Signup.ClientIPSource).To(Equal(StackdomeCloudClientIPSourceCloudflare))
		Expect(cloudConfig.Signup.Throttle.IP.MaxTrackedClients).To(Equal(10_000))
		Expect(cloudConfig.Signup.Throttle.IP.Window.Duration()).To(Equal(10 * time.Minute))
		Expect(cloudConfig.Signup.Throttle.Email.MaxTrackedAddresses).To(Equal(10_000))
		Expect(cloudConfig.Signup.Throttle.Email.Window.Duration()).To(Equal(time.Hour))
		Expect(cloudConfig.Signup.Turnstile.Enabled).To(BeTrue())
		Expect(cloudConfig.Signup.Turnstile.VerificationTimeout.Duration()).To(Equal(10 * time.Second))
	})

	It("requires the Turnstile secret env value while loading enabled cloud signup", func() {
		cfg := NewApplicationConfig()
		cfg.RuntimeMode = RuntimeModeStackdomeCloud
		cfg.StackdomeCloudConfigPath = "stackdome_cloud.example.yaml"

		Expect(cfg.LoadStackdomeCloudConfig()).To(MatchError(ContainSubstring("TURNSTILE_SECRET is required")))
	})

	It("requires Turnstile to be enabled", func() {
		cloudConfig := validStackdomeCloudConfigForTest()
		cloudConfig.Signup.Turnstile.Enabled = false

		Expect(cloudConfig.Validate()).To(MatchError(ContainSubstring("signup.turnstile.enabled must be true")))
	})

	DescribeTable("validates the client IP source",
		func(source StackdomeCloudClientIPSource, expectedError string) {
			cloudConfig := validStackdomeCloudConfigForTest()
			cloudConfig.Signup.ClientIPSource = source

			err := cloudConfig.Validate()
			if expectedError == "" {
				Expect(err).NotTo(HaveOccurred())
				return
			}
			Expect(err).To(MatchError(expectedError))
		},
		Entry("Cloudflare", StackdomeCloudClientIPSourceCloudflare, ""),
		Entry("direct peer", StackdomeCloudClientIPSourceRemoteAddr, ""),
		Entry("empty", StackdomeCloudClientIPSource(""), `signup.clientIPSource must be "cloudflare" or "remote_addr"`),
		Entry("unknown", StackdomeCloudClientIPSource("forwarded_for"), `signup.clientIPSource must be "cloudflare" or "remote_addr"`),
	)

	DescribeTable("requires complete Turnstile configuration",
		func(mutate func(*StackdomeCloudTurnstileConfig), expectedError string) {
			cloudConfig := validStackdomeCloudConfigForTest()
			mutate(&cloudConfig.Signup.Turnstile)

			Expect(cloudConfig.Validate()).To(MatchError(expectedError))
		},
		Entry("site key", func(c *StackdomeCloudTurnstileConfig) { c.SiteKey = "" }, "signup.turnstile.siteKey is required"),
		Entry("hostname", func(c *StackdomeCloudTurnstileConfig) { c.ExpectedHostname = "" }, "signup.turnstile.expectedHostname is required"),
		Entry("action", func(c *StackdomeCloudTurnstileConfig) { c.ExpectedAction = "" }, "signup.turnstile.expectedAction is required"),
		Entry("verification timeout", func(c *StackdomeCloudTurnstileConfig) { c.VerificationTimeout = 0 }, "signup.turnstile.verificationTimeout must be greater than zero"),
	)

	DescribeTable("requires positive IP throttle limits",
		func(mutate func(*StackdomeCloudIPThrottleConfig), expectedError string) {
			cloudConfig := validStackdomeCloudConfigForTest()
			mutate(&cloudConfig.Signup.Throttle.IP)

			Expect(cloudConfig.Validate()).To(MatchError(expectedError))
		},
		Entry("tracked clients", func(c *StackdomeCloudIPThrottleConfig) { c.MaxTrackedClients = 0 }, "signup.throttle.ip.maxTrackedClients must be greater than zero"),
		Entry("attempts", func(c *StackdomeCloudIPThrottleConfig) { c.MaxAttempts = 0 }, "signup.throttle.ip.maxAttempts must be greater than zero"),
		Entry("window", func(c *StackdomeCloudIPThrottleConfig) { c.Window = 0 }, "signup.throttle.ip.window must be greater than zero"),
	)

	DescribeTable("requires positive email throttle limits",
		func(mutate func(*StackdomeCloudEmailThrottleConfig), expectedError string) {
			cloudConfig := validStackdomeCloudConfigForTest()
			mutate(&cloudConfig.Signup.Throttle.Email)

			Expect(cloudConfig.Validate()).To(MatchError(expectedError))
		},
		Entry("tracked addresses", func(c *StackdomeCloudEmailThrottleConfig) { c.MaxTrackedAddresses = 0 }, "signup.throttle.email.maxTrackedAddresses must be greater than zero"),
		Entry("attempts", func(c *StackdomeCloudEmailThrottleConfig) { c.MaxAttempts = 0 }, "signup.throttle.email.maxAttempts must be greater than zero"),
		Entry("window", func(c *StackdomeCloudEmailThrottleConfig) { c.Window = 0 }, "signup.throttle.email.window must be greater than zero"),
	)

	DescribeTable("rejects legacy signup fields",
		func(existing, legacyField string) {
			path := filepath.Join(GinkgoT().TempDir(), "cloud.yaml")
			raw, err := os.ReadFile("stackdome_cloud.example.yaml")
			Expect(err).NotTo(HaveOccurred())
			configText := string(raw)
			Expect(configText).To(ContainSubstring(existing))
			raw = []byte(strings.Replace(configText, existing, existing+legacyField, 1))
			Expect(os.WriteFile(path, raw, 0o600)).To(Succeed())

			_, err = LoadStackdomeCloudConfig(path)
			Expect(err).To(MatchError(ContainSubstring("field")))
		},
		Entry("trustCloudflareProxy", "    expectedAction: turnstile-spin-v2\n", "    trustCloudflareProxy: true\n"),
		Entry("flat maxTrackedKeys", "  throttle:\n", "    maxTrackedKeys: 10000\n"),
	)

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

var _ = Describe("Shared compute environment", func() {
	BeforeEach(func() {
		for _, name := range []string{
			EnvSharedComputeClusterAPIURL.Name,
			EnvSharedComputeClusterCAData.Name,
			EnvSharedComputeClusterToken.Name,
		} {
			Expect(os.Unsetenv(name)).To(Succeed())
		}
	})

	It("loads the shared compute cluster variables", func() {
		GinkgoT().Setenv(EnvSharedComputeClusterAPIURL.Name, "https://shared.example.com")
		GinkgoT().Setenv(EnvSharedComputeClusterCAData.Name, "shared-ca")
		GinkgoT().Setenv(EnvSharedComputeClusterToken.Name, "shared-token")

		cfg := NewApplicationConfig()
		Expect(cfg.LoadEnvVariables()).To(Succeed())
		Expect(cfg.SharedComputeCluster).To(Equal(&ClusterConfig{
			ClusterURL:    "https://shared.example.com",
			ClusterCAData: "shared-ca",
			Token:         "shared-token",
		}))
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
			ClientIPSource: StackdomeCloudClientIPSourceCloudflare,
			Turnstile: StackdomeCloudTurnstileConfig{
				Enabled:             true,
				SiteKey:             "test-site-key",
				ExpectedHostname:    "stackdome.example.com",
				ExpectedAction:      "turnstile-spin-v2",
				VerificationTimeout: ConfigDuration(10 * time.Second),
			},
			Throttle: StackdomeCloudThrottleConfig{
				IP: StackdomeCloudIPThrottleConfig{
					MaxTrackedClients: 10_000,
					MaxAttempts:       5,
					Window:            ConfigDuration(time.Hour),
				},
				Email: StackdomeCloudEmailThrottleConfig{
					MaxTrackedAddresses: 10_000,
					MaxAttempts:         3,
					Window:              ConfigDuration(time.Hour),
				},
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
