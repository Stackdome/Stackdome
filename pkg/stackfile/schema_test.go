package stackfile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

func compileSchema() *jsonschema.Schema {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(SchemaJSON))
	Expect(err).NotTo(HaveOccurred())
	c := jsonschema.NewCompiler()
	Expect(c.AddResource("stackfile.schema.json", doc)).To(Succeed())
	schema, err := c.Compile("stackfile.schema.json")
	Expect(err).NotTo(HaveOccurred())
	return schema
}

// yamlToJSONValue re-encodes YAML through JSON so the validator sees
// json-typed values (float64 numbers, string-keyed maps).
func yamlToJSONValue(content []byte) any {
	var v any
	Expect(yaml.Unmarshal(content, &v)).To(Succeed())
	raw, err := json.Marshal(v)
	Expect(err).NotTo(HaveOccurred())
	parsed, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	Expect(err).NotTo(HaveOccurred())
	return parsed
}

var _ = Describe("Stackfile JSON Schema", func() {
	var schema *jsonschema.Schema

	BeforeEach(func() {
		schema = compileSchema()
	})

	It("accepts every testdata fixture", func() {
		entries, err := os.ReadDir(testdataPath(""))
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).NotTo(BeEmpty())
		for _, e := range entries {
			content, err := os.ReadFile(filepath.Join(testdataPath(""), e.Name()))
			Expect(err).NotTo(HaveOccurred())
			Expect(schema.Validate(yamlToJSONValue(content))).To(Succeed(), "fixture %s", e.Name())
		}
	})

	DescribeTable("rejects invalid documents",
		func(doc string) {
			Expect(schema.Validate(yamlToJSONValue([]byte(doc)))).NotTo(Succeed())
		},
		Entry("missing name", `
resources:
  web:
    image: nginx:latest
`),
		Entry("no resources", `
name: empty
resources: {}
`),
		Entry("both image and build", `
name: bad
resources:
  web:
    image: nginx:latest
    build:
      repo: https://example.com/repo.git
`),
		Entry("neither image nor build", `
name: bad
resources:
  web:
    ports:
      - name: http
        port: 80
`),
		Entry("unknown resource field", `
name: bad
resources:
  web:
    image: nginx:latest
    stateful: true
`),
		Entry("branch and tag together", `
name: bad
resources:
  web:
    build:
      repo: https://example.com/repo.git
      branch: main
      tag: v1.0.0
`),
		Entry("commit without branch or tag", `
name: bad
resources:
  web:
    build:
      repo: https://example.com/repo.git
      commit: abc123
`),
		Entry("port out of range", `
name: bad
resources:
  web:
    image: nginx:latest
    ports:
      - name: http
        port: 70000
`),
		Entry("invalid access mode", `
name: bad
resources:
  web:
    image: nginx:latest
volumes:
  data:
    size: 1Gi
    access_mode: ReadWriteSometimes
`),
		Entry("invalid volume size format", `
name: bad
resources:
  web:
    image: nginx:latest
volumes:
  data:
    size: five-gigs
`),
		Entry("invalid workload type", `
name: bad
resources:
  web:
    image: nginx:latest
    workload_type: Deployment
`),
	)

	It("accepts commit pinned to a branch", func() {
		doc := `
name: pinned
resources:
  web:
    build:
      repo: https://example.com/repo.git
      branch: main
      commit: abc123
`
		Expect(schema.Validate(yamlToJSONValue([]byte(doc)))).To(Succeed())
	})

	It("accepts a minimal valid document", func() {
		doc := `
name: minimal
resources:
  web:
    image: nginx:latest
`
		Expect(schema.Validate(yamlToJSONValue([]byte(doc)))).To(Succeed())
	})
})
