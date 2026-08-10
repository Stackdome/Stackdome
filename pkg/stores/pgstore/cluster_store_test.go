package pgstore_test

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/db"
	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ClusterStore", func() {
	var (
		store stores.ClusterStore
		sf    *sqliteSessionFactory
		ctx   context.Context
	)

	newCluster := func(id, orgID, clusterURL string, shared bool) *models.Cluster {
		return &models.Cluster{
			ID:                     id,
			OrganisationID:         orgID,
			Name:                   models.SharedComputeClusterName,
			SharedCompute:          shared,
			ClusterURL:             clusterURL,
			EncryptedClusterCAData: "encrypted-ca",
			EncryptedToken:         "encrypted-token",
		}
	}

	BeforeEach(func() {
		sf = newSQLiteSessionFactory(`
			CREATE TABLE clusters (
				id TEXT PRIMARY KEY,
				organisation_id TEXT NOT NULL,
				name TEXT NOT NULL,
				created_at DATETIME,
				updated_at DATETIME,
				deletion_timestamp DATETIME,
				shared_compute BOOLEAN NOT NULL DEFAULT FALSE,
				cluster_url TEXT NOT NULL,
				encrypted_cluster_ca_data TEXT NOT NULL,
				encrypted_token TEXT NOT NULL,
				manager_running BOOLEAN NOT NULL DEFAULT FALSE,
				cluster_info TEXT
			)
		`, `
			CREATE UNIQUE INDEX idx_clusters_cluster_url_unique ON clusters (cluster_url)
		`, `
			CREATE TABLE cluster_image_registries (
				id TEXT PRIMARY KEY,
				cluster_id TEXT,
				status TEXT
			)
		`)
		store = pgstore.NewClusterStore(pgstore.ClusterStoreSpec{SessionFactory: sf})
		ctx = context.Background()
	})

	It("maps a global cluster URL violation to Conflict", func() {
		first, err := store.Create(ctx, newCluster("cluster-1", "org-1", "https://cluster.example.com", false))
		Expect(err).To(BeNil())
		Expect(first).NotTo(BeNil())

		second, err := store.Create(ctx, newCluster("cluster-2", "org-2", "https://cluster.example.com", false))
		Expect(second).To(BeNil())
		Expect(err).NotTo(BeNil())
		Expect(err.IsConflict()).To(BeTrue())
	})

	It("finds only a matching cluster ID by shared-compute ownership", func() {
		_, err := store.Create(ctx, newCluster("tenant-cluster", "org-1", "https://tenant.example.com", false))
		Expect(err).To(BeNil())
		_, err = store.Create(ctx, newCluster("shared-cluster", "org-2", "https://shared.example.com", true))
		Expect(err).To(BeNil())

		sharedID, err := store.FindAnyClusterIDBySharedCompute(ctx, true)
		Expect(err).To(BeNil())
		Expect(sharedID).To(Equal("shared-cluster"))

		tenantID, err := store.FindAnyClusterIDBySharedCompute(ctx, false)
		Expect(err).To(BeNil())
		Expect(tenantID).To(Equal("tenant-cluster"))
	})

	It("returns an empty ID when no cluster has the requested ownership", func() {
		clusterID, err := store.FindAnyClusterIDBySharedCompute(ctx, true)
		Expect(err).To(BeNil())
		Expect(clusterID).To(BeEmpty())
	})

	It("queries only clusters with pending registry work or cluster deletion intent", func() {
		pending := newCluster("cluster-pending", "org-1", "https://pending.example.com", false)
		running := newCluster("cluster-running", "org-2", "https://running.example.com", false)
		deleting := newCluster("cluster-deleting", "org-3", "https://deleting.example.com", false)
		deletedAt := time.Now().UTC()
		deleting.DeletionTimestamp = &deletedAt

		_, err := store.Create(ctx, pending)
		Expect(err).To(BeNil())
		_, err = store.Create(ctx, running)
		Expect(err).To(BeNil())
		_, err = store.Create(ctx, deleting)
		Expect(err).To(BeNil())
		Expect(sf.New(ctx).Exec(
			`INSERT INTO cluster_image_registries (id, cluster_id, status) VALUES (?, ?, ?), (?, ?, ?)`,
			"registry-pending", pending.ID, `{"state":"Pending"}`,
			"registry-running", running.ID, `{"state":"Running"}`,
		).Error).NotTo(HaveOccurred())

		clusterIDs, listErr := store.ListIDsForImageRegistryReconciliation(ctx)
		Expect(listErr).To(BeNil())
		Expect(clusterIDs).To(ConsistOf(pending.ID, deleting.ID))
	})

	It("maps a transactional cluster URL violation to Conflict", func() {
		_, err := store.Create(ctx, newCluster("cluster-1", "org-1", "https://cluster.example.com", false))
		Expect(err).To(BeNil())
		tx := sf.New(ctx).Begin()
		DeferCleanup(func() { tx.Rollback() })

		created, err := store.CreateWithTx(
			db.CtxWithTransaction(ctx, tx),
			newCluster("cluster-2", "org-2", "https://cluster.example.com", false),
		)
		Expect(created).To(BeNil())
		Expect(err).NotTo(BeNil())
		Expect(err.IsConflict()).To(BeTrue())
	})

	It("updates only the requested organisation-owned shared cluster", func() {
		_, err := store.Create(ctx, newCluster("cluster-1", "org-1", "https://one.example.com", true))
		Expect(err).To(BeNil())
		_, err = store.Create(ctx, newCluster("cluster-2", "org-2", "https://two.example.com", true))
		Expect(err).To(BeNil())

		clusters, err := store.ListSharedComputeClustersForOrg(ctx, "org-1")
		Expect(err).To(BeNil())
		Expect(clusters).To(HaveLen(1))
		Expect(clusters[0].ID).To(Equal("cluster-1"))

		updated := newCluster("cluster-1", "org-1", "https://three.example.com", true)
		Expect(store.UpdateSharedComputeCluster(ctx, updated)).To(BeNil())
		stored, err := store.Get(ctx, "cluster-1")
		Expect(err).To(BeNil())
		Expect(stored.ClusterURL).To(Equal(updated.ClusterURL))
	})

	It("maps a conflicting shared cluster URL update to Conflict", func() {
		_, err := store.Create(ctx, newCluster("cluster-1", "org-1", "https://one.example.com", true))
		Expect(err).To(BeNil())
		_, err = store.Create(ctx, newCluster("cluster-2", "org-2", "https://two.example.com", true))
		Expect(err).To(BeNil())

		updated := newCluster("cluster-1", "org-1", "https://two.example.com", true)
		err = store.UpdateSharedComputeCluster(ctx, updated)
		Expect(err).NotTo(BeNil())
		Expect(err.IsConflict()).To(BeTrue())
	})

	It("requires and uses the active transaction for deletion", func() {
		_, err := store.Create(ctx, newCluster("cluster-1", "org-1", "https://one.example.com", true))
		Expect(err).To(BeNil())
		Expect(store.DeleteWithTx(ctx, "cluster-1")).To(MatchError(ContainSubstring("transaction not found")))

		tx := sf.New(ctx).Begin()
		txCtx := db.CtxWithTransaction(ctx, tx)
		Expect(store.DeleteWithTx(txCtx, "cluster-1")).To(BeNil())
		Expect(tx.Commit().Error).NotTo(HaveOccurred())
		_, err = store.Get(ctx, "cluster-1")
		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(Equal(apperrors.ErrorNotFound))
	})
})
