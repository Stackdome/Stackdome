package pgstore_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPgstore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PgStore Suite")
}
