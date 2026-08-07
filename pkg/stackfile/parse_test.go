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

var _ = Describe("Load strictness", func() {
	It("rejects unknown fields", func() {
		_, err := Load([]byte(`
name: bad
resources:
  web:
    image: nginx:latest
    stateful: true
`))
		Expect(err).To(MatchError(ContainSubstring("stateful")))
	})

	It("rejects unknown fields in addon blocks", func() {
		_, err := Load([]byte(`
name: bad
resources:
  web:
    image: nginx:latest
    addons:
      db:
        type: postgres
        size: 10Gi
        env:
          DB_HOST: "{{ host }}"
`))
		Expect(err).To(MatchError(ContainSubstring(`unknown field "size" in addon config`)))
	})

	It("rejects dotted refs in addon env", func() {
		_, err := Load([]byte(`
name: bad
resources:
  web:
    image: nginx:latest
    addons:
      db:
        type: postgres
        env:
          DB_HOST: "{{ postgres.host }}"
`))
		Expect(err).To(MatchError(ContainSubstring("bare names")))
	})
})

var _ = Describe("build ref validation", func() {
	load := func(build string) error {
		_, err := Load([]byte(`
name: refs
resources:
  app:
    build:
      repo: https://example.com/repo.git
` + build))
		return err
	}

	It("rejects branch with tag", func() {
		Expect(load("      branch: main\n      tag: v1.0.0\n")).To(MatchError(ContainSubstring("mutually exclusive")))
	})

	It("rejects a bare commit", func() {
		Expect(load("      commit: abc123\n")).To(MatchError(ContainSubstring("requires 'branch' or 'tag'")))
	})

	It("accepts commit pinned to a branch", func() {
		Expect(load("      branch: main\n      commit: abc123\n")).To(Succeed())
	})

	It("accepts commit pinned to a tag", func() {
		Expect(load("      tag: v1.0.0\n      commit: abc123\n")).To(Succeed())
	})
})
