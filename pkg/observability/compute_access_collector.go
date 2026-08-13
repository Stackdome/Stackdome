package observability

import (
	"context"
	"fmt"
	"time"

	"github.com/Stackdome/stackdome/pkg/computeaccess"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	ComputeEntitlementsIssuedMetricName    = "stackdome_cloud_compute_entitlements_issued"
	ComputeEntitlementsExpiredMetricName   = "stackdome_cloud_compute_entitlements_expired"
	SharedComputeLeasesMetricName          = "stackdome_cloud_shared_compute_leases"
	SharedComputeCapacityLimitMetricName   = "stackdome_cloud_shared_compute_capacity_limit"
	ComputeAccessSnapshotSuccessMetricName = "stackdome_cloud_compute_access_snapshot_collection_success"

	LabelSource = "source"

	UnknownSource = "Unknown"

	computeAccessSnapshotTimeout = 3 * time.Second
)

var supportedSharedComputeLeaseStates = []computeaccess.SharedComputeLeaseState{
	computeaccess.SharedComputeLeaseStateActive,
	computeaccess.SharedComputeLeaseStateCleanupPending,
	computeaccess.SharedComputeLeaseStateCleaning,
	computeaccess.SharedComputeLeaseStateCleaned,
	computeaccess.SharedComputeLeaseStateError,
}

type ComputeEntitlementMetricCounts struct {
	Issued  int64
	Expired int64
}

type ComputeAccessMetricSnapshot struct {
	EntitlementsBySource map[string]ComputeEntitlementMetricCounts
	LeaseStates          map[string]int64
}

type ComputeAccessSnapshotSource interface {
	Snapshot(context.Context) (ComputeAccessMetricSnapshot, error)
}

type ComputeAccessSnapshotSourceFunc func(context.Context) (ComputeAccessMetricSnapshot, error)

func (f ComputeAccessSnapshotSourceFunc) Snapshot(ctx context.Context) (ComputeAccessMetricSnapshot, error) {
	return f(ctx)
}

type databaseComputeAccessSnapshotSource struct {
	session db.SessionFactory
}

func NewDatabaseComputeAccessSnapshotSource(session db.SessionFactory) ComputeAccessSnapshotSource {
	return &databaseComputeAccessSnapshotSource{session: session}
}

func (s *databaseComputeAccessSnapshotSource) Snapshot(ctx context.Context) (ComputeAccessMetricSnapshot, error) {
	var entitlementRows []computeEntitlementMetricRow
	if err := s.session.New(ctx).Raw(computeEntitlementCountsQuery).Scan(&entitlementRows).Error; err != nil {
		return ComputeAccessMetricSnapshot{}, fmt.Errorf("load compute entitlement counts: %w", err)
	}

	leaseStates, err := s.queryLeaseStates(ctx)
	if err != nil {
		return ComputeAccessMetricSnapshot{}, fmt.Errorf("load shared compute lease state counts: %w", err)
	}

	entitlements := make(map[string]ComputeEntitlementMetricCounts, len(entitlementRows))
	for _, row := range entitlementRows {
		entitlements[row.Source] = ComputeEntitlementMetricCounts{
			Issued:  row.Issued,
			Expired: row.Expired,
		}
	}
	return ComputeAccessMetricSnapshot{
		EntitlementsBySource: entitlements,
		LeaseStates:          leaseStates,
	}, nil
}

type computeEntitlementMetricRow struct {
	Source  string
	Issued  int64
	Expired int64
}

func (s *databaseComputeAccessSnapshotSource) queryLeaseStates(ctx context.Context) (map[string]int64, error) {
	var rows []metricCountRow
	if err := s.session.New(ctx).Raw(sharedComputeLeaseStateCountsQuery).Scan(&rows).Error; err != nil {
		return nil, err
	}

	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.Label] = row.Count
	}
	return counts, nil
}

const computeEntitlementCountsQuery = `
	SELECT
		source,
		COUNT(*) AS issued,
		SUM(CASE
			WHEN expires_at IS NOT NULL AND expires_at <= CURRENT_TIMESTAMP THEN 1
			ELSE 0
		END) AS expired
	FROM compute_entitlements
	GROUP BY source
`

