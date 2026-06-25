package preview

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPreviewWorker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Preview Worker Suite")
}
