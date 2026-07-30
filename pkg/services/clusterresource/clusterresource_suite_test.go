package clusterresource

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestClusterResourceSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ClusterResource Suite")
}
