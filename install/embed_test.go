package install_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	"github.com/Stackdome/stackdome/install"
	"github.com/Stackdome/stackdome/pkg/models"
)

func TestInstall(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Install Manifests Suite")
}

var _ = Describe("BootstrapConfigRevision", func() {
	It("changes with TLS, routing, and shared-compute configuration without exposing credentials", func() {
		platform := install.PlatformConfig{
			BaseDomain:         "apps.example.com",
			CloudflareAPIToken: "cloudflare-token",
			ACMEEnvironment:    "production",
		}
		shared := install.SharedComputeConfig{
			ClusterAPIURL: "https://10.0.0.1:443",
			ClusterCAData: "ca-data",
			ClusterToken:  "cluster-token",
		}

		baseline := install.BootstrapConfigRevision(platform, shared)
		Expect(baseline).To(HaveLen(16))
		Expect(baseline).NotTo(ContainSubstring("cluster-token"))

		platform.TLSEnabled = true
		Expect(install.BootstrapConfigRevision(platform, shared)).NotTo(Equal(baseline))

		platform.TLSEnabled = false
		shared.ClusterToken = "rotated-token"
		Expect(install.BootstrapConfigRevision(platform, shared)).NotTo(Equal(baseline))
	})
})

var _ = Describe("RenderManifest", func() {
	It("labels StackResources for the Stack controller", func() {
		for _, name := range []string{"db-resource-cr.yaml", "api-server-resource-cr.yaml"} {
			out, err := install.RenderManifest(name, install.TemplateValues{})
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).To(ContainSubstring("core.stackdome.io/stack-name: stackdome-platform"))
			Expect(string(out)).NotTo(ContainSubstring("\n    stackdome.io/stack-name:"))
		}
	})

	Describe("db-resource-cr.yaml", func() {
		It("renders the requested workload type", func() {
			out, err := install.RenderManifest("db-resource-cr.yaml", install.TemplateValues{
				DBWorkloadType: string(models.WorkloadTypeStatefulService),
				DBPassword:     "pw",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).To(ContainSubstring("workloadType: " + string(models.WorkloadTypeStatefulService)))
			Expect(string(out)).To(ContainSubstring("value: \"pw\""))
		})

		It("owns the resource by the platform Stack", func() {
			out, err := install.RenderManifest("db-resource-cr.yaml", install.TemplateValues{
				DBWorkloadType: string(models.WorkloadTypeStatefulService),
				StackUID:       "5a3a3a1e-0000-4000-8000-000000000001",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).To(ContainSubstring("kind: Stack"))
			Expect(string(out)).To(ContainSubstring("uid: \"5a3a3a1e-0000-4000-8000-000000000001\""))
		})

		It("preserves an existing workload type on upgrade", func() {
			out, err := install.RenderManifest("db-resource-cr.yaml", install.TemplateValues{
				DBWorkloadType: string(models.WorkloadTypeService),
				DBPassword:     "pw",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).To(ContainSubstring("workloadType: " + string(models.WorkloadTypeService)))
		})
	})

	Describe("api-server-resource-cr.yaml", func() {
		It("renders the compute mode from platform availability", func() {
			type renderedEnvironmentVariable struct {
				Name  string `yaml:"name"`
				Value string `yaml:"value"`
			}
			var manifest struct {
				Spec struct {
					EnvironmentVariables []renderedEnvironmentVariable `yaml:"environmentVariables"`
				} `yaml:"spec"`
			}

			renderComputeMode := func(platform install.PlatformConfig) string {
				out, err := install.RenderManifest("api-server-resource-cr.yaml", install.TemplateValues{
					Domain:   "stackdome.example.com",
					Platform: platform,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(yaml.Unmarshal(out, &manifest)).To(Succeed())

				for _, environmentVariable := range manifest.Spec.EnvironmentVariables {
					if environmentVariable.Name == "COMPUTE_MODE" {
						return environmentVariable.Value
					}
				}
				return ""
			}

			Expect(renderComputeMode(install.PlatformConfig{BaseDomain: "apps.example.com"})).To(Equal("shared"))
			Expect(renderComputeMode(install.PlatformConfig{})).To(Equal("bring_your_own"))
		})

		It("exposes metrics only through the internal service", func() {
			type renderedPort struct {
				Name           string `yaml:"name"`
				Number         int    `yaml:"number"`
				Protocol       string `yaml:"protocol"`
				ExposeToPublic bool   `yaml:"exposeToPublic"`
			}
			var manifest struct {
				Spec struct {
					Ports []renderedPort `yaml:"ports"`
				} `yaml:"spec"`
			}

			out, err := install.RenderManifest("api-server-resource-cr.yaml", install.TemplateValues{
				Domain: "stackdome.example.com",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(yaml.Unmarshal(out, &manifest)).To(Succeed())
			Expect(manifest.Spec.Ports).To(ConsistOf(
				renderedPort{Name: "http", Number: 8000, Protocol: "http", ExposeToPublic: true},
				renderedPort{Name: "metrics", Number: 9090, Protocol: "http", ExposeToPublic: false},
			))
		})

		It("renders the image, public port and platform configuration", func() {
			platformConfig := install.PlatformConfig{
				BaseDomain:         "apps.example.com",
				TLSEnabled:         true,
				CloudflareAPIToken: "cloudflare-token",
				ACMEEnvironment:    "staging",
			}
			sharedComputeConfig := install.SharedComputeConfig{
				ClusterAPIURL: "https://10.0.0.1:443",
				ClusterCAData: "ca-data",
				ClusterToken:  "cluster-token",
			}
			out, err := install.RenderManifest("api-server-resource-cr.yaml", install.TemplateValues{
				APIServerImage: "quay.io/stackdome/stackdome:main-abc1234",
				Domain:         "stackdome.example.com",
				StackUID:       "5a3a3a1e-0000-4000-8000-000000000001",
				TLSEnabled:     true,
				Platform:       platformConfig,
				SharedCompute:  sharedComputeConfig,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(yaml.Unmarshal(out, &map[string]any{})).To(Succeed())
			Expect(string(out)).To(ContainSubstring("uid: \"5a3a3a1e-0000-4000-8000-000000000001\""))
			Expect(string(out)).To(ContainSubstring("image: \"quay.io/stackdome/stackdome:main-abc1234\""))
			Expect(string(out)).To(ContainSubstring("fqdn: \"stackdome.example.com\""))
			Expect(string(out)).To(ContainSubstring("tls: true"))
			Expect(string(out)).To(ContainSubstring("name: PLATFORM_BASE_DOMAIN"))
			Expect(string(out)).To(ContainSubstring("key: platform-base-domain"))
			Expect(string(out)).To(ContainSubstring("name: PLATFORM_TLS_ENABLED\n      value: \"true\""))
			Expect(string(out)).To(ContainSubstring("name: SHARED_COMPUTE_CLUSTER_API_URL"))
			Expect(string(out)).To(ContainSubstring("key: shared-compute-cluster-api-url"))
			Expect(string(out)).To(ContainSubstring("name: SHARED_COMPUTE_CLUSTER_CA_DATA"))
			Expect(string(out)).To(ContainSubstring("key: shared-compute-cluster-ca-data"))
			Expect(string(out)).To(ContainSubstring("name: SHARED_COMPUTE_CLUSTER_TOKEN"))
			Expect(string(out)).To(ContainSubstring("key: shared-compute-cluster-token"))
			Expect(string(out)).To(ContainSubstring("name: PLATFORM_DNS_CLOUDFLARE_API_TOKEN"))
			Expect(string(out)).To(ContainSubstring("key: platform-cloudflare-api-token"))
			Expect(string(out)).To(ContainSubstring("name: PLATFORM_EMAIL"))
			Expect(string(out)).To(ContainSubstring("name: PLATFORM_ACME_ENVIRONMENT"))
			Expect(string(out)).NotTo(ContainSubstring("cloudflare-token"))

			secret, err := install.RenderManifest("bootstrap-secret.yaml", install.TemplateValues{
				AdminEmail:    "admin@example.com",
				Platform:      platformConfig,
				SharedCompute: sharedComputeConfig,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(yaml.Unmarshal(secret, &map[string]any{})).To(Succeed())
			Expect(string(secret)).To(ContainSubstring("platform-base-domain: \"apps.example.com\""))
			Expect(string(secret)).To(ContainSubstring("platform-tls-enabled: \"true\""))
			Expect(string(secret)).To(ContainSubstring("platform-cloudflare-api-token: \"cloudflare-token\""))
			Expect(string(secret)).To(ContainSubstring("shared-compute-cluster-api-url: \"https://10.0.0.1:443\""))
			Expect(string(secret)).To(ContainSubstring("shared-compute-cluster-ca-data: \"ca-data\""))
			Expect(string(secret)).To(ContainSubstring("shared-compute-cluster-token: \"cluster-token\""))
		})

		It("derives SERVER_EXTERNAL_URL from the domain and TLS mode", func() {
			out, err := install.RenderManifest("api-server-resource-cr.yaml", install.TemplateValues{
				Domain:     "stackdome.example.com",
				TLSEnabled: true,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).To(ContainSubstring("value: \"https://stackdome.example.com\""))

			out, err = install.RenderManifest("api-server-resource-cr.yaml", install.TemplateValues{
				Domain:     "stackdome.10.0.0.1.nip.io",
				TLSEnabled: false,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).To(ContainSubstring("value: \"http://stackdome.10.0.0.1.nip.io\""))
		})

		It("renders the GitHub OAuth credentials only when set", func() {
			out, err := install.RenderManifest("api-server-resource-cr.yaml", install.TemplateValues{
				Domain:             "stackdome.example.com",
				GitHubClientID:     "Iv1.abc",
				GitHubClientSecret: "shhh",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).To(ContainSubstring("name: GITHUB_CLIENT_ID"))
			Expect(string(out)).To(ContainSubstring("value: \"Iv1.abc\""))
			Expect(string(out)).To(ContainSubstring("value: \"shhh\""))

			out, err = install.RenderManifest("api-server-resource-cr.yaml", install.TemplateValues{
				Domain: "stackdome.example.com",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).NotTo(ContainSubstring("GITHUB_CLIENT_ID"))
		})

		It("keeps a multi-line private key intact as a YAML scalar", func() {
			pem := "-----BEGIN RSA PRIVATE KEY-----\nabc\ndef\n-----END RSA PRIVATE KEY-----\n"
			out, err := install.RenderManifest("api-server-resource-cr.yaml", install.TemplateValues{
				Domain:              "stackdome.example.com",
				GitHubAppID:         "4242",
				GitHubAppSlug:       "stackdome-cloud",
				GitHubAppPrivateKey: pem,
			})
			Expect(err).NotTo(HaveOccurred())

			var parsed struct {
				Spec struct {
					EnvironmentVariables []struct {
						Name  string `yaml:"name"`
						Value string `yaml:"value"`
					} `yaml:"environmentVariables"`
				} `yaml:"spec"`
			}
			Expect(yaml.Unmarshal(out, &parsed)).To(Succeed())

			env := map[string]string{}
			for _, e := range parsed.Spec.EnvironmentVariables {
				env[e.Name] = e.Value
			}
			Expect(env["GITHUB_APP_PRIVATE_KEY"]).To(Equal(pem))
			Expect(env["GITHUB_APP_ID"]).To(Equal("4242"))
		})

		It("renders shared compute without TLS-only platform configuration when TLS is disabled", func() {
			out, err := install.RenderManifest("api-server-resource-cr.yaml", install.TemplateValues{
				Domain: "stackdome.10.0.0.1.nip.io",
				Platform: install.PlatformConfig{
					BaseDomain: "apps.example.com",
					TLSEnabled: false,
				},
				SharedCompute: install.SharedComputeConfig{
					ClusterAPIURL: "https://10.0.0.1:443",
					ClusterCAData: "ca-data",
					ClusterToken:  "cluster-token",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).NotTo(ContainSubstring("cluster-issuer"))
			Expect(string(out)).To(ContainSubstring("name: PLATFORM_BASE_DOMAIN"))
			Expect(string(out)).To(ContainSubstring("name: PLATFORM_TLS_ENABLED\n      value: \"false\""))
			Expect(string(out)).To(ContainSubstring("name: SHARED_COMPUTE_CLUSTER_API_URL"))
			Expect(string(out)).NotTo(ContainSubstring("name: PLATFORM_EMAIL"))
			Expect(string(out)).NotTo(ContainSubstring("name: PLATFORM_DNS_CLOUDFLARE_API_TOKEN"))
			Expect(string(out)).NotTo(ContainSubstring("name: PLATFORM_ACME_ENVIRONMENT"))
		})
	})
})
