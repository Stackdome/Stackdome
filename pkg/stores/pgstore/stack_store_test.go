package pgstore_test

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
})
