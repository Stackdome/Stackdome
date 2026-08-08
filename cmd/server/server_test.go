package server

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/observability"
	"github.com/gorilla/mux"
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

var _ = Describe("request metrics", func() {
	It("uses the matched route template instead of the raw path", func() {
		metrics := observability.NewMetrics()
		router := mux.NewRouter()
		router.HandleFunc("/api/v1/stacks/{stack_id}", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}).Methods(http.MethodGet)
		handler := injectLoggerMiddleware(router, router, logger.NewLogger(), metrics)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/stacks/stack-secret-id", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)

		families, err := metrics.Gatherer().Gather()
		Expect(err).NotTo(HaveOccurred())
		var routes []string
		for _, family := range families {
			if family.GetName() != observability.HTTPRequestsMetricName {
				continue
			}
			for _, metric := range family.Metric {
				for _, label := range metric.Label {
					if label.GetName() == observability.LabelRoute {
						routes = append(routes, label.GetValue())
					}
				}
			}
		}
		Expect(routes).To(ConsistOf("/api/v1/stacks/{stack_id}"))
		Expect(routes).NotTo(ContainElement(ContainSubstring("stack-secret-id")))
	})

	It("serves Prometheus metrics from a dedicated handler", func() {
		metrics := observability.NewMetrics()
		metrics.ObserveHTTPRequest(http.MethodGet, "/health", http.StatusOK, 0)
		server := newMetricsServer("127.0.0.1:19090", metrics.Handler())
		recorder := httptest.NewRecorder()

		server.Handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))

		Expect(server.Addr).To(Equal("127.0.0.1:19090"))
		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(recorder.Header().Get("Content-Type")).To(ContainSubstring("text/plain"))
		Expect(recorder.Body.String()).To(ContainSubstring(observability.HTTPRequestsMetricName))
	})
})
