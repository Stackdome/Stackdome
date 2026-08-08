package observability

import (
	"context"
	"database/sql"
	stderrors "errors"
	"time"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/glebarez/sqlite"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

var _ = Describe("stack snapshot metrics", func() {
	It("groups current stack, resource, and active failure states", func() {
		source := StackSnapshotSourceFunc(func(context.Context) (StackMetricSnapshot, error) {
			return StackMetricSnapshot{
				StackStates: map[string]int64{
					string(models.StackReady):  1,
					string(models.StackFailed): 2,
					"":                         1,
					"unexpected-stack-state":   1,
				},
				ResourceStates: map[string]int64{
					string(models.StackResourcePhaseReady):   1,
					string(models.StackResourcePhaseFailed):  1,
					string(models.StackResourcePhaseUnknown): 1,
					"":                                       1,
					"unexpected-resource-state":              1,
				},
				FailureTypes: map[string]int64{
					string(models.FailureTypeRuntimeCrash):     1,
					string(models.FailureTypeReadinessFailure): 1,
					"unexpected-failure":                       1,
				},
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
		source := StackSnapshotSourceFunc(func(context.Context) (StackMetricSnapshot, error) {
			return StackMetricSnapshot{}, stderrors.New("database unavailable")
		})
		registry := prometheus.NewRegistry()
		registry.MustRegister(newStackCollector(source))

		families, err := registry.Gather()
		Expect(err).NotTo(HaveOccurred())
		Expect(familyNamed(families, StacksMetricName)).To(BeNil())
		Expect(metricValue(familyNamed(families, StackSnapshotSuccessMetricName), nil)).To(Equal(float64(0)))
	})

	It("includes stacks with deletion intent and bounds database status labels", func() {
		session := newStackCollectorTestSession()
		DeferCleanup(session.Close)

		deletionStartedAt := time.Now().UTC()
		Expect(session.New(context.Background()).Exec(`
			INSERT INTO stacks (id, status, deletion_timestamp) VALUES
				('ready', CAST('{"state":"Ready"}' AS BLOB), NULL),
				('deleting', CAST('{"state":"Deleting"}' AS BLOB), ?),
				('unexpected', CAST('{"state":"brand-new-state"}' AS BLOB), NULL),
				('missing-status', NULL, NULL)
		`, deletionStartedAt).Error).NotTo(HaveOccurred())
		Expect(session.New(context.Background()).Exec(`
			INSERT INTO stack_resources (id, status) VALUES
				('ready', CAST('{"state":"Ready","last_failure":{"type":"build_failure"}}' AS BLOB)),
				('failed', CAST('{"state":"Failed","last_failure":{"type":"runtime_crash"}}' AS BLOB)),
				('unexpected', CAST('{"state":"brand-new-state","last_failure":{"type":"brand-new-failure"}}' AS BLOB)),
				('missing-status', NULL)
		`).Error).NotTo(HaveOccurred())

		registry := prometheus.NewRegistry()
		registry.MustRegister(newStackCollector(NewDatabaseStackSnapshotSource(session)))

		families, err := registry.Gather()
		Expect(err).NotTo(HaveOccurred())
		Expect(metricValue(familyNamed(families, StacksMetricName), map[string]string{LabelState: string(models.StackReady)})).To(Equal(float64(1)))
		Expect(metricValue(familyNamed(families, StacksMetricName), map[string]string{LabelState: string(models.StackDeleting)})).To(Equal(float64(1)))
		Expect(metricValue(familyNamed(families, StacksMetricName), map[string]string{LabelState: UnknownState})).To(Equal(float64(2)))
		Expect(metricValue(familyNamed(families, StackResourcesMetricName), map[string]string{LabelState: string(models.StackResourcePhaseReady)})).To(Equal(float64(1)))
		Expect(metricValue(familyNamed(families, StackResourcesMetricName), map[string]string{LabelState: string(models.StackResourcePhaseFailed)})).To(Equal(float64(1)))
		Expect(metricValue(familyNamed(families, StackResourcesMetricName), map[string]string{LabelState: UnknownState})).To(Equal(float64(2)))
		Expect(metricValue(familyNamed(families, StackFailuresMetricName), map[string]string{LabelType: string(models.FailureTypeRuntimeCrash)})).To(Equal(float64(1)))
		Expect(metricValue(familyNamed(families, StackFailuresMetricName), map[string]string{LabelType: UnknownState})).To(Equal(float64(1)))
		Expect(familyNamed(families, StackFailuresMetricName).Metric).To(HaveLen(2), "ready resources must not expose historical failures")
	})
})

type stackCollectorTestSession struct {
	database *gorm.DB
}

func newStackCollectorTestSession() *stackCollectorTestSession {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	Expect(err).NotTo(HaveOccurred())
	Expect(database.Exec(`
		CREATE TABLE stacks (
			id TEXT PRIMARY KEY,
			status BLOB,
			deletion_timestamp DATETIME
		)
	`).Error).NotTo(HaveOccurred())
	Expect(database.Exec(`
		CREATE TABLE stack_resources (
			id TEXT PRIMARY KEY,
			status BLOB
		)
	`).Error).NotTo(HaveOccurred())
	return &stackCollectorTestSession{database: database}
}

func (s *stackCollectorTestSession) Init(*config.DatabaseConfig) {}

func (s *stackCollectorTestSession) DirectDB() *sql.DB {
	database, _ := s.database.DB()
	return database
}

func (s *stackCollectorTestSession) New(ctx context.Context) *gorm.DB {
	if transaction := db.TxFromContext(ctx); transaction != nil {
		return transaction
	}
	return s.database.WithContext(ctx)
}

func (s *stackCollectorTestSession) CheckConnection() error { return nil }

func (s *stackCollectorTestSession) Close() error {
	database, _ := s.database.DB()
	return database.Close()
}
