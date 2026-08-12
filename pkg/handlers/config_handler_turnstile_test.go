package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Public signup configuration", func() {
	It("exposes only the Turnstile browser settings", func() {
		handler := NewConfigHandler(ConfigHandlerSpec{
			SignupTurnstileEnabled: true,
			SignupTurnstileSiteKey: "public-site-key",
			SignupTurnstileAction:  "turnstile-spin-v2",
		})
		request := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)
		response := httptest.NewRecorder()

		handler.Get(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		var body openapi.AppConfigResponse
		Expect(json.Unmarshal(response.Body.Bytes(), &body)).To(Succeed())
		Expect(body.Signup).NotTo(BeNil())
		Expect(body.Signup.Turnstile).To(Equal(openapi.TurnstileConfigResponse{
			Enabled: true,
			SiteKey: "public-site-key",
			Action:  "turnstile-spin-v2",
		}))
	})
})
