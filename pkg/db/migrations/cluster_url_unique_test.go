package migrations

import (
	"github.com/glebarez/sqlite"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = Describe("Shared-compute cluster migration", func() {
	openClustersTable := func() *gorm.DB {
		database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		Expect(database.Exec(`
			CREATE TABLE clusters (
				id TEXT PRIMARY KEY,
				cluster_url TEXT NOT NULL,
				platform BOOLEAN NOT NULL DEFAULT FALSE
			)
		`).Error).NotTo(HaveOccurred())
		return database
	}

	It("fails with the duplicate URL and count before adding the unique index", func() {
		database := openClustersTable()
		Expect(database.Exec(`
			INSERT INTO clusters (id, cluster_url, platform) VALUES
				('cluster-1', 'https://duplicate.example.com', TRUE),
				('cluster-2', 'https://duplicate.example.com', FALSE)
		`).Error).NotTo(HaveOccurred())

		err := renameClusterPlatformToSharedCompute().Migrate(database)

		Expect(err).To(MatchError(ContainSubstring(
			`cannot enforce unique clusters.cluster_url: URL "https://duplicate.example.com" is used by 2 clusters`,
		)))
	})

	It("preserves the renamed flag and enforces globally unique URLs", func() {
		database := openClustersTable()
		Expect(database.Exec(`
			INSERT INTO clusters (id, cluster_url, platform) VALUES
				('cluster-1', 'https://one.example.com', TRUE),
				('cluster-2', 'https://two.example.com', FALSE)
		`).Error).NotTo(HaveOccurred())

		Expect(renameClusterPlatformToSharedCompute().Migrate(database)).To(Succeed())

		var sharedCompute bool
		Expect(database.Raw(`SELECT shared_compute FROM clusters WHERE id = 'cluster-1'`).Scan(&sharedCompute).Error).NotTo(HaveOccurred())
		Expect(sharedCompute).To(BeTrue())
		Expect(database.Exec(`
			INSERT INTO clusters (id, cluster_url, shared_compute)
			VALUES ('cluster-3', 'https://one.example.com', FALSE)
		`).Error).To(HaveOccurred())
	})
})
