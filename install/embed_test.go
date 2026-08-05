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

var _ = Describe("RenderManifest", func() {
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
		It("renders the image and public port", func() {
			out, err := install.RenderManifest("api-server-resource-cr.yaml", install.TemplateValues{
				APIServerImage: "quay.io/stackdome/stackdome:main-abc1234",
				Domain:         "stackdome.example.com",
				StackUID:       "5a3a3a1e-0000-4000-8000-000000000001",
				TLSEnabled:     true,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).To(ContainSubstring("uid: \"5a3a3a1e-0000-4000-8000-000000000001\""))
			Expect(string(out)).To(ContainSubstring("image: \"quay.io/stackdome/stackdome:main-abc1234\""))
			Expect(string(out)).To(ContainSubstring("fqdn: \"stackdome.example.com\""))
			Expect(string(out)).To(ContainSubstring("tls: true"))
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

		It("omits the cluster-issuer annotation without TLS", func() {
			out, err := install.RenderManifest("api-server-resource-cr.yaml", install.TemplateValues{
				Domain:     "stackdome.10.0.0.1.nip.io",
				TLSEnabled: false,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).NotTo(ContainSubstring("cluster-issuer"))
		})
	})
})
