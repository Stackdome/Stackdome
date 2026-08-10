package pgstore_test

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ComputeUsageStore", func() {
	var (
		ctx      context.Context
		sf       *sqliteSessionFactory
		store    stores.ComputeUsageStore
		executor stores.AtomicExecutor
	)

	BeforeEach(func() {
		ctx = context.Background()
		sf = newSQLiteSessionFactory(
			`CREATE TABLE organisations (id TEXT PRIMARY KEY)`,
			`CREATE TABLE stacks (id TEXT PRIMARY KEY, organisation_id TEXT NOT NULL)`,
			`CREATE TABLE stack_resources (id TEXT PRIMARY KEY, stack_id TEXT NOT NULL)`,
			`CREATE TABLE volumes (id TEXT PRIMARY KEY, organisation_id TEXT NOT NULL)`,
			`CREATE TABLE postgres_addons (id TEXT PRIMARY KEY, organisation_id TEXT NOT NULL)`,
		)
		Expect(sf.New(ctx).Exec(`INSERT INTO organisations (id) VALUES ('org-1')`).Error).NotTo(HaveOccurred())
		store = pgstore.NewComputeUsageStore()
		executor = pgstore.NewAtomicExecutor(sf)
	})

	It("requires the caller transaction", func() {
		_, serr := store.LockOrganisationAndGetUsageWithTx(ctx, "org-1", "")
		Expect(serr).ToNot(BeNil())
		Expect(serr.Error()).To(ContainSubstring("transaction"))
	})

	It("locks the organisation and counts stacks and resources excluding one stack", func() {
		Expect(sf.New(ctx).Exec(`INSERT INTO stacks (id, organisation_id) VALUES ('stack-1', 'org-1'), ('stack-2', 'org-1')`).Error).NotTo(HaveOccurred())
		Expect(sf.New(ctx).Exec(`INSERT INTO stack_resources (id, stack_id) VALUES ('r1', 'stack-1'), ('r2', 'stack-1'), ('r3', 'stack-2')`).Error).NotTo(HaveOccurred())
		var usage stores.ComputeUsage
		serr := executor.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
			var getErr *errors.ServiceError
			usage, getErr = store.LockOrganisationAndGetUsageWithTx(txCtx, "org-1", "stack-1")
			return getErr
		})
		Expect(serr).To(BeNil())
		Expect(usage.StackCount).To(Equal(int64(2)))
		Expect(usage.StackResourceCount).To(Equal(int64(1)))
	})
})
