package turnstile

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTurnstile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Turnstile Suite")
}
