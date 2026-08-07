package pgstore_test

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StackDomainsStore", func() {
	It("returns a generic conflict for a transactional FQDN collision", func() {
		sf := newSQLiteSessionFactory(`
			CREATE TABLE IF NOT EXISTS stack_domains (
				id TEXT PRIMARY KEY,
				organisation_id TEXT NOT NULL,
				fqdn TEXT NOT NULL UNIQUE,
				stack_id TEXT NOT NULL,
				stack_resource_id TEXT NOT NULL,
				stack_resource_name TEXT NOT NULL,
				target_port INTEGER NOT NULL,
				created_at DATETIME,
				updated_at DATETIME
			)
		`)
		store := pgstore.NewStackDomainsStore(pgstore.StackDomainsStoreSpec{SessionFactory: sf})
		ctx := context.Background()
		tx := sf.New(ctx).Begin()
		DeferCleanup(func() { _ = tx.Rollback().Error })
		txCtx := db.CtxWithTransaction(ctx, tx)

		first := &models.StackDomain{
			ID: "domain-1", OrganisationID: "org-1", Fqdn: "api.customer.example",
			StackID: "stack-1", StackResourceID: "resource-1", StackResourceName: "api", TargetPort: 8080,
		}
		created, createErr := store.CreateWithTx(txCtx, first)
		Expect(createErr).To(BeNil())
		Expect(created.Fqdn).To(Equal(first.Fqdn))

		_, collisionErr := store.CreateWithTx(txCtx, &models.StackDomain{
			ID: "domain-2", OrganisationID: "org-2", Fqdn: first.Fqdn,
			StackID: "stack-2", StackResourceID: "resource-2", StackResourceName: "other", TargetPort: 9090,
		})

		Expect(collisionErr).NotTo(BeNil())
		Expect(collisionErr.Code).To(Equal(errors.ErrorConflict))
		Expect(collisionErr.Reason).To(Equal("domain 'api.customer.example' is already in use"))
	})
})
