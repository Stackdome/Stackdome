package pgstore_test

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = Describe("StackStore", func() {
	var (
		store stores.StackStore
		sf    *sqliteSessionFactory
		ctx   context.Context
	)

	const (
		orgID     = "org-1"
		projectA  = "project-a"
		userID    = "user-1"
		clusterID = "cluster-1"
		stackName = "demo"
	)

	newStackSpec := func(id, name, projectID string) *models.Stack {
		return &models.Stack{
			ID:             id,
			OrganisationID: orgID,
			ProjectID:      projectID,
			ClusterID:      clusterID,
			UserID:         userID,
			Name:           name,
			NamespaceID:    "ns-id-" + id,
			Namespace:      "ns-" + id,
		}
	}

	BeforeEach(func() {
		sf = newSQLiteSessionFactory(`
			CREATE TABLE IF NOT EXISTS stacks (
				id TEXT PRIMARY KEY,
				organisation_id TEXT NOT NULL,
				project_id TEXT NOT NULL,
				cluster_id TEXT,
				user_id TEXT NOT NULL,
				name TEXT NOT NULL,
				namespace_id TEXT,
				namespace TEXT,
				labels TEXT,
				annotations TEXT,
				cr_revision TEXT,
				status TEXT,
				settings TEXT,
				created_at DATETIME,
				updated_at DATETIME,
				deletion_timestamp DATETIME
			)
		`, `
			CREATE UNIQUE INDEX IF NOT EXISTS idx_stacks_project_id_name ON stacks (project_id, name)
		`, `
			CREATE TABLE IF NOT EXISTS stack_releases (
				id TEXT PRIMARY KEY,
				stack_id TEXT NOT NULL,
				sequence INTEGER NOT NULL,
				state TEXT NOT NULL
			)
		`)
		store = pgstore.NewStackStore(&pgstore.StackStoreSpec{SessionFactory: sf})
		ctx = context.Background()

		Expect(sf.New(ctx).Exec(
			`INSERT INTO stacks (id, organisation_id, project_id, cluster_id, user_id, name, namespace_id, namespace)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			"stack-1", orgID, projectA, clusterID, userID, stackName, "ns-id-1", "ns-1",
		).Error).NotTo(HaveOccurred())
	})

	Describe("Create", func() {
		It("maps a (project_id, name) unique violation to Conflict", func() {
			created, serr := store.Create(ctx, newStackSpec("stack-2", stackName, projectA))
			Expect(created).To(BeNil())
			Expect(serr).NotTo(BeNil())
			Expect(serr.IsConflict()).To(BeTrue())
		})
	})

	Describe("CreateWithTx", func() {
		It("maps a (project_id, name) unique violation to Conflict", func() {
			tx := sf.New(ctx).Begin()
			defer tx.Rollback()
			txCtx := db.CtxWithTransaction(ctx, tx)

			created, serr := store.CreateWithTx(txCtx, newStackSpec("stack-3", stackName, projectA))
			Expect(created).To(BeNil())
			Expect(serr).NotTo(BeNil())
			Expect(serr.IsConflict()).To(BeTrue())
		})
	})

	Describe("ListWorkloadAuthorityCandidates", func() {
		It("uses one bounded query as the active stack count grows", func() {
			for i := 2; i <= 25; i++ {
				stackID := fmt.Sprintf("stack-%d", i)
				Expect(sf.New(ctx).Exec(
					`INSERT INTO stacks (id, organisation_id, project_id, cluster_id, user_id, name, namespace_id, namespace)
					 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
					stackID, orgID, projectA, clusterID, userID, stackID, "ns-id-"+stackID, "ns-"+stackID,
				).Error).NotTo(HaveOccurred())
				Expect(sf.New(ctx).Exec(
					`INSERT INTO stack_releases (id, stack_id, sequence, state) VALUES (?, ?, ?, ?)`,
					"release-"+stackID, stackID, 1, models.ReleaseStatePending,
				).Error).NotTo(HaveOccurred())
			}

			var queryCount atomic.Int32
			const callbackName = "test:count-workload-authority-queries"
			database := sf.New(ctx)
			Expect(database.Callback().Query().Before("gorm:query").Register(callbackName, func(*gorm.DB) {
				queryCount.Add(1)
			})).To(Succeed())
			DeferCleanup(func() {
				Expect(database.Callback().Query().Remove(callbackName)).To(Succeed())
			})

			stacks, serr := store.ListWorkloadAuthorityCandidates(ctx)
			Expect(serr).To(BeNil())
			Expect(stacks).To(HaveLen(24))
			Expect(queryCount.Load()).To(BeNumerically("<=", 2))
		})
	})

	Describe("GetLatestSummariesByStackIDs", func() {
		It("returns the highest sequence without loading release snapshots", func() {
			Expect(sf.New(ctx).Exec(
				`INSERT INTO stack_releases (id, stack_id, sequence, state) VALUES
				 (?, ?, ?, ?), (?, ?, ?, ?), (?, ?, ?, ?)`,
				"release-a1", "stack-1", 1, models.ReleaseStateReleased,
				"release-a2", "stack-1", 2, models.ReleaseStateFailed,
				"release-b1", "stack-2", 1, models.ReleaseStatePending,
			).Error).NotTo(HaveOccurred())
			releaseStore := pgstore.NewStackReleaseStore(pgstore.StackReleaseStoreSpec{SessionFactory: sf})

			summaries, serr := releaseStore.GetLatestSummariesByStackIDs(ctx, []string{"stack-1", "stack-2"})

			Expect(serr).To(BeNil())
			Expect(summaries).To(HaveLen(2))
			Expect(summaries["stack-1"].ID).To(Equal("release-a2"))
			Expect(summaries["stack-1"].State).To(Equal(models.ReleaseStateFailed))
			Expect(summaries["stack-1"].Snapshot).To(Equal(models.StackSnapshot{}))
			Expect(summaries["stack-2"].ID).To(Equal("release-b1"))
		})
	})
})
