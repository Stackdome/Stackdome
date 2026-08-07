package stackfile

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/Stackdome/stackdome/pkg/api/openapi"
	"k8s.io/utils/ptr"
)

func loadFixtureBytes(name string) []byte {
	content, err := os.ReadFile(filepath.Join(testdataPath(""), name))
	Expect(err).NotTo(HaveOccurred())
	return content
}

var _ = Describe("FromStack", func() {
	It("round-trips every fixture through ToStack and back", func() {
		entries, err := os.ReadDir(testdataPath(""))
		Expect(err).NotTo(HaveOccurred())
		for _, e := range entries {
			sf, err := Load(loadFixtureBytes(e.Name()))
			Expect(err).NotTo(HaveOccurred(), "fixture %s", e.Name())

			stack, err := sf.ToStack()
			Expect(err).NotTo(HaveOccurred(), "fixture %s", e.Name())

			exported, err := FromStack(&stack)
			Expect(err).NotTo(HaveOccurred(), "fixture %s", e.Name())

			stack2, err := exported.ToStack()
			Expect(err).NotTo(HaveOccurred(), "fixture %s", e.Name())
			Expect(stack2).To(Equal(stack), "fixture %s must survive a round trip", e.Name())
		}
	})

	It("rejects constructs a stackfile cannot express", func() {
		stack := openapi.Stack{
			Name: "unsupported",
			Spec: openapi.StackSpec{
				StackResources: []openapi.StackResource{{
					Name:     "web",
					Source:   &openapi.SourceSpec{Image: openapi.NewImageSource("nginx:latest")},
					InitSpec: &openapi.InitSpec{Command: []string{"sh"}},
				}},
			},
		}
		_, err := FromStack(&stack)
		Expect(err).To(MatchError(ContainSubstring("init_spec")))
	})

	It("rejects id-only node refs", func() {
		stack := openapi.Stack{
			Name: "id-refs",
			Spec: openapi.StackSpec{
				StackResources: []openapi.StackResource{{
					Name:   "web",
					Source: &openapi.SourceSpec{Image: openapi.NewImageSource("nginx:latest")},
				}},
				Connections: []openapi.StackConnection{{
					Kind: connectionKindEnv,
					From: openapi.TopologyNodeRef{Type: nodeTypeSecret, Id: ptr.To("some-uuid")},
					To:   openapi.TopologyNodeRef{Type: nodeTypeStackResource, Name: ptr.To("web")},
				}},
			},
		}
		_, err := FromStack(&stack)
		Expect(err).To(MatchError(ContainSubstring("id-only")))
	})
})

var _ = Describe("ToStack determinism", func() {
	It("emits identical documents across repeated conversions", func() {
		sf, err := Load(loadFixtureBytes("kitchen_sink.yaml"))
		Expect(err).NotTo(HaveOccurred())

		first, err := sf.ToStack()
		Expect(err).NotTo(HaveOccurred())
		for range 5 {
			again, err := sf.ToStack()
			Expect(err).NotTo(HaveOccurred())
			Expect(again).To(Equal(first))
		}
	})
})
