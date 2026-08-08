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
	Snapshot(context.Context) (StackMetricSnapshot, error)
}

type StackMetricSnapshot struct {
	StackStates    map[string]int64
	ResourceStates map[string]int64
	FailureTypes   map[string]int64
}

type StackSnapshotSourceFunc func(context.Context) (StackMetricSnapshot, error)

func (f StackSnapshotSourceFunc) Snapshot(ctx context.Context) (StackMetricSnapshot, error) {
	return f(ctx)
}

type databaseStackSnapshotSource struct {
	session db.SessionFactory
}

func NewDatabaseStackSnapshotSource(session db.SessionFactory) StackSnapshotSource {
	return &databaseStackSnapshotSource{session: session}
}

func (s *databaseStackSnapshotSource) Snapshot(ctx context.Context) (StackMetricSnapshot, error) {
	stackStates, err := s.queryCounts(ctx, stackStateCountsQuery)
	if err != nil {
		return StackMetricSnapshot{}, fmt.Errorf("load stack state counts: %w", err)
	}

	resourceStates, err := s.queryCounts(ctx, resourceStateCountsQuery)
	if err != nil {
		return StackMetricSnapshot{}, fmt.Errorf("load stack resource state counts: %w", err)
	}

	failureTypes, err := s.queryCounts(ctx, resourceFailureCountsQuery)
	if err != nil {
		return StackMetricSnapshot{}, fmt.Errorf("load stack resource failure counts: %w", err)
	}

	return StackMetricSnapshot{
		StackStates:    stackStates,
		ResourceStates: resourceStates,
		FailureTypes:   failureTypes,
	}, nil
}

type metricCountRow struct {
	Label string
	Count int64
}

func (s *databaseStackSnapshotSource) queryCounts(ctx context.Context, query string) (map[string]int64, error) {
	var rows []metricCountRow
	if err := s.session.New(ctx).Raw(query).Scan(&rows).Error; err != nil {
		return nil, err
	}

	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.Label] = row.Count
	}
	return counts, nil
}

const stackStateCountsQuery = `
	SELECT
		CASE status->>'state'
			WHEN 'Ready' THEN 'Ready'
			WHEN 'Deleting' THEN 'Deleting'
			WHEN 'Pending' THEN 'Pending'
			WHEN 'Failed' THEN 'Failed'
			WHEN 'Error' THEN 'Error'
			WHEN 'Degraded' THEN 'Degraded'
			WHEN 'Progressing' THEN 'Progressing'
			ELSE 'Unknown'
		END AS label,
		COUNT(*) AS count
	FROM stacks
	GROUP BY 1
`

const resourceStateCountsQuery = `
	SELECT
		CASE status->>'state'
			WHEN 'Pending' THEN 'Pending'
			WHEN 'Ready' THEN 'Ready'
			WHEN 'Failed' THEN 'Failed'
			WHEN 'Error' THEN 'Error'
			ELSE 'Unknown'
		END AS label,
		COUNT(*) AS count
	FROM stack_resources
	GROUP BY 1
`

const resourceFailureCountsQuery = `
	SELECT
		CASE status->'last_failure'->>'type'
			WHEN 'runtime_crash' THEN 'runtime_crash'
			WHEN 'build_failure' THEN 'build_failure'
			WHEN 'readiness_failure' THEN 'readiness_failure'
			ELSE 'Unknown'
		END AS label,
		COUNT(*) AS count
	FROM stack_resources
	WHERE COALESCE(status->>'state', '') <> 'Ready'
		AND status->>'last_failure' IS NOT NULL
	GROUP BY 1
`

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
	snapshot, err := c.source.Snapshot(ctx)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.success, prometheus.GaugeValue, 0)
		return
	}

	for state, count := range boundedCounts(snapshot.StackStates, func(label string) string {
		return boundedStackState(models.StackState(label))
	}) {
		ch <- prometheus.MustNewConstMetric(c.stacks, prometheus.GaugeValue, float64(count), state)
	}
	for state, count := range boundedCounts(snapshot.ResourceStates, func(label string) string {
		return boundedResourceState(models.StackResourceState(label))
	}) {
		ch <- prometheus.MustNewConstMetric(c.resources, prometheus.GaugeValue, float64(count), state)
	}
	for failureType, count := range boundedCounts(snapshot.FailureTypes, func(label string) string {
		return boundedFailureType(models.StackResourceFailureType(label))
	}) {
		ch <- prometheus.MustNewConstMetric(c.failures, prometheus.GaugeValue, float64(count), failureType)
	}
	ch <- prometheus.MustNewConstMetric(c.success, prometheus.GaugeValue, 1)
}

func boundedCounts(counts map[string]int64, bound func(string) string) map[string]int64 {
	bounded := make(map[string]int64, len(counts))
	for label, count := range counts {
		bounded[bound(label)] += count
	}
	return bounded
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
