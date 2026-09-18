package stackdeploy

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStackDeploy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "StackDeploy Suite")
}
