package pgstore_test

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SecretStore", func() {
	var (
		store stores.SecretStore
		sf    *sqliteSessionFactory
		ctx   context.Context
	)

	const (
		orgID  = "org-1"
		orgID2 = "org-2"
		projectA  = "project-a"
		projectB  = "project-b"
		userID = "user-1"
	)

	BeforeEach(func() {
		sf = newSQLiteSessionFactory(`
			CREATE TABLE IF NOT EXISTS secrets (
				id TEXT PRIMARY KEY,
				organisation_id TEXT NOT NULL,
				project_id TEXT,
				user_id TEXT NOT NULL,
				name TEXT NOT NULL,
				description TEXT,
				type TEXT NOT NULL,
				encrypted_data TEXT NOT NULL,
				keys TEXT,
				data_hash TEXT NOT NULL,
				managed BOOLEAN NOT NULL DEFAULT false,
				managed_by_kind TEXT,
				managed_by_id TEXT,
				managed_slot TEXT,
				created_at DATETIME,
				updated_at DATETIME
			)
		`)
		store = pgstore.NewSecretStore(pgstore.SecretStoreSpec{SessionFactory: sf})
		ctx = context.Background()

		now := time.Now()
		secrets := []models.Secret{
			{
				ID:             "sec-1",
				OrganisationID: orgID,
				ProjectID:         projectA,
				UserID:         userID,
				Name:           "github-pat",
				Type:           models.SecretTypeToken,
				EncryptedData:  "enc-data-1",
				DataHash:       "hash-1",
				CreatedAt:      now.Add(-2 * time.Hour),
				UpdatedAt:      now.Add(-2 * time.Hour),
			},
			{
				ID:             "sec-2",
				OrganisationID: orgID,
				ProjectID:         projectA,
				UserID:         userID,
				Name:           "docker-creds",
				Type:           models.SecretTypeDockerRegistry,
				EncryptedData:  "enc-data-2",
				DataHash:       "hash-2",
				CreatedAt:      now.Add(-1 * time.Hour),
				UpdatedAt:      now.Add(-1 * time.Hour),
			},
			{
				ID:             "sec-3",
				OrganisationID: orgID,
				ProjectID:         projectB,
				UserID:         userID,
				Name:           "ssh-deploy-key",
				Type:           models.SecretTypeSSHKey,
				EncryptedData:  "enc-data-3",
				DataHash:       "hash-3",
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			{
				ID:             "sec-4",
				OrganisationID: orgID2,
				ProjectID:         projectA,
				UserID:         userID,
				Name:           "other-org-secret",
				Type:           models.SecretTypeGeneric,
				EncryptedData:  "enc-data-4",
				DataHash:       "hash-4",
				CreatedAt:      now,
				UpdatedAt:      now,
			},
		}
		for i := range secrets {
			Expect(sf.db.Create(&secrets[i]).Error).NotTo(HaveOccurred())
		}
	})

	Describe("ListByOrganisation", func() {
		Context("with name filter", func() {
			It("returns only the matching secret", func() {
				params := stores.ListParams{
					Filters: []stores.Filter{{Field: "name", Value: "github-pat"}},
				}
				results, err := store.ListByOrganisation(ctx, orgID, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(1))
				Expect(results[0].Name).To(Equal("github-pat"))
				Expect(results[0].ID).To(Equal("sec-1"))
			})

			It("returns empty when no match", func() {
				params := stores.ListParams{
					Filters: []stores.Filter{{Field: "name", Value: "nonexistent"}},
				}
				results, err := store.ListByOrganisation(ctx, orgID, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
			})
		})

		Context("without filter", func() {
			It("returns all secrets for the org", func() {
				results, err := store.ListByOrganisation(ctx, orgID, stores.ListParams{})
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(3))
			})

			It("does not return secrets from other orgs", func() {
				results, err := store.ListByOrganisation(ctx, orgID2, stores.ListParams{})
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(1))
				Expect(results[0].Name).To(Equal("other-org-secret"))
			})
		})

		Context("with type filter", func() {
			It("returns only secrets matching the type", func() {
				params := stores.ListParams{
					Filters: []stores.Filter{{Field: "type", Value: models.SecretTypeToken}},
				}
				results, err := store.ListByOrganisation(ctx, orgID, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(1))
				Expect(results[0].Name).To(Equal("github-pat"))
			})
		})

		Context("with multiple filters", func() {
			It("returns only secrets matching all filters", func() {
				params := stores.ListParams{
					Filters: []stores.Filter{
						{Field: "name", Value: "github-pat"},
						{Field: "type", Value: models.SecretTypeToken},
					},
				}
				results, err := store.ListByOrganisation(ctx, orgID, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(1))
				Expect(results[0].ID).To(Equal("sec-1"))
			})

			It("returns empty when filters contradict", func() {
				params := stores.ListParams{
					Filters: []stores.Filter{
						{Field: "name", Value: "github-pat"},
						{Field: "type", Value: models.SecretTypeSSHKey},
					},
				}
				results, err := store.ListByOrganisation(ctx, orgID, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
			})
		})
	})

	Describe("ListByProjectID", func() {
		Context("with name filter", func() {
			It("returns only the matching secret within the project", func() {
				params := stores.ListParams{
					Filters: []stores.Filter{{Field: "name", Value: "docker-creds"}},
				}
				results, err := store.ListByProjectID(ctx, projectA, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(1))
				Expect(results[0].Name).To(Equal("docker-creds"))
			})
		})

		Context("without filter", func() {
			It("returns all secrets for the project", func() {
				results, err := store.ListByProjectID(ctx, projectA, stores.ListParams{})
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(3))
			})

			It("returns only secrets for the specified project", func() {
				results, err := store.ListByProjectID(ctx, projectB, stores.ListParams{})
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(1))
				Expect(results[0].Name).To(Equal("ssh-deploy-key"))
			})
		})

		Context("with non-matching filter", func() {
			It("returns empty", func() {
				params := stores.ListParams{
					Filters: []stores.Filter{{Field: "name", Value: "nonexistent"}},
				}
				results, err := store.ListByProjectID(ctx, projectA, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
			})
		})
	})

	Describe("ListByProjectIDs", func() {
		Context("with name filter", func() {
			It("returns only matching secrets across projects", func() {
				params := stores.ListParams{
					Filters: []stores.Filter{{Field: "name", Value: "github-pat"}},
				}
				results, err := store.ListByProjectIDs(ctx, []string{projectA, projectB}, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(1))
				Expect(results[0].Name).To(Equal("github-pat"))
			})
		})

		Context("without filter", func() {
			It("returns all secrets across the specified projects", func() {
				results, err := store.ListByProjectIDs(ctx, []string{projectA, projectB}, stores.ListParams{})
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(4))
			})
		})

		Context("with empty project IDs", func() {
			It("returns empty result", func() {
				results, err := store.ListByProjectIDs(ctx, []string{}, stores.ListParams{})
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
			})
		})

		Context("with single project ID", func() {
			It("returns only secrets for that project", func() {
				results, err := store.ListByProjectIDs(ctx, []string{projectB}, stores.ListParams{})
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(1))
				Expect(results[0].ID).To(Equal("sec-3"))
			})
		})
	})
})
