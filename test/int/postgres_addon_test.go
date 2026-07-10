package int

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/test/int/shared"
)

var _ = Describe("PostgreSQL Addon", func() {
	var client *openapi.APIClient
	var orgID string
	var teamName = models.DefaultTeamName

	BeforeEach(func() {
		testEnv := GetEnvironment()
		Expect(testEnv).NotTo(BeNil(), "Test environment should be initialized")

		client = testEnv.Client
		orgID = testEnv.OrgID
	})

	Context("CRUD Operations", func() {
		It("should create a minimal postgres addon", func() {
			By("Creating a minimal postgres addon")
			addon := shared.CreateMinimalPostgresAddon("test-minimal")

			createdAddon := shared.CreatePostgresAddon(client, orgID, teamName, addon)

			Expect(createdAddon.GetId()).NotTo(BeEmpty())
			Expect(createdAddon.GetName()).To(Equal("test-minimal"))
			Expect(createdAddon.Spec.Version.GetMajor()).To(Equal(int32(16)))
			Expect(createdAddon.Spec.Instances.GetCount()).To(Equal(int32(2)))
			Expect(createdAddon.Spec.Storage.GetSize()).To(Equal("5Gi"))
			Expect(createdAddon.Spec.Storage.GetStorageClass()).To(Equal("standard"))
		})

		It("should create a postgres addon with resources", func() {
			By("Creating a postgres addon with CPU and memory resources")
			addon := shared.CreatePostgresAddonWithResources("test-with-resources")

			createdAddon := shared.CreatePostgresAddon(client, orgID, teamName, addon)

			Expect(createdAddon.GetId()).NotTo(BeEmpty())
			Expect(createdAddon.Spec.HasResources()).To(BeTrue())

			resources := createdAddon.Spec.GetResources()
			Expect(resources.HasCpu()).To(BeTrue())
			Expect(resources.HasMemory()).To(BeTrue())

			Expect(resources.Cpu.GetRequest()).To(Equal("250m"))
			Expect(resources.Cpu.GetLimit()).To(Equal("1"))
			Expect(resources.Memory.GetRequest()).To(Equal("512Mi"))
			Expect(resources.Memory.GetLimit()).To(Equal("2Gi"))

			Expect(createdAddon.Spec.HasDatabases()).To(BeTrue())
			databases := createdAddon.Spec.GetDatabases()
			Expect(len(databases)).To(Equal(2))
			dbNames := []string{databases[0].GetName(), databases[1].GetName()}
			Expect(dbNames).To(ContainElements("testdb", "app"))
		})

		It("should retrieve a postgres addon by ID", func() {
			By("Creating a postgres addon first")
			addon := shared.CreateMinimalPostgresAddon("test-get")
			createdAddon := shared.CreatePostgresAddon(client, orgID, teamName, addon)

			By("Retrieving the addon by ID")
			retrievedAddon := shared.GetPostgresAddon(client, orgID, teamName, createdAddon.GetId())

			shared.ExpectPostgresAddonEqual(createdAddon, retrievedAddon)
		})

		It("should list postgres addons", func() {
			By("Creating multiple postgres addons")
			addon1 := shared.CreateMinimalPostgresAddon("test-list-1")
			addon2 := shared.CreateMinimalPostgresAddon("test-list-2")

			shared.CreatePostgresAddon(client, orgID, teamName, addon1)
			shared.CreatePostgresAddon(client, orgID, teamName, addon2)

			By("Listing all postgres addons")
			addonList := shared.ListPostgresAddons(client, orgID, teamName)

			Expect(len(addonList.GetItems())).To(BeNumerically(">=", 2))

			// Find our created addons in the list
			var foundNames []string
			for _, addon := range addonList.GetItems() {
				foundNames = append(foundNames, addon.GetName())
			}
			Expect(foundNames).To(ContainElement("test-list-1"))
			Expect(foundNames).To(ContainElement("test-list-2"))
		})

		It("should update a postgres addon", func() {
			By("Creating a postgres addon first")
			addon := shared.CreateMinimalPostgresAddon("test-update")
			createdAddon := shared.CreatePostgresAddon(client, orgID, teamName, addon)

			By("Updating the addon with more resources and databases")
			updateAddon := shared.CreatePostgresAddonForUpdate("test-update")

			updatedAddon := shared.UpdatePostgresAddon(client, orgID, teamName, createdAddon.GetId(), updateAddon)

			Expect(updatedAddon.GetId()).To(Equal(createdAddon.GetId()))
			Expect(updatedAddon.Spec.Instances.GetCount()).To(Equal(int32(3)))
			Expect(updatedAddon.Spec.HasResources()).To(BeTrue())
			Expect(updatedAddon.Spec.HasDatabases()).To(BeTrue())

			databases := updatedAddon.Spec.GetDatabases()
			Expect(len(databases)).To(Equal(2))

			dbNames := []string{databases[0].GetName(), databases[1].GetName()}
			Expect(dbNames).To(ContainElement("app"))
			Expect(dbNames).To(ContainElement("analytics"))
		})

		It("should delete a postgres addon", func() {
			By("Creating a postgres addon first")
			addon := shared.CreateMinimalPostgresAddon("test-delete")
			createdAddon := shared.CreatePostgresAddon(client, orgID, teamName, addon)

			By("Deleting the addon")
			shared.DeletePostgresAddon(client, orgID, teamName, createdAddon.GetId())

			By("Waiting for the addon to be fully deleted")
			shared.WaitForAddonDeleted(client, orgID, teamName, createdAddon.GetId(), 30*time.Second)
		})
	})

	Context("Validation", func() {
		It("should reject addon with invalid storage size", func() {
			By("Creating addon with invalid storage size")
			addon := shared.CreateMinimalPostgresAddon("test-invalid-storage")
			addon.Spec.Storage.SetSize("invalid-size")

			_ = shared.CreatePostgresAddonExpectError(client, orgID, teamName, addon, 400)
		})

		It("should reject addon with zero instances", func() {
			By("Creating addon with zero instances")
			addon := shared.CreateMinimalPostgresAddon("test-zero-instances")
			addon.Spec.Instances.SetCount(0)

			_ = shared.CreatePostgresAddonExpectError(client, orgID, teamName, addon, 400)
		})

		It("should reject addon with invalid postgres version", func() {
			By("Creating addon with invalid postgres version")
			addon := shared.CreateMinimalPostgresAddon("test-invalid-version")
			addon.Spec.Version.SetMajor(99) // Invalid version

			_ = shared.CreatePostgresAddonExpectError(client, orgID, teamName, addon, 400)
		})
	})

})

