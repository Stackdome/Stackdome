package signupprotection

import (
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client IP resolution", func() {
	Describe("Cloudflare", func() {
		It("accepts one valid CF-Connecting-IP value", func() {
			request := httptest.NewRequest("POST", "/api/v1/user-signup", nil)
			request.Header.Set("CF-Connecting-IP", "::ffff:203.0.113.10")

			clientIP, err := NewCloudflareClientIPResolver().Resolve(request)

			Expect(err).NotTo(HaveOccurred())
			Expect(clientIP.String()).To(Equal("203.0.113.10"))
		})

		DescribeTable("rejecting ambiguous or invalid values",
			func(values ...string) {
				request := httptest.NewRequest("POST", "/api/v1/user-signup", nil)
				for _, value := range values {
					request.Header.Add("CF-Connecting-IP", value)
				}

				_, err := NewCloudflareClientIPResolver().Resolve(request)
				Expect(err).To(MatchError(ErrClientIPUnavailable))
			},
			Entry("when the header is absent"),
			Entry("when the value is invalid", "not-an-ip"),
			Entry("when one value contains a list", "203.0.113.10, 203.0.113.11"),
			Entry("when multiple values are present", "203.0.113.10", "203.0.113.11"),
		)
	})

	Describe("direct connections", func() {
		It("uses RemoteAddr and ignores forwarding headers", func() {
			request := httptest.NewRequest("POST", "/api/v1/user-signup", nil)
			request.RemoteAddr = "[2001:db8::10]:4321"
			request.Header.Set("X-Forwarded-For", "203.0.113.99")

			clientIP, err := NewDirectClientIPResolver().Resolve(request)

			Expect(err).NotTo(HaveOccurred())
			Expect(clientIP.String()).To(Equal("2001:db8::10"))
		})
	})
})
