package stack

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStackWorker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stack Worker Suite")
}
