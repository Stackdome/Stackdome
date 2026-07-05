package pgstore_test

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RegistryCredentialStore", func() {
	var (
		store stores.RegistryCredentialStore
		ctx   context.Context
	)

	const (
		orgID  = "org-1"
		orgID2 = "org-2"
	)

	BeforeEach(func() {
		sf := newSQLiteSessionFactory(`
			CREATE TABLE IF NOT EXISTS registry_credentials (
				id TEXT PRIMARY KEY,
				organisation_id TEXT NOT NULL,
				host TEXT NOT NULL,
				purpose TEXT NOT NULL DEFAULT 'both',
				username TEXT NOT NULL,
				encrypted_password TEXT NOT NULL,
				data_hash TEXT NOT NULL,
				created_at DATETIME,
				updated_at DATETIME,
				UNIQUE (organisation_id, host, purpose)
			)
		`)
		store = pgstore.NewRegistryCredentialStore(pgstore.RegistryCredentialStoreSpec{SessionFactory: sf})
		ctx = context.Background()

		credentials := []models.RegistryCredential{
			{
				ID:                "rc-1",
				OrganisationID:    orgID,
				Host:              "ghcr.io",
				Purpose:           models.RegistryCredentialPurposePull,
				Username:          "acme-pull",
				EncryptedPassword: "enc-1",
				DataHash:          "hash-1",
			},
			{
				ID:                "rc-2",
				OrganisationID:    orgID,
				Host:              "ghcr.io",
				Purpose:           models.RegistryCredentialPurposePush,
				Username:          "acme-push",
				EncryptedPassword: "enc-2",
				DataHash:          "hash-2",
			},
			{
				ID:                "rc-3",
				OrganisationID:    orgID2,
				Host:              "ghcr.io",
				Purpose:           models.RegistryCredentialPurposeBoth,
				Username:          "other-org",
				EncryptedPassword: "enc-3",
				DataHash:          "hash-3",
			},
		}
		for i := range credentials {
			_, err := store.Create(ctx, &credentials[i])
			Expect(err).To(BeNil())
		}
	})

	It("gets credentials by id", func() {
		credential, err := store.GetByID(ctx, "rc-1")
		Expect(err).To(BeNil())
		Expect(credential.Username).To(Equal("acme-pull"))
	})

	It("returns 404 for a missing credential", func() {
		_, err := store.GetByID(ctx, "missing")
		Expect(err).NotTo(BeNil())
		Expect(err.Is404()).To(BeTrue())
	})

	It("scopes host lookups to the organisation", func() {
		credentials, err := store.GetByOrgAndHost(ctx, orgID, "ghcr.io")
		Expect(err).To(BeNil())
		Expect(credentials).To(HaveLen(2))
		for _, credential := range credentials {
			Expect(credential.OrganisationID).To(Equal(orgID))
		}
	})

	It("lists credentials for an organisation", func() {
		credentials, err := store.ListByOrgID(ctx, orgID2)
		Expect(err).To(BeNil())
		Expect(credentials).To(HaveLen(1))
		Expect(credentials[0].ID).To(Equal("rc-3"))
	})

	It("updates credentials", func() {
		credential, err := store.GetByID(ctx, "rc-1")
		Expect(err).To(BeNil())
		credential.Username = "rotated"
		credential.EncryptedPassword = "enc-rotated"
		credential.DataHash = "hash-rotated"

		updated, err := store.Update(ctx, credential)
		Expect(err).To(BeNil())
		Expect(updated.Username).To(Equal("rotated"))
	})

	It("deletes credentials and 404s on repeat", func() {
		Expect(store.Delete(ctx, "rc-1")).To(BeNil())
		err := store.Delete(ctx, "rc-1")
		Expect(err).NotTo(BeNil())
		Expect(err.Is404()).To(BeTrue())
	})
})
