package pgstore_test

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GitInstallationStore", func() {
	var (
		store stores.GitInstallationStore
		ctx   context.Context
	)

	BeforeEach(func() {
		sf := newSQLiteSessionFactory(`
			CREATE TABLE IF NOT EXISTS git_installations (
				id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
				git_integration_id TEXT NOT NULL,
				installation_id INTEGER NOT NULL,
				account_login TEXT NOT NULL,
				account_type TEXT,
				repository_selection TEXT,
				created_at DATETIME,
				updated_at DATETIME,
				UNIQUE (git_integration_id, installation_id)
			)
		`)
		store = pgstore.NewGitInstallationStore(pgstore.GitInstallationStoreSpec{SessionFactory: sf})
		ctx = context.Background()

		_, err := store.Upsert(ctx, &models.GitInstallation{
			GitIntegrationID:    "gi-1",
			InstallationID:      77,
			AccountLogin:        "acme",
			AccountType:         models.GitAccountTypeOrganization,
			RepositorySelection: "all",
		})
		Expect(err).To(BeNil())
	})

	It("upserts on the (integration, installation) key", func() {
		_, err := store.Upsert(ctx, &models.GitInstallation{
			GitIntegrationID:    "gi-1",
			InstallationID:      77,
			AccountLogin:        "acme-renamed",
			AccountType:         models.GitAccountTypeOrganization,
			RepositorySelection: "selected",
		})
		Expect(err).To(BeNil())

		installations, err := store.ListByIntegrationID(ctx, "gi-1")
		Expect(err).To(BeNil())
		Expect(installations).To(HaveLen(1))
		Expect(installations[0].AccountLogin).To(Equal("acme-renamed"))
		Expect(installations[0].RepositorySelection).To(Equal("selected"))
	})

	It("gets an installation by its global installation id", func() {
		installation, err := store.GetByInstallationID(ctx, 77)
		Expect(err).To(BeNil())
		Expect(installation.GitIntegrationID).To(Equal("gi-1"))
	})

	It("returns 404 for an unknown installation id", func() {
		_, err := store.GetByInstallationID(ctx, 999)
		Expect(err).ToNot(BeNil())
		Expect(err.Is404()).To(BeTrue())
	})

	It("matches accounts case-insensitively", func() {
		installation, err := store.GetByIntegrationAndAccount(ctx, "gi-1", "ACME")
		Expect(err).To(BeNil())
		Expect(installation.InstallationID).To(Equal(int64(77)))
	})

	It("gets an installation by its own id, scoped to the integration", func() {
		installations, err := store.ListByIntegrationID(ctx, "gi-1")
		Expect(err).To(BeNil())
		Expect(installations).To(HaveLen(1))
		id := installations[0].ID

		got, err := store.GetByIntegrationAndID(ctx, "gi-1", id)
		Expect(err).To(BeNil())
		Expect(got.InstallationID).To(Equal(int64(77)))

		_, err = store.GetByIntegrationAndID(ctx, "gi-1", "does-not-exist")
		Expect(err).ToNot(BeNil())
		Expect(err.Is404()).To(BeTrue())

		_, err = store.GetByIntegrationAndID(ctx, "other-integration", id)
		Expect(err).ToNot(BeNil())
		Expect(err.Is404()).To(BeTrue())
	})

	It("deletes by installation id", func() {
		Expect(store.DeleteByInstallationID(ctx, "gi-1", 77)).To(BeNil())
		installations, err := store.ListByIntegrationID(ctx, "gi-1")
		Expect(err).To(BeNil())
		Expect(installations).To(BeEmpty())
	})
})
