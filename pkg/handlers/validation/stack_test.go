package validation

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"k8s.io/utils/ptr"
)

func TestValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Handlers Validation Suite")
}

// Regression: ValidateStack/ValidateStackShell reflected on Stack.Status after
// the release-centric contract removed that field from the generated
// openapi.Stack. reflect.FieldByName on a missing field yields an invalid
// Value whose String() is "<invalid Value>", so validateEmpty rejected EVERY
// stack create/apply with "status must be empty".
var _ = Describe("ValidateStack", func() {
	validStack := func() *openapi.Stack {
		return &openapi.Stack{
			Name: "web-app",
			Spec: openapi.StackSpec{
				StackResources: []openapi.StackResource{{
					Name: "web",
					Source: &openapi.SourceSpec{
						Image: &openapi.ImageSource{Ref: "nginx:latest"},
					},
				}},
			},
		}
	}

	It("accepts a valid create payload", func() {
		Expect(ValidateStack(validStack())()).To(BeNil())
	})

	It("accepts a valid shell payload", func() {
		Expect(ValidateStackShell(validStack())()).To(BeNil())
	})

	It("still rejects read-only fields the client must not set", func() {
		s := validStack()
		s.Id = ptr.To("someid")
		err := ValidateStack(s)()
		Expect(err).NotTo(BeNil())
		Expect(err.Reason).To(ContainSubstring("id must be empty"))
	})

	It("still requires a name", func() {
		s := validStack()
		s.Name = ""
		Expect(ValidateStack(s)()).NotTo(BeNil())
	})
})
