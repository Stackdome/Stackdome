package install_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
				TLSEnabled:     true,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).To(ContainSubstring("image: \"quay.io/stackdome/stackdome:main-abc1234\""))
			Expect(string(out)).To(ContainSubstring("fqdn: \"stackdome.example.com\""))
			Expect(string(out)).To(ContainSubstring("tls: true"))
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
