package observability

import (
	"context"
	stderrors "errors"

	"github.com/Stackdome/stackdome/pkg/computeaccess"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
)

var _ = Describe("cloud compute access snapshot metrics", func() {
	It("publishes entitlement, lease, and configured capacity gauges with bounded labels", func() {
		source := ComputeAccessSnapshotSourceFunc(func(context.Context) (ComputeAccessMetricSnapshot, error) {
			return ComputeAccessMetricSnapshot{
				EntitlementsBySource: map[string]ComputeEntitlementMetricCounts{
					string(computeaccess.ComputeEntitlementSourceTrial): {Issued: 4, Expired: 2},
					"unexpected-source": {Issued: 1, Expired: 1},
				},
				LeaseStates: map[string]int64{
					string(computeaccess.SharedComputeLeaseStateActive):         2,
					string(computeaccess.SharedComputeLeaseStateCleanupPending): 1,
					string(computeaccess.SharedComputeLeaseStateCleaning):       1,
					string(computeaccess.SharedComputeLeaseStateCleaned):        3,
					string(computeaccess.SharedComputeLeaseStateError):          1,
					"unexpected-state": 1,
				},
			}, nil
		})
		registry := prometheus.NewRegistry()
		registry.MustRegister(newComputeAccessCollector(source, 10))

		families, err := registry.Gather()
		Expect(err).NotTo(HaveOccurred())
		Expect(metricValue(familyNamed(families, ComputeEntitlementsIssuedMetricName), map[string]string{
			LabelSource: string(computeaccess.ComputeEntitlementSourceTrial),
		})).To(Equal(float64(4)))
		Expect(metricValue(familyNamed(families, ComputeEntitlementsIssuedMetricName), map[string]string{
			LabelSource: UnknownSource,
		})).To(Equal(float64(1)))
		Expect(metricValue(familyNamed(families, ComputeEntitlementsExpiredMetricName), map[string]string{
			LabelSource: string(computeaccess.ComputeEntitlementSourceTrial),
		})).To(Equal(float64(2)))
		Expect(metricValue(familyNamed(families, SharedComputeLeasesMetricName), map[string]string{
			LabelState: string(computeaccess.SharedComputeLeaseStateActive),
		})).To(Equal(float64(2)))
		Expect(metricValue(familyNamed(families, SharedComputeLeasesMetricName), map[string]string{
			LabelState: UnknownState,
		})).To(Equal(float64(1)))
		Expect(metricValue(familyNamed(families, SharedComputeCapacityLimitMetricName), nil)).To(Equal(float64(10)))
		Expect(metricValue(familyNamed(families, ComputeAccessSnapshotSuccessMetricName), nil)).To(Equal(float64(1)))
	})

	It("reports collection failure without publishing database-derived gauges", func() {
		source := ComputeAccessSnapshotSourceFunc(func(context.Context) (ComputeAccessMetricSnapshot, error) {
			return ComputeAccessMetricSnapshot{}, stderrors.New("database unavailable")
		})
		registry := prometheus.NewRegistry()
		registry.MustRegister(newComputeAccessCollector(source, 10))

		families, err := registry.Gather()
		Expect(err).NotTo(HaveOccurred())
		Expect(familyNamed(families, ComputeEntitlementsIssuedMetricName)).To(BeNil())
		Expect(familyNamed(families, ComputeEntitlementsExpiredMetricName)).To(BeNil())
		Expect(familyNamed(families, SharedComputeLeasesMetricName)).To(BeNil())
		Expect(metricValue(familyNamed(families, SharedComputeCapacityLimitMetricName), nil)).To(Equal(float64(10)))
		Expect(metricValue(familyNamed(families, ComputeAccessSnapshotSuccessMetricName), nil)).To(Equal(float64(0)))
	})

	It("publishes zero values for every currently supported source and lease state", func() {
		source := ComputeAccessSnapshotSourceFunc(func(context.Context) (ComputeAccessMetricSnapshot, error) {
			return ComputeAccessMetricSnapshot{}, nil
		})
		registry := prometheus.NewRegistry()
		registry.MustRegister(newComputeAccessCollector(source, 10))

		families, err := registry.Gather()
		Expect(err).NotTo(HaveOccurred())
		Expect(metricValue(familyNamed(families, ComputeEntitlementsIssuedMetricName), map[string]string{
			LabelSource: string(computeaccess.ComputeEntitlementSourceTrial),
		})).To(BeZero())
		Expect(metricValue(familyNamed(families, ComputeEntitlementsExpiredMetricName), map[string]string{
			LabelSource: string(computeaccess.ComputeEntitlementSourceTrial),
		})).To(BeZero())
		for _, state := range []computeaccess.SharedComputeLeaseState{
			computeaccess.SharedComputeLeaseStateActive,
			computeaccess.SharedComputeLeaseStateCleanupPending,
			computeaccess.SharedComputeLeaseStateCleaning,
			computeaccess.SharedComputeLeaseStateCleaned,
			computeaccess.SharedComputeLeaseStateError,
		} {
			Expect(metricValue(familyNamed(families, SharedComputeLeasesMetricName), map[string]string{
				LabelState: string(state),
			})).To(BeZero())
		}
	})

	It("loads retained entitlements, database-expired entitlements, and lease states", func() {
		session := newComputeAccessCollectorTestSession()
		DeferCleanup(session.Close)
		database := session.New(context.Background())
		Expect(database.Exec(`
			INSERT INTO compute_entitlements (id, source, expires_at) VALUES
				('expired-trial', 'trial', '2000-01-01 00:00:00'),
				('current-trial', 'trial', '2100-01-01 00:00:00'),
				('unlimited-trial', 'trial', NULL)
		`).Error).NotTo(HaveOccurred())
		Expect(database.Exec(`
			INSERT INTO shared_compute_leases (id, state) VALUES
				('active-1', 'active'),
				('active-2', 'active'),
				('cleanup-pending', 'cleanup_pending'),
				('cleaned', 'cleaned')
		`).Error).NotTo(HaveOccurred())

		snapshot, err := NewDatabaseComputeAccessSnapshotSource(session).Snapshot(context.Background())

		Expect(err).NotTo(HaveOccurred())
		Expect(snapshot.EntitlementsBySource).To(Equal(map[string]ComputeEntitlementMetricCounts{
			string(computeaccess.ComputeEntitlementSourceTrial): {Issued: 3, Expired: 1},
		}))
		Expect(snapshot.LeaseStates).To(Equal(map[string]int64{
			string(computeaccess.SharedComputeLeaseStateActive):         2,
			string(computeaccess.SharedComputeLeaseStateCleanupPending): 1,
			string(computeaccess.SharedComputeLeaseStateCleaned):        1,
		}))
	})
})

func newComputeAccessCollectorTestSession() *stackCollectorTestSession {
	session := newStackCollectorTestSession()
	database := session.New(context.Background())
	Expect(database.Exec(`
		CREATE TABLE compute_entitlements (
			id TEXT PRIMARY KEY,
			source TEXT NOT NULL,
			expires_at DATETIME
		)
	`).Error).NotTo(HaveOccurred())
	Expect(database.Exec(`
		CREATE TABLE shared_compute_leases (
			id TEXT PRIMARY KEY,
			state TEXT NOT NULL
		)
	`).Error).NotTo(HaveOccurred())
	return session
}