var _ = Describe("PostgreSQL Addon Advanced Features", func() {
	var client *openapi.APIClient
	var orgID string
	var teamName = models.DefaultTeamName

	BeforeEach(func() {
		testEnv := GetEnvironment()
		Expect(testEnv).NotTo(BeNil(), "Test environment should be initialized")

		client = testEnv.Client
		orgID = testEnv.OrgID

		By("Clearing test data")
		Expect(testEnv.Database.ClearData(context.Background())).To(Succeed())
	})

	Context("Backup and Restore", func() {
		It("should create addon with backup configuration", func() {
			By("Creating postgres addon with backup config")
			addon := shared.CreatePostgresAddonWithBackup("test-backup")

			createdAddon := shared.CreatePostgresAddon(client, orgID, teamName, addon)

			Expect(createdAddon.Spec.HasBackup()).To(BeTrue())
		})
	})

	Context("High Availability", func() {
		It("should create HA postgres addon", func() {
			By("Creating HA postgres addon with multiple instances")
			addon := shared.CreatePostgresAddonWithHA("test-ha")

			createdAddon := shared.CreatePostgresAddon(client, orgID, teamName, addon)

			Expect(createdAddon.Spec.Instances.GetCount()).To(Equal(int32(3)))
			Expect(createdAddon.Spec.Instances.HasPlacement()).To(BeTrue())

			placement := createdAddon.Spec.Instances.GetPlacement()
			Expect(placement.GetTopologyKey()).To(Equal("kubernetes.io/zone"))
		})
	})
})
