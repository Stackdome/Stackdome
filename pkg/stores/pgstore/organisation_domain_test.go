package pgstore_test

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OrganisationDomainStore", func() {
	var (
		store stores.OrganisationDomainStore
		sf    *sqliteSessionFactory
		ctx   context.Context
	)

	const (
		orgID      = "org-1"
		otherOrgID = "org-2"
		takenDoma  = "acme.stackdome.io"
	)

	BeforeEach(func() {
		sf = newSQLiteSessionFactory(`
			CREATE TABLE IF NOT EXISTS organisation_domains (
				id TEXT PRIMARY KEY,
				domain TEXT NOT NULL UNIQUE,
				organisation_id TEXT NOT NULL,
				created_at DATETIME,
				updated_at DATETIME
			)
		`)
		store = pgstore.NewOrganisationDomainStore(pgstore.OrganisationDomainStoreSpec{SessionFactory: sf})
		ctx = context.Background()

		_, serr := store.Create(ctx, &models.OrganisationDomain{
			ID:             "od-1",
			Domain:         takenDoma,
			OrganisationID: orgID,
		})
		Expect(serr).To(BeNil())
	})

	Describe("Create", func() {
		It("returns a conflict when the domain is already taken by another organisation", func() {
			_, serr := store.Create(ctx, &models.OrganisationDomain{
				ID:             "od-2",
				Domain:         takenDoma,
				OrganisationID: otherOrgID,
			})
			Expect(serr).NotTo(BeNil())
			Expect(serr.Code).To(Equal(errors.ErrorConflict))
		})

		It("creates a domain that is not taken", func() {
			created, serr := store.Create(ctx, &models.OrganisationDomain{
				ID:             "od-3",
				Domain:         "acme-x7f2.stackdome.io",
				OrganisationID: otherOrgID,
			})
			Expect(serr).To(BeNil())
			Expect(created.Domain).To(Equal("acme-x7f2.stackdome.io"))
		})
	})

	Describe("CreateWithTx", func() {
		It("returns a conflict when the domain is already taken", func() {
			txCtx := db.CtxWithTransaction(ctx, sf.New(ctx))
			_, serr := store.CreateWithTx(txCtx, &models.OrganisationDomain{
				ID:             "od-4",
				Domain:         takenDoma,
				OrganisationID: otherOrgID,
			})
			Expect(serr).NotTo(BeNil())
			Expect(serr.Code).To(Equal(errors.ErrorConflict))
		})
	})
})
