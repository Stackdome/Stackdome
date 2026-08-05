package stackfile

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LoadWithWarnings", func() {
	loadYAML := func(content string) []string {
		sf, warnings, err := LoadWithWarnings([]byte(content))
		Expect(err).NotTo(HaveOccurred())
		Expect(sf).NotTo(BeNil())
		return warnings
	}

	const cleanFile = `
name: demo
resources:
  web:
    image: nginx:latest
    ports:
      - name: http
        port: 80
`

	It("returns no warnings for a clean file", func() {
		Expect(loadYAML(cleanFile)).To(BeEmpty())
	})

	It("warns about an unknown top-level key", func() {
		warnings := loadYAML(cleanFile + "bogus_key_xyz: true\n")
		Expect(warnings).To(ConsistOf(`unknown key "bogus_key_xyz" in top level`))
	})

	It("warns about a legacy stateful flag on a resource", func() {
		warnings := loadYAML(cleanFile + "    stateful: true\n")
		Expect(warnings).To(ConsistOf(`unknown key "stateful" in resources.web`))
	})

	It("warns about a misspelled resource key", func() {
		warnings := loadYAML(`
name: demo
resources:
  web:
    image: nginx:latest
    imagee: nginx:latest
`)
		Expect(warnings).To(ConsistOf(`unknown key "imagee" in resources.web`))
	})

	It("warns about an unknown key in a build block", func() {
		warnings := loadYAML(`
name: demo
resources:
  api:
    build:
      repo: https://github.com/acme/app.git
      branchh: main
`)
		Expect(warnings).To(ConsistOf(`unknown key "branchh" in resources.api.build`))
	})

	It("warns about an unknown key in a port entry", func() {
		warnings := loadYAML(`
name: demo
resources:
  web:
    image: nginx:latest
    ports:
      - name: http
        port: 80
        publicc: true
`)
		Expect(warnings).To(ConsistOf(`unknown key "publicc" in resources.web.ports[0]`))
	})

	It("warns about an unknown key in a volume definition", func() {
		warnings := loadYAML(`
name: demo
volumes:
  data:
    size: 1Gi
    sizee: 2Gi
resources:
  web:
    image: nginx:latest
`)
		Expect(warnings).To(ConsistOf(`unknown key "sizee" in volumes.data`))
	})

	It("warns about an unknown key in a postgres addon block", func() {
		warnings := loadYAML(`
name: demo
resources:
  web:
    image: nginx:latest
    addons:
      pg:
        type: postgres
        database: appdb
        superuser: true
        databse: typo
        env:
          DB_HOST: "{{ host }}"
`)
		Expect(warnings).To(ConsistOf(`unknown key "databse" in resources.web.addons.pg`))
	})

	It("does not flag env or secret keys", func() {
		warnings := loadYAML(`
name: demo
resources:
  web:
    image: nginx:latest
    env:
      IMAGEE: anything
      stateful: "true"
    secrets:
      api-keys:
        STRIPE_KEY: stripe
`)
		Expect(warnings).To(BeEmpty())
	})

	DescribeTable("legacy testdata fixtures load with only the stateful warning",
		func(fixture string) {
			content, err := os.ReadFile(testdataPath(fixture))
			Expect(err).NotTo(HaveOccurred())

			sf, warnings, err := LoadWithWarnings(content)
			Expect(err).NotTo(HaveOccurred())
			Expect(sf).NotTo(BeNil())
			Expect(warnings).To(HaveLen(1))
			Expect(warnings[0]).To(ContainSubstring(`unknown key "stateful" in resources.`))

			viaLoad, err := Load(content)
			Expect(err).NotTo(HaveOccurred())
			Expect(viaLoad).To(Equal(sf))
		},
		Entry("infisical", "infisical.yaml"),
		Entry("kitchen sink", "kitchen_sink.yaml"),
	)
})
