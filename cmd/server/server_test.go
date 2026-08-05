package server

import (
	"regexp"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Server Suite")
}

var _ = Describe("publicAPIPaths", func() {
	matches := func(path string) bool {
		for _, expr := range publicAPIPaths {
			if regexp.MustCompile(expr).MatchString(path) {
				return true
			}
		}
		return false
	}

	It("lets the unauthenticated GitHub routes through", func() {
		// Browser redirects from GitHub and the webhook receiver carry no JWT;
		// each is protected by its own mechanism (state / HMAC).
		Expect(matches("/api/v1/git-integrations/github/manifest/callback")).To(BeTrue())
		Expect(matches("/api/v1/git-integrations/github/setup")).To(BeTrue())
		Expect(matches("/api/v1/webhooks/github")).To(BeTrue())
	})

	It("keeps the git-integrations API authenticated", func() {
		Expect(matches("/api/v1/organizations/org-1/git-integrations")).To(BeFalse())
		Expect(matches("/api/v1/git-integrations/github/manifest")).To(BeFalse())
		Expect(matches("/api/v1/stacks")).To(BeFalse())
	})
})
