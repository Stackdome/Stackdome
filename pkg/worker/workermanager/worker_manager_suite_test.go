package workermanager

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWorkerManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Worker Manager Suite")
}
