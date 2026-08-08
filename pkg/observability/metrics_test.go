package observability

import (
	"reflect"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dto "github.com/prometheus/client_model/go"
)

const (
	testMethod      = "GET"
	testRoute       = "/api/v1/stacks/{stack_id}"
	testStatus      = 503
	testStatusClass = "5xx"
)

var _ = Describe("HTTP metrics", func() {
	It("records request count, status, and duration with bounded labels", func() {
		metrics := NewMetrics()

		metrics.ObserveHTTPRequest(testMethod, testRoute, testStatus, 250*time.Millisecond)

		families, err := metrics.Gatherer().Gather()
		Expect(err).NotTo(HaveOccurred())
		requests := familyNamed(families, HTTPRequestsMetricName)
		Expect(metricValue(requests, map[string]string{
			LabelMethod: testMethod,
			LabelRoute:  testRoute,
			LabelStatus: "503",
		})).To(Equal(float64(1)))
		duration := familyNamed(families, HTTPDurationMetricName)
		Expect(histogramCount(duration, map[string]string{
			LabelMethod:      testMethod,
			LabelRoute:       testRoute,
			LabelStatusClass: testStatusClass,
		})).To(Equal(uint64(1)))
	})

	It("collapses non-standard HTTP methods into one bounded label", func() {
		metrics := NewMetrics()

		metrics.ObserveHTTPRequest("BREW", testRoute, 418, time.Millisecond)

		families, err := metrics.Gatherer().Gather()
		Expect(err).NotTo(HaveOccurred())
		requests := familyNamed(families, HTTPRequestsMetricName)
		Expect(metricValue(requests, map[string]string{
			LabelMethod: UnknownMethod,
			LabelRoute:  testRoute,
			LabelStatus: "418",
		})).To(Equal(float64(1)))
	})

	It("tracks in-flight requests without a path label", func() {
		metrics := NewMetrics()

		metrics.IncHTTPInFlight(testMethod)
		defer metrics.DecHTTPInFlight(testMethod)

		families, err := metrics.Gatherer().Gather()
		Expect(err).NotTo(HaveOccurred())
		inFlight := familyNamed(families, HTTPInFlightMetricName)
		Expect(metricValue(inFlight, map[string]string{LabelMethod: testMethod})).To(Equal(float64(1)))
	})

	It("records worker execution outcomes", func() {
		metrics := NewMetrics()

		metrics.ObserveWorkerExecution("release-worker", WorkerResultError)

		families, err := metrics.Gatherer().Gather()
		Expect(err).NotTo(HaveOccurred())
		outcomes := familyNamed(families, "stackdome_worker_executions_total")
		Expect(metricValue(outcomes, map[string]string{
			LabelWorker: "release-worker",
			LabelResult: WorkerResultError,
		})).To(Equal(float64(1)))
	})
})

func familyNamed(families []*dto.MetricFamily, name string) *dto.MetricFamily {
	for _, family := range families {
		if family.GetName() == name {
			return family
		}
	}
	return nil
}

func metricValue(family *dto.MetricFamily, labels map[string]string) float64 {
	Expect(family).NotTo(BeNil())
	for _, metric := range family.Metric {
		if hasLabels(metric, labels) {
			if metric.Counter != nil {
				return metric.Counter.GetValue()
			}
			return metric.Gauge.GetValue()
		}
	}
	Fail("metric with expected labels not found")
	return 0
}

func histogramCount(family *dto.MetricFamily, labels map[string]string) uint64 {
	Expect(family).NotTo(BeNil())
	for _, metric := range family.Metric {
		if hasLabels(metric, labels) {
			return metric.Histogram.GetSampleCount()
		}
	}
	Fail("histogram with expected labels not found")
	return 0
}

func hasLabels(metric *dto.Metric, want map[string]string) bool {
	got := map[string]string{}
	for _, pair := range metric.Label {
		got[pair.GetName()] = pair.GetValue()
	}
	if len(want) == 0 {
		return len(got) == 0
	}
	return reflect.DeepEqual(got, want)
}
