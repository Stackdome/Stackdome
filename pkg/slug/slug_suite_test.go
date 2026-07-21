package slug_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSlug(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Slug Suite")
}
