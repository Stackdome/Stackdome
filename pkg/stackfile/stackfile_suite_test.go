package stackfile

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStackfile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stackfile Suite")
}
