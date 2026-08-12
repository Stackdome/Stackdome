package turnstile

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Turnstile verifier", func() {
	It("submits the canonical siteverify fields", func() {
		server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			defer GinkgoRecover()

			Expect(request.Method).To(Equal(http.MethodPost))
			Expect(request.Header.Get("Content-Type")).To(Equal("application/x-www-form-urlencoded"))
			Expect(request.ParseForm()).To(Succeed())
			Expect(request.Form.Get("secret")).To(Equal("test-secret"))
			Expect(request.Form.Get("response")).To(Equal("test-token"))
			Expect(request.Form.Get("remoteip")).To(Equal("203.0.113.10"))

			Expect(json.NewEncoder(response).Encode(map[string]any{
				"success":  true,
				"hostname": "stackdome.com",
				"action":   "turnstile-spin-v2",
			})).To(Succeed())
		}))
		DeferCleanup(server.Close)

		verifier := newClient(validClientSpec(), server.URL)
		Expect(verifier.Verify(
			context.Background(),
			"test-token",
			netip.MustParseAddr("203.0.113.10"),
		)).To(Succeed())
	})

	DescribeTable("rejecting untrusted siteverify results",
		func(responseBody map[string]any) {
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
				defer GinkgoRecover()

				Expect(json.NewEncoder(response).Encode(responseBody)).To(Succeed())
			}))
			DeferCleanup(server.Close)

			verifier := newClient(validClientSpec(), server.URL)
			err := verifier.Verify(context.Background(), "test-token", netip.MustParseAddr("203.0.113.10"))
			Expect(err).To(HaveOccurred())
		},
		Entry(
			"when the challenge is rejected",
			map[string]any{"success": false},
		),
		Entry(
			"when the hostname differs",
			map[string]any{
				"success": true, "hostname": "attacker.example", "action": "turnstile-spin-v2",
			},
		),
		Entry(
			"when the action differs",
			map[string]any{
				"success": true, "hostname": "stackdome.com", "action": "different-action",
			},
		),
	)

	It("fails closed when siteverify is unavailable", func() {
		server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
			response.WriteHeader(http.StatusServiceUnavailable)
		}))
		serverURL := server.URL
		server.Close()

		verifier := newClient(validClientSpec(), serverURL)
		err := verifier.Verify(context.Background(), "test-token", netip.MustParseAddr("203.0.113.10"))
		Expect(err).To(HaveOccurred())
	})

	It("rejects incomplete configuration", func() {
		_, err := NewClient(ClientSpec{Timeout: time.Second})
		Expect(err).To(MatchError("turnstile secret is required"))
	})
})

func validClientSpec() ClientSpec {
	return ClientSpec{
		Secret:           "test-secret",
		ExpectedHostname: "stackdome.com",
		ExpectedAction:   "turnstile-spin-v2",
		Timeout:          time.Second,
	}
}
