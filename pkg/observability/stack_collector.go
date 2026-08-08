package observability

import (
	"context"
	"fmt"
	"time"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	StacksMetricName               = "stackdome_stacks"
	StackResourcesMetricName       = "stackdome_stack_resources"
	StackFailuresMetricName        = "stackdome_stack_failures"
	StackSnapshotSuccessMetricName = "stackdome_stack_snapshot_collection_success"

	LabelState = "state"
	LabelType  = "type"

	UnknownState = "Unknown"

	stackSnapshotTimeout = 3 * time.Second
)

type StackSnapshotSource interface {
	Snapshot(context.Context) ([]*models.Stack, []*models.StackResource, error)
}

type StackSnapshotSourceFunc func(context.Context) ([]*models.Stack, []*models.StackResource, error)

func (f StackSnapshotSourceFunc) Snapshot(ctx context.Context) ([]*models.Stack, []*models.StackResource, error) {
	return f(ctx)
}

type databaseStackSnapshotSource struct {
	session db.SessionFactory
}

func NewDatabaseStackSnapshotSource(session db.SessionFactory) StackSnapshotSource {
	return &databaseStackSnapshotSource{session: session}
}

func (s *databaseStackSnapshotSource) Snapshot(ctx context.Context) ([]*models.Stack, []*models.StackResource, error) {
	var stacks []*models.Stack
	if err := s.session.New(ctx).Model(&models.Stack{}).
		Select("status").
		Where("deletion_timestamp IS NULL").
		Find(&stacks).Error; err != nil {
		return nil, nil, fmt.Errorf("load stack status snapshot: %w", err)
	}

	var resources []*models.StackResource
	if err := s.session.New(ctx).Model(&models.StackResource{}).
		Select("status").
		Find(&resources).Error; err != nil {
		return nil, nil, fmt.Errorf("load stack resource status snapshot: %w", err)
	}
	return stacks, resources, nil
}

type stackCollector struct {
	source    StackSnapshotSource
	stacks    *prometheus.Desc
	resources *prometheus.Desc
	failures  *prometheus.Desc
	success   *prometheus.Desc
}

func newStackCollector(source StackSnapshotSource) prometheus.Collector {
	return &stackCollector{
		source:    source,
		stacks:    prometheus.NewDesc(StacksMetricName, "Current stacks grouped by state.", []string{LabelState}, nil),
		resources: prometheus.NewDesc(StackResourcesMetricName, "Current stack resources grouped by state.", []string{LabelState}, nil),
		failures:  prometheus.NewDesc(StackFailuresMetricName, "Current non-ready stack resources grouped by failure type.", []string{LabelType}, nil),
		success:   prometheus.NewDesc(StackSnapshotSuccessMetricName, "Whether the latest stack status snapshot succeeded.", nil, nil),
	}
}

func (c *stackCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.stacks
	ch <- c.resources
	ch <- c.failures
	ch <- c.success
}

func (c *stackCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), stackSnapshotTimeout)
	defer cancel()
	stacks, resources, err := c.source.Snapshot(ctx)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.success, prometheus.GaugeValue, 0)
		return
	}

	for state, count := range countStackStates(stacks) {
		ch <- prometheus.MustNewConstMetric(c.stacks, prometheus.GaugeValue, float64(count), state)
	}
	resourceStates, failureTypes := countResourceStates(resources)
	for state, count := range resourceStates {
		ch <- prometheus.MustNewConstMetric(c.resources, prometheus.GaugeValue, float64(count), state)
	}
	for failureType, count := range failureTypes {
		ch <- prometheus.MustNewConstMetric(c.failures, prometheus.GaugeValue, float64(count), failureType)
	}
	ch <- prometheus.MustNewConstMetric(c.success, prometheus.GaugeValue, 1)
}

func countStackStates(stacks []*models.Stack) map[string]int {
	counts := map[string]int{}
	for _, stack := range stacks {
		state := UnknownState
		if stack.Status != nil && stack.Status.State != "" {
			state = boundedStackState(stack.Status.State)
		}
		counts[state]++
	}
	return counts
}

func countResourceStates(resources []*models.StackResource) (map[string]int, map[string]int) {
	states := map[string]int{}
	failures := map[string]int{}
	for _, resource := range resources {
		state := UnknownState
		if resource.Status != nil && resource.Status.State != "" {
			state = boundedResourceState(resource.Status.State)
		}
		states[state]++
		if resource.Status != nil && resource.Status.State != models.StackResourcePhaseReady && resource.Status.LastFailure != nil {
			failures[boundedFailureType(resource.Status.LastFailure.Type)]++
		}
	}
	return states, failures
}

func boundedStackState(state models.StackState) string {
	switch state {
	case models.StackReady, models.StackDeleting, models.StackPending, models.StackFailed,
		models.StackError, models.StackDegraded, models.StackProgressing:
		return string(state)
	default:
		return UnknownState
	}
}

func boundedResourceState(state models.StackResourceState) string {
	switch state {
	case models.StackResourcePhasePending, models.StackResourcePhaseReady,
		models.StackResourcePhaseFailed, models.StackResourcePhaseUnknown:
		return string(state)
	default:
		return UnknownState
	}
}

func boundedFailureType(failureType models.StackResourceFailureType) string {
	switch failureType {
	case models.FailureTypeRuntimeCrash, models.FailureTypeBuildFailure, models.FailureTypeReadinessFailure:
		return string(failureType)
	default:
		return UnknownState
	}
}
