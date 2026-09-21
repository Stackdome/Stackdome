package stackdeploy

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStackdeploy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stackdeploy Suite")
}
