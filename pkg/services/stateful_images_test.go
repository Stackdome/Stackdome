package services

import (
	"github.com/Stackdome/stackdome/pkg/models"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("applyStatefulWorkloadDefault", func() {
	resourceWith := func(image string, workloadType models.WorkloadType) *models.StackResource {
		return &models.StackResource{
			ImageConfig:  &models.ImageConfigSpec{Image: image},
			WorkloadType: workloadType,
		}
	}

	ginkgo.DescribeTable("promotes a datastore image",
		func(image string) {
			resource := resourceWith(image, models.WorkloadTypeService)

			applyStatefulWorkloadDefault(resource)

			gomega.Expect(resource.WorkloadType).To(gomega.Equal(models.WorkloadTypeStatefulService))
		},
		ginkgo.Entry("bare name", "postgres"),
		ginkgo.Entry("with tag", "postgres:16"),
		ginkgo.Entry("fully qualified", "docker.io/library/postgres:16"),
		ginkgo.Entry("with digest", "postgres@sha256:abc123"),
		ginkgo.Entry("private registry", "ghcr.io/acme/redis:7"),
		ginkgo.Entry("registry with port", "registry.internal:5000/mysql:8"),
		ginkgo.Entry("vendor prefixed", "bitnami/postgresql:15"),
		ginkgo.Entry("mixed case", "Postgres:16"),
		ginkgo.Entry("product name in the penultimate segment", "mcr.microsoft.com/mssql/server:2022-latest"),
	)

	// Every datastore offered by the frontend block catalog and template
	// browser, so the list cannot drift away from what users can actually pick.
	ginkgo.DescribeTable("promotes everything the frontend offers as a data store",
		func(image string) {
			resource := resourceWith(image, models.WorkloadTypeService)

			applyStatefulWorkloadDefault(resource)

			gomega.Expect(resource.WorkloadType).To(gomega.Equal(models.WorkloadTypeStatefulService))
		},
		ginkgo.Entry("blocks: Postgres", "postgres:16"),
		ginkgo.Entry("blocks: Redis", "redis:7"),
		ginkgo.Entry("blocks: MySQL", "mysql:8"),
		ginkgo.Entry("blocks: MongoDB", "mongo:7"),
		ginkgo.Entry("blocks: MariaDB", "mariadb:11.4"),
		ginkgo.Entry("blocks: MS SQL Server", "mcr.microsoft.com/mssql/server:2022-latest"),
		ginkgo.Entry("blocks: Elasticsearch", "docker.elastic.co/elasticsearch/elasticsearch:8.15.0"),
		ginkgo.Entry("blocks: CouchDB", "couchdb:3.3"),
		ginkgo.Entry("blocks: InfluxDB", "influxdb:2.7"),
		ginkgo.Entry("blocks: ClickHouse", "clickhouse/clickhouse-server:24.8"),
		ginkgo.Entry("template: grafana", "grafana/grafana:13.0.2"),
		ginkgo.Entry("template: prometheus", "prom/prometheus:v3.12.0"),
		ginkgo.Entry("template: gitea", "gitea/gitea:1.26.4"),
		ginkgo.Entry("template: immich valkey", "valkey/valkey:9"),
		ginkgo.Entry("template: immich postgres", "ghcr.io/immich-app/postgres:14-vectorchord0.4.3-pgvectors0.2.0"),
		ginkgo.Entry("template: tooljet otel-lgtm", "grafana/otel-lgtm:0.29.1"),
	)

	ginkgo.It("promotes when no workload type was chosen", func() {
		resource := resourceWith("postgres:16", "")

		applyStatefulWorkloadDefault(resource)

		gomega.Expect(resource.WorkloadType).To(gomega.Equal(models.WorkloadTypeStatefulService))
	})

	ginkgo.DescribeTable("leaves an application image alone",
		func(image string) {
			resource := resourceWith(image, models.WorkloadTypeService)

			applyStatefulWorkloadDefault(resource)

			gomega.Expect(resource.WorkloadType).To(gomega.Equal(models.WorkloadTypeService))
		},
		ginkgo.Entry("n8n keeps its zero-downtime rollout", "n8nio/n8n:2.27.4"),
		ginkgo.Entry("metrics scraper is not a datastore", "oliver006/redis-exporter:v1.55"),
		ginkgo.Entry("suffixed name does not match", "acme/postgres-backup:1.0"),
		ginkgo.Entry("prefixed name does not match", "acme/my-postgres:1.0"),
		ginkgo.Entry("postgrest is a stateless API, not postgres", "postgrest/postgrest:v12.2.0"),
		ginkgo.Entry("a generic penultimate segment does not match", "acme/server:1.0"),
		ginkgo.Entry("template: tooljet", "tooljet/tooljet:v3.20.189-lts"),
		ginkgo.Entry("template: immich server", "ghcr.io/immich-app/immich-server:v2.7.5"),
		ginkgo.Entry("template: openclaw", "ghcr.io/openclaw/openclaw:2026.6.10"),
		ginkgo.Entry("unrelated image", "nginx:1.27"),
	)

	ginkgo.It("leaves a git-built resource alone", func() {
		resource := &models.StackResource{WorkloadType: models.WorkloadTypeService}

		applyStatefulWorkloadDefault(resource)

		gomega.Expect(resource.WorkloadType).To(gomega.Equal(models.WorkloadTypeService))
	})

	ginkgo.DescribeTable("leaves an explicitly chosen workload type alone",
		func(workloadType models.WorkloadType) {
			resource := resourceWith("postgres:16", workloadType)

			applyStatefulWorkloadDefault(resource)

			gomega.Expect(resource.WorkloadType).To(gomega.Equal(workloadType))
		},
		ginkgo.Entry("Worker", models.WorkloadTypeWorker),
		ginkgo.Entry("Job", models.WorkloadTypeJob),
		ginkgo.Entry("CronJob", models.WorkloadTypeCronJob),
		ginkgo.Entry("StatefulService", models.WorkloadTypeStatefulService),
	)
})
