package pgstore_test

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OrganisationStore", func() {
	const organisationTableDDL = `
		CREATE TABLE organisations (
			id TEXT PRIMARY KEY,
			name TEXT,
			platform BOOLEAN,
			created_at DATETIME,
			updated_at DATETIME
		)
	`

	It("requires a transaction for the organisation row lock", func() {
		sf := newSQLiteSessionFactory(organisationTableDDL)
		ctx := context.Background()
		organisationStore := pgstore.NewOrganisationStore(pgstore.OrganisationStoreSpec{SessionFactory: sf})

		err := organisationStore.LockByID(ctx, "org-1")

		Expect(err).To(MatchError("error: transaction not found in context"))
	})

	It("locks an existing organisation through the transaction context", func() {
		sf := newSQLiteSessionFactory(organisationTableDDL)
		ctx := context.Background()
		Expect(sf.New(ctx).Exec(`INSERT INTO organisations (id, name, platform) VALUES (?, ?, ?)`, "org-1", "Org", false).Error).
			NotTo(HaveOccurred())
		tx := sf.New(ctx).Begin()
		DeferCleanup(func() { tx.Rollback() })
		txCtx := db.CtxWithTransaction(ctx, tx)
		organisationStore := pgstore.NewOrganisationStore(pgstore.OrganisationStoreSpec{SessionFactory: sf})

		err := organisationStore.LockByID(txCtx, "org-1")

		Expect(err).To(BeNil())
	})
})
