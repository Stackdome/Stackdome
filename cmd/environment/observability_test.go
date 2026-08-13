package environment

import (
	"context"
	"database/sql"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/observability"
	"github.com/glebarez/sqlite"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dto "github.com/prometheus/client_model/go"
	"gorm.io/gorm"
)

var _ = Describe("compute access metrics registration", func() {
	It("registers cloud compute access metrics with the configured capacity limit in cloud mode", func() {
		session := newObservabilityTestSession()
		DeferCleanup(session.Close)
		applicationConfig := config.NewApplicationConfig()
		applicationConfig.RuntimeMode = config.RuntimeModeStackdomeCloud
		applicationConfig.StackdomeCloud = &config.StackdomeCloudConfig{
			Access: config.StackdomeCloudComputeAccessConfig{MaxActiveSharedComputeLeases: 12},
		}
		testEnv := NewTestEnvironment(session, WithApplicationConfig(applicationConfig)).(*environmentImpl)

		Expect(testEnv.setupObservability(context.Background())).To(Succeed())
		families, err := testEnv.Observability.Gatherer().Gather()
		Expect(err).NotTo(HaveOccurred())
		Expect(environmentMetricValue(
			environmentFamilyNamed(families, observability.SharedComputeCapacityLimitMetricName),
		)).To(Equal(float64(12)))
		Expect(environmentFamilyNamed(families, observability.ComputeEntitlementsIssuedMetricName)).NotTo(BeNil())
		Expect(environmentFamilyNamed(families, observability.ComputeEntitlementsExpiredMetricName)).NotTo(BeNil())
		Expect(environmentFamilyNamed(families, observability.SharedComputeLeasesMetricName)).NotTo(BeNil())
		Expect(environmentMetricValue(
			environmentFamilyNamed(families, observability.ComputeAccessSnapshotSuccessMetricName),
		)).To(Equal(float64(1)))
	})

	It("does not register cloud compute access metrics in self-hosted mode", func() {
		session := newObservabilityTestSession()
		DeferCleanup(session.Close)
		applicationConfig := config.NewApplicationConfig()
		applicationConfig.RuntimeMode = config.RuntimeModeSelfHosted
		testEnv := NewTestEnvironment(session, WithApplicationConfig(applicationConfig)).(*environmentImpl)

		Expect(testEnv.setupObservability(context.Background())).To(Succeed())
		families, err := testEnv.Observability.Gatherer().Gather()
		Expect(err).NotTo(HaveOccurred())
		for _, name := range []string{
			observability.ComputeEntitlementsIssuedMetricName,
			observability.ComputeEntitlementsExpiredMetricName,
			observability.SharedComputeLeasesMetricName,
			observability.SharedComputeCapacityLimitMetricName,
			observability.ComputeAccessSnapshotSuccessMetricName,
		} {
			Expect(environmentFamilyNamed(families, name)).To(BeNil(), "metric %s must not exist", name)
		}
	})
})

type observabilityTestSession struct {
	database *gorm.DB
}

func newObservabilityTestSession() *observabilityTestSession {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	Expect(err).NotTo(HaveOccurred())
	for _, statement := range []string{
		`CREATE TABLE stacks (id TEXT PRIMARY KEY, status BLOB, deletion_timestamp DATETIME)`,
		`CREATE TABLE stack_resources (id TEXT PRIMARY KEY, status BLOB)`,
		`CREATE TABLE compute_entitlements (id TEXT PRIMARY KEY, source TEXT NOT NULL, expires_at DATETIME)`,
		`CREATE TABLE shared_compute_leases (id TEXT PRIMARY KEY, state TEXT NOT NULL)`,
	} {
		Expect(database.Exec(statement).Error).NotTo(HaveOccurred())
	}
	return &observabilityTestSession{database: database}
}

func (s *observabilityTestSession) Init(*config.DatabaseConfig) {}

func (s *observabilityTestSession) DirectDB() *sql.DB {
	database, _ := s.database.DB()
	return database
}

func (s *observabilityTestSession) New(ctx context.Context) *gorm.DB {
	if transaction := db.TxFromContext(ctx); transaction != nil {
		return transaction
	}
	return s.database.WithContext(ctx)
}

func (s *observabilityTestSession) CheckConnection() error { return nil }

func (s *observabilityTestSession) Close() error {
	database, _ := s.database.DB()
	return database.Close()
}

func environmentFamilyNamed(families []*dto.MetricFamily, name string) *dto.MetricFamily {
	for _, family := range families {
		if family.GetName() == name {
			return family
		}
	}
	return nil
}

func environmentMetricValue(family *dto.MetricFamily) float64 {
	Expect(family).NotTo(BeNil())
	Expect(family.Metric).To(HaveLen(1))
	return family.Metric[0].GetGauge().GetValue()
}
