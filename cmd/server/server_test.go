package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Server Suite")
}

var _ = Describe("newRequestTimeoutMiddleware", func() {
	const testTimeout = 50 * time.Millisecond

	var (
		router            *mux.Router
		normalDeadlineSet bool
		streamDeadlineSet bool
	)

	captureDeadline := func(target *bool) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			_, ok := r.Context().Deadline()
			*target = ok
			w.WriteHeader(http.StatusOK)
		}
	}

	BeforeEach(func() {
		normalDeadlineSet = false
		streamDeadlineSet = false

		// Mirror production wiring: streaming routes live on a nested subrouter
		// and carry a name; the timeout middleware is applied to the parent.
		router = mux.NewRouter()
		sub := router.PathPrefix("/api/v1").Subrouter()
		sub.HandleFunc("/stacks/{id}/logs", captureDeadline(&streamDeadlineSet)).
			Methods(http.MethodGet).Name(routeNameStreamStackLogs)
		sub.HandleFunc("/stacks/{id}", captureDeadline(&normalDeadlineSet)).
			Methods(http.MethodGet)

		router.Use(newRequestTimeoutMiddleware(testTimeout, streamingRouteNames()))
	})

	It("applies a request timeout to a normal route", func() {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/stacks/abc", nil)
		router.ServeHTTP(httptest.NewRecorder(), req)
		Expect(normalDeadlineSet).To(BeTrue())
	})

	It("does not apply a request timeout to a streaming (SSE) route", func() {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/stacks/abc/logs", nil)
		router.ServeHTTP(httptest.NewRecorder(), req)
		Expect(streamDeadlineSet).To(BeFalse())
	})
})
