package pgstore_test

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GitIntegrationStore", func() {
	var (
		store stores.GitIntegrationStore
		ctx   context.Context
	)

	const (
		orgID  = "org-1"
		orgID2 = "org-2"
	)

	BeforeEach(func() {
		sf := newSQLiteSessionFactory(`
			CREATE TABLE IF NOT EXISTS git_integrations (
				id TEXT PRIMARY KEY,
				organisation_id TEXT NOT NULL,
				type TEXT NOT NULL,
				host TEXT,
				encrypted_auth TEXT,
				status TEXT NOT NULL DEFAULT 'active',
				data_hash TEXT NOT NULL,
				created_at DATETIME,
				updated_at DATETIME
			)
		`)
		store = pgstore.NewGitIntegrationStore(pgstore.GitIntegrationStoreSpec{SessionFactory: sf})
		ctx = context.Background()

		integrations := []models.GitIntegration{
			{
				ID:             "gi-1",
				OrganisationID: orgID,
				Type:           models.GitIntegrationTypeGitCredentials,
				Host:           "gitlab.example.com",
				Status:         models.GitIntegrationStatusActive,
				EncryptedAuth:  "enc-1",
				DataHash:       "hash-1",
			},
			{
				ID:             "gi-2",
				OrganisationID: orgID,
				Type:           models.GitIntegrationTypeGitHubApp,
				Host:           "github.com",
				Status:         models.GitIntegrationStatusPendingInstall,
				EncryptedAuth:  "enc-2",
				DataHash:       "hash-2",
			},
			{
				ID:             "gi-3",
				OrganisationID: orgID2,
				Type:           models.GitIntegrationTypeGitCredentials,
				Host:           "gitlab.example.com",
				Status:         models.GitIntegrationStatusActive,
				EncryptedAuth:  "enc-3",
				DataHash:       "hash-3",
			},
		}
		for i := range integrations {
			_, err := store.Create(ctx, &integrations[i])
			Expect(err).To(BeNil())
		}
	})

	It("gets integrations by id", func() {
		integration, err := store.GetByID(ctx, "gi-1")
		Expect(err).To(BeNil())
		Expect(integration.Host).To(Equal("gitlab.example.com"))
	})

	It("scopes host lookups to the organisation and git_credentials type", func() {
		integration, err := store.GetByOrgAndHost(ctx, orgID, "gitlab.example.com")
		Expect(err).To(BeNil())
		Expect(integration.ID).To(Equal("gi-1"))

		_, err = store.GetByOrgAndHost(ctx, orgID, "github.com")
		Expect(err).NotTo(BeNil())
		Expect(err.Is404()).To(BeTrue(), "github_app rows must not match the credentials host lookup")
	})

	It("finds the org's GitHub App integration", func() {
		integration, err := store.GetGitHubAppForOrg(ctx, orgID)
		Expect(err).To(BeNil())
		Expect(integration.ID).To(Equal("gi-2"))

		_, err = store.GetGitHubAppForOrg(ctx, orgID2)
		Expect(err).NotTo(BeNil())
		Expect(err.Is404()).To(BeTrue())
	})

	It("lists integrations for an organisation", func() {
		integrations, err := store.ListByOrgID(ctx, orgID)
		Expect(err).To(BeNil())
		Expect(integrations).To(HaveLen(2))
	})

	It("deletes integrations and 404s on repeat", func() {
		Expect(store.Delete(ctx, "gi-1")).To(BeNil())
		err := store.Delete(ctx, "gi-1")
		Expect(err).NotTo(BeNil())
		Expect(err.Is404()).To(BeTrue())
	})
})