const sharedComputeLeaseStateCountsQuery = `
	SELECT state AS label, COUNT(*) AS count
	FROM shared_compute_leases
	GROUP BY state
`

type computeAccessCollector struct {
	source        ComputeAccessSnapshotSource
	capacityLimit float64
	entitlements  *prometheus.Desc
	expired       *prometheus.Desc
	leases        *prometheus.Desc
	capacity      *prometheus.Desc
	success       *prometheus.Desc
}

func newComputeAccessCollector(source ComputeAccessSnapshotSource, capacityLimit int) prometheus.Collector {
	return &computeAccessCollector{
		source:        source,
		capacityLimit: float64(capacityLimit),
		entitlements: prometheus.NewDesc(
			ComputeEntitlementsIssuedMetricName,
			"Total compute entitlements issued and retained, grouped by source.",
			[]string{LabelSource}, nil,
		),
		expired: prometheus.NewDesc(
			ComputeEntitlementsExpiredMetricName,
			"Current retained compute entitlements past their expiry time, grouped by source.",
			[]string{LabelSource}, nil,
		),
		leases: prometheus.NewDesc(
			SharedComputeLeasesMetricName,
			"Current shared-compute leases grouped by lifecycle state.",
			[]string{LabelState}, nil,
		),
		capacity: prometheus.NewDesc(
			SharedComputeCapacityLimitMetricName,
			"Configured maximum number of non-cleaned shared-compute leases.",
			nil, nil,
		),
		success: prometheus.NewDesc(
			ComputeAccessSnapshotSuccessMetricName,
			"Whether the latest cloud compute-access database snapshot succeeded.",
			nil, nil,
		),
	}
}

func (c *computeAccessCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.entitlements
	ch <- c.expired
	ch <- c.leases
	ch <- c.capacity
	ch <- c.success
}

func (c *computeAccessCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.capacity, prometheus.GaugeValue, c.capacityLimit)

	ctx, cancel := context.WithTimeout(context.Background(), computeAccessSnapshotTimeout)
	defer cancel()
	snapshot, err := c.source.Snapshot(ctx)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.success, prometheus.GaugeValue, 0)
		return
	}

	entitlements := boundedEntitlementCounts(snapshot.EntitlementsBySource)
	for source, counts := range entitlements {
		ch <- prometheus.MustNewConstMetric(c.entitlements, prometheus.GaugeValue, float64(counts.Issued), source)
		ch <- prometheus.MustNewConstMetric(c.expired, prometheus.GaugeValue, float64(counts.Expired), source)
	}
	for state, count := range boundedLeaseStateCounts(snapshot.LeaseStates) {
		ch <- prometheus.MustNewConstMetric(c.leases, prometheus.GaugeValue, float64(count), state)
	}
	ch <- prometheus.MustNewConstMetric(c.success, prometheus.GaugeValue, 1)
}

func boundedEntitlementCounts(counts map[string]ComputeEntitlementMetricCounts) map[string]ComputeEntitlementMetricCounts {
	bounded := map[string]ComputeEntitlementMetricCounts{
		string(computeaccess.ComputeEntitlementSourceTrial): {},
	}
	for source, count := range counts {
		label := boundedComputeEntitlementSource(computeaccess.ComputeEntitlementSource(source))
		current := bounded[label]
		current.Issued += count.Issued
		current.Expired += count.Expired
		bounded[label] = current
	}
	return bounded
}

func boundedComputeEntitlementSource(source computeaccess.ComputeEntitlementSource) string {
	if source == computeaccess.ComputeEntitlementSourceTrial {
		return string(source)
	}
	return UnknownSource
}

func boundedLeaseStateCounts(counts map[string]int64) map[string]int64 {
	bounded := make(map[string]int64, len(supportedSharedComputeLeaseStates)+1)
	for _, state := range supportedSharedComputeLeaseStates {
		bounded[string(state)] = 0
	}
	for state, count := range counts {
		bounded[boundedSharedComputeLeaseState(computeaccess.SharedComputeLeaseState(state))] += count
	}
	return bounded
}

func boundedSharedComputeLeaseState(state computeaccess.SharedComputeLeaseState) string {
	for _, supported := range supportedSharedComputeLeaseStates {
		if state == supported {
			return string(state)
		}
	}
	return UnknownState
}
