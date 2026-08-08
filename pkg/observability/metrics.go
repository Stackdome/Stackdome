package observability

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/component-base/metrics/legacyregistry"
)

const (
	metricsNamespace    = "stackdome"
	httpServerSubsystem = "http_server"

	HTTPRequestsMetricName = "stackdome_http_server_requests_total"
	HTTPDurationMetricName = "stackdome_http_server_request_duration_seconds"
	HTTPInFlightMetricName = "stackdome_http_server_in_flight_requests"

	LabelMethod      = "method"
	LabelRoute       = "route"
	LabelStatus      = "status"
	LabelStatusClass = "status_class"
	LabelWorker      = "worker"
	LabelResult      = "result"

	UnmatchedRoute = "unmatched"
	UnknownMethod  = "UNKNOWN"

	WorkerResultSuccess = "success"
	WorkerResultError   = "error"
	WorkerResultRequeue = "requeue"
	WorkerResultPanic   = "panic"
)

var requestDurationBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 120}

type Metrics struct {
	registry       *prometheus.Registry
	httpRequests   *prometheus.CounterVec
	httpDuration   *prometheus.HistogramVec
	httpInFlight   *prometheus.GaugeVec
	workerOutcomes *prometheus.CounterVec
}

func NewMetrics() *Metrics {
	m := &Metrics{
		registry: prometheus.NewRegistry(),
		httpRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: httpServerSubsystem,
			Name:      "requests_total",
			Help:      "Total HTTP requests completed by the API server.",
		}, []string{LabelMethod, LabelRoute, LabelStatus}),
		httpDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: httpServerSubsystem,
			Name:      "request_duration_seconds",
			Help:      "API request duration in seconds.",
			Buckets:   requestDurationBuckets,
		}, []string{LabelMethod, LabelRoute, LabelStatusClass}),
		httpInFlight: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: httpServerSubsystem,
			Name:      "in_flight_requests",
			Help:      "Current API requests being processed.",
		}, []string{LabelMethod}),
		workerOutcomes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: "worker",
			Name:      "executions_total",
			Help:      "Total worker executions by outcome.",
		}, []string{LabelWorker, LabelResult}),
	}
	m.registry.MustRegister(m.httpRequests, m.httpDuration, m.httpInFlight, m.workerOutcomes)
	return m
}

func (m *Metrics) Gatherer() prometheus.Gatherer {
	return m.registry
}

func (m *Metrics) Handler() http.Handler {
	gatherers := prometheus.Gatherers{m.registry, legacyregistry.DefaultGatherer}
	return promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{})
}

func (m *Metrics) RegisterStackCollector(source StackSnapshotSource) {
	m.registry.MustRegister(newStackCollector(source))
}

func (m *Metrics) ObserveHTTPRequest(method, route string, status int, elapsed time.Duration) {
	method = boundedHTTPMethod(method)
	m.httpRequests.WithLabelValues(method, route, strconv.Itoa(status)).Inc()
	m.httpDuration.WithLabelValues(method, route, statusClass(status)).Observe(elapsed.Seconds())
}

func (m *Metrics) IncHTTPInFlight(method string) {
	m.httpInFlight.WithLabelValues(boundedHTTPMethod(method)).Inc()
}

func (m *Metrics) DecHTTPInFlight(method string) {
	m.httpInFlight.WithLabelValues(boundedHTTPMethod(method)).Dec()
}

func (m *Metrics) ObserveWorkerExecution(worker, result string) {
	m.workerOutcomes.WithLabelValues(worker, result).Inc()
}

func statusClass(status int) string {
	return fmt.Sprintf("%dxx", status/100)
}

func boundedHTTPMethod(method string) string {
	switch method {
	case http.MethodConnect, http.MethodDelete, http.MethodGet, http.MethodHead,
		http.MethodOptions, http.MethodPatch, http.MethodPost, http.MethodPut, http.MethodTrace:
		return method
	default:
		return UnknownMethod
	}
}
