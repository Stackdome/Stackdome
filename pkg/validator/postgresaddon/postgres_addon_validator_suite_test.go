package postgresaddon

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPostgresAddonValidator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PostgresAddon Validator Suite")
}
