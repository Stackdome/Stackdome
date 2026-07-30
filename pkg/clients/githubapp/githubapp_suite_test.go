package githubapp

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGithubApp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GithubApp Suite")
}
