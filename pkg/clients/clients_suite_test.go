package clients

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestClientsSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clients Suite")
}
