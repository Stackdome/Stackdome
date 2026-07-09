package postgresaddon

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPostgresAddonWorker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Postgres Addon Worker Suite")
}
