package signupprotection

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSignupProtection(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Signup Protection Suite")
}
