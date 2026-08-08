package observability

import (
	"context"
	stderrors "errors"

	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
)

var _ = Describe("stack snapshot metrics", func() {
	It("groups current stack, resource, and active failure states", func() {
		source := StackSnapshotSourceFunc(func(context.Context) ([]*models.Stack, []*models.StackResource, error) {
			return []*models.Stack{
					{Status: &models.StackStatus{State: models.StackReady}},
					{Status: &models.StackStatus{State: models.StackFailed}},
					{Status: &models.StackStatus{State: models.StackFailed}},
					{},
					{Status: &models.StackStatus{State: models.StackState("unexpected-stack-state")}},
				}, []*models.StackResource{
					{Status: &models.StackResourceStatus{State: models.StackResourcePhaseReady, LastFailure: failure(models.FailureTypeBuildFailure)}},
					{Status: &models.StackResourceStatus{State: models.StackResourcePhaseFailed, LastFailure: failure(models.FailureTypeRuntimeCrash)}},
					{Status: &models.StackResourceStatus{State: models.StackResourcePhaseUnknown, LastFailure: failure(models.FailureTypeReadinessFailure)}},
					{},
					{Status: &models.StackResourceStatus{State: models.StackResourceState("unexpected-resource-state"), LastFailure: failure(models.StackResourceFailureType("unexpected-failure"))}},
				}, nil
		})
		registry := prometheus.NewRegistry()
		registry.MustRegister(newStackCollector(source))

		families, err := registry.Gather()
		Expect(err).NotTo(HaveOccurred())
		Expect(metricValue(familyNamed(families, StacksMetricName), map[string]string{LabelState: string(models.StackReady)})).To(Equal(float64(1)))
		Expect(metricValue(familyNamed(families, StacksMetricName), map[string]string{LabelState: string(models.StackFailed)})).To(Equal(float64(2)))
		Expect(metricValue(familyNamed(families, StacksMetricName), map[string]string{LabelState: UnknownState})).To(Equal(float64(2)))
		Expect(metricValue(familyNamed(families, StackResourcesMetricName), map[string]string{LabelState: string(models.StackResourcePhaseReady)})).To(Equal(float64(1)))
		Expect(metricValue(familyNamed(families, StackFailuresMetricName), map[string]string{LabelType: string(models.FailureTypeRuntimeCrash)})).To(Equal(float64(1)))
		Expect(metricValue(familyNamed(families, StackFailuresMetricName), map[string]string{LabelType: string(models.FailureTypeReadinessFailure)})).To(Equal(float64(1)))
		Expect(metricValue(familyNamed(families, StackFailuresMetricName), map[string]string{LabelType: UnknownState})).To(Equal(float64(1)))
		Expect(familyNamed(families, StackFailuresMetricName).Metric).To(HaveLen(3), "ready resources must not expose historical failures")
		Expect(metricValue(familyNamed(families, StackSnapshotSuccessMetricName), nil)).To(Equal(float64(1)))
	})

	It("reports collection failure without publishing stale state", func() {
		source := StackSnapshotSourceFunc(func(context.Context) ([]*models.Stack, []*models.StackResource, error) {
			return nil, nil, stderrors.New("database unavailable")
		})
		registry := prometheus.NewRegistry()
		registry.MustRegister(newStackCollector(source))

		families, err := registry.Gather()
		Expect(err).NotTo(HaveOccurred())
		Expect(familyNamed(families, StacksMetricName)).To(BeNil())
		Expect(metricValue(familyNamed(families, StackSnapshotSuccessMetricName), nil)).To(Equal(float64(0)))
	})
})

func failure(failureType models.StackResourceFailureType) *models.StackResourceFailure {
	return &models.StackResourceFailure{Type: failureType}
}
