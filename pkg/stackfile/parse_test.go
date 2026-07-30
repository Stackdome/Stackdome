package stackfile

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/models"
)

var _ = Describe("output validation (human-readable scheme)", func() {
	// singlePort = one private port named "3306"; multiPort adds a public "web".
	// Use the package's existing Stackfile-parse entry point to exercise these.

	DescribeTable("single-port resource accepts unsuffixed keys",
		func(output string, valid bool) {
			err := validateSelfOutput("app", "ENV", output, []PortDef{{Name: "3306", Public: false}})
			if valid {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
			}
		},
		Entry("host", models.OutputNameHost, true),
		Entry("port", models.OutputNamePort, true),
		Entry("url", models.OutputNameURL, true),
		Entry("old suffixed port rejected", "port.3306", false),
		Entry("public on a private port rejected", models.OutputNamePublicURL, false),
	)

	DescribeTable("multi-port resource requires the port suffix",
		func(output string, valid bool) {
			err := validateSelfOutput("app", "ENV", output, []PortDef{{Name: "3306", Public: false}, {Name: "web", Public: true}})
			if valid {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
			}
		},
		Entry("host stays unsuffixed", models.OutputNameHost, true),
		Entry("url.3306", "url."+"3306", true),
		Entry("public_url.web", models.OutputNamePublicURL+".web", true),
		Entry("public_url.3306 (private port) rejected", models.OutputNamePublicURL+".3306", false),
		Entry("bare url (ambiguous) rejected", models.OutputNameURL, false),
	)
})
