package migrations

import (
	"github.com/glebarez/sqlite"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = Describe("resource_validation_records stack foreign key migration", func() {
	openTables := func() *gorm.DB {
		database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		Expect(database.Exec(`CREATE TABLE stacks (id TEXT PRIMARY KEY)`).Error).NotTo(HaveOccurred())
		Expect(database.Exec(`
			CREATE TABLE resource_validation_records (
				stack_id TEXT NOT NULL,
				resource_name TEXT NOT NULL,
				check_kind TEXT NOT NULL,
				fingerprint TEXT NOT NULL,
				validated_at DATETIME NOT NULL,
				PRIMARY KEY (stack_id, resource_name, check_kind)
			)
		`).Error).NotTo(HaveOccurred())
		return database
	}

	It("removes only records whose stack no longer exists", func() {
		database := openTables()
		Expect(database.Exec(`INSERT INTO stacks (id) VALUES ('stack-1')`).Error).NotTo(HaveOccurred())
		Expect(database.Exec(`
			INSERT INTO resource_validation_records
				(stack_id, resource_name, check_kind, fingerprint, validated_at)
			VALUES
				('stack-1', 'web', 'image_pull', 'sha256:live', CURRENT_TIMESTAMP),
				('stack-deleted', 'web', 'image_pull', 'sha256:orphan', CURRENT_TIMESTAMP),
				('stack-deleted', 'worker', 'image_pull', 'sha256:orphan2', CURRENT_TIMESTAMP)
		`).Error).NotTo(HaveOccurred())

		Expect(deleteOrphanedResourceValidationRecords(database)).To(Succeed())

		var stackIDs []string
		Expect(database.Raw(`SELECT stack_id FROM resource_validation_records`).Scan(&stackIDs).Error).NotTo(HaveOccurred())
		Expect(stackIDs).To(Equal([]string{"stack-1"}))
	})
})
