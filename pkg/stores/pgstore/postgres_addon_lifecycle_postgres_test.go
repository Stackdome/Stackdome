package pgstore_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	serviceerrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PostgresAddonStore PostgreSQL lifecycle locking", func() {
	It("preserves concurrent hibernation and fencing changes", func() {
		dsn := os.Getenv("STACKDOME_TEST_POSTGRES_DSN")
		if dsn == "" {
			Skip("STACKDOME_TEST_POSTGRES_DSN is not set")
		}

		admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		adminSQL, err := admin.DB()
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(adminSQL.Close)
		schemaName := "postgres_lifecycle_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		Expect(admin.Exec("CREATE SCHEMA " + schemaName).Error).NotTo(HaveOccurred())
		DeferCleanup(func() {
			Expect(admin.Exec("DROP SCHEMA " + schemaName + " CASCADE").Error).NotTo(HaveOccurred())
		})

		database, err := gorm.Open(postgres.Open(dsn+" search_path="+schemaName), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		postgresDDL := strings.ReplaceAll(postgresAddonsDDL, "datetime", "timestamptz")
		Expect(database.Exec(postgresDDL).Error).NotTo(HaveOccurred())
		Expect(database.Exec(`
			CREATE TABLE postgres_addon_databases (id TEXT PRIMARY KEY, postgres_addon_id TEXT NOT NULL);
			CREATE TABLE postgres_backups (id TEXT PRIMARY KEY, postgres_addon_id TEXT NOT NULL);
		`).Error).NotTo(HaveOccurred())
		sessionFactory := &postgresTestSessionFactory{database: database}
		DeferCleanup(sessionFactory.Close)
		store := pgstore.NewPostgresAddonStore(pgstore.PostgresAddonStoreSpec{SessionFactory: sessionFactory})
		addon := &models.PostgresAddon{
			ID: "addon-1", OrganisationID: "org-1", ProjectID: "project-1", UserID: "user-1",
			ClusterID: "cluster-1", Name: "database", NamespaceID: "namespace-1", Namespace: "database",
			Status: models.PostgresAddonStatus{State: models.PostgresAddonStateReady},
		}
		ctx := context.Background()
		Expect(database.Omit(clause.Associations).Create(addon).Error).To(Succeed())

		start := make(chan struct{})
		results := make(chan *serviceerrors.ServiceError, 2)
		var requests sync.WaitGroup
		requests.Add(2)
		go func() {
			defer GinkgoRecover()
			defer requests.Done()
			<-start
			results <- store.WithTransaction(ctx, func(txCtx context.Context) *serviceerrors.ServiceError {
				_, serr := store.SetHibernationWithTx(txCtx, addon.ID, true)
				return serr
			})
		}()
		go func() {
			defer GinkgoRecover()
			defer requests.Done()
			<-start
			results <- store.WithTransaction(ctx, func(txCtx context.Context) *serviceerrors.ServiceError {
				_, serr := store.SetFencingWithTx(txCtx, addon.ID, true)
				return serr
			})
		}()
		close(start)
		requests.Wait()
		close(results)
		for serr := range results {
			Expect(serr).To(BeNil(), fmt.Sprintf("concurrent lifecycle update failed: %v", serr))
		}

		restartedStore := pgstore.NewPostgresAddonStore(pgstore.PostgresAddonStoreSpec{SessionFactory: sessionFactory})
		persisted, serr := restartedStore.GetByID(ctx, addon.ID)
		Expect(serr).To(BeNil())
		Expect(persisted.LifecycleConfig.HibernationEnabled).To(BeTrue())
		Expect(persisted.LifecycleConfig.FencingEnabled).To(BeTrue())
		Expect(persisted.Status.State).To(Equal(models.PostgresAddonStatePending))
	})

	It("serializes ordinary configuration with lifecycle changes and leaves restart work pending", func() {
		dsn := os.Getenv("STACKDOME_TEST_POSTGRES_DSN")
		if dsn == "" {
			Skip("STACKDOME_TEST_POSTGRES_DSN is not set")
		}

		admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		adminSQL, err := admin.DB()
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(adminSQL.Close)
		schemaName := "postgres_config_lifecycle_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		Expect(admin.Exec("CREATE SCHEMA " + schemaName).Error).NotTo(HaveOccurred())
		DeferCleanup(func() {
			Expect(admin.Exec("DROP SCHEMA " + schemaName + " CASCADE").Error).NotTo(HaveOccurred())
		})

		database, err := gorm.Open(postgres.Open(dsn+" search_path="+schemaName), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		postgresDDL := strings.ReplaceAll(postgresAddonsDDL, "datetime", "timestamptz")
		Expect(database.Exec(postgresDDL).Error).NotTo(HaveOccurred())
		Expect(database.Exec(`
			CREATE TABLE postgres_addon_databases (id TEXT PRIMARY KEY, postgres_addon_id TEXT NOT NULL);
			CREATE TABLE postgres_backups (id TEXT PRIMARY KEY, postgres_addon_id TEXT NOT NULL);
		`).Error).NotTo(HaveOccurred())
		sessionFactory := &postgresTestSessionFactory{database: database}
		DeferCleanup(sessionFactory.Close)
		store := pgstore.NewPostgresAddonStore(pgstore.PostgresAddonStoreSpec{SessionFactory: sessionFactory})
		addon := &models.PostgresAddon{
			ID: "addon-1", OrganisationID: "org-1", ProjectID: "project-1", UserID: "user-1",
			ClusterID: "cluster-1", Name: "database", NamespaceID: "namespace-1", Namespace: "database",
			Instances: models.PostgresInstances{Count: 1},
			Status:    models.PostgresAddonStatus{State: models.PostgresAddonStateReady},
		}
		ctx := context.Background()
		Expect(database.Omit(clause.Associations).Create(addon).Error).To(Succeed())

		lifecycleLocked := make(chan struct{})
		releaseLifecycle := make(chan struct{})
		lifecycleDone := make(chan *serviceerrors.ServiceError, 1)
		go func() {
			defer GinkgoRecover()
			lifecycleDone <- store.WithTransaction(ctx, func(txCtx context.Context) *serviceerrors.ServiceError {
				if _, serr := store.SetHibernationWithTx(txCtx, addon.ID, true); serr != nil {
					return serr
				}
				close(lifecycleLocked)
				<-releaseLifecycle
				return nil
			})
		}()
		Eventually(lifecycleLocked).Should(BeClosed())

		desired := *addon
		desired.Instances.Count = 2
		type updateResult struct {
			addon *models.PostgresAddon
			err   *serviceerrors.ServiceError
		}
		updateDone := make(chan updateResult, 1)
		go func() {
			defer GinkgoRecover()
			var updated *models.PostgresAddon
			serr := store.WithTransaction(ctx, func(txCtx context.Context) *serviceerrors.ServiceError {
				var updateErr *serviceerrors.ServiceError
				updated, updateErr = store.UpdateWithTx(txCtx, &desired)
				return updateErr
			})
			updateDone <- updateResult{addon: updated, err: serr}
		}()

		Consistently(updateDone, 100*time.Millisecond).ShouldNot(Receive())
		close(releaseLifecycle)
		Expect(<-lifecycleDone).To(BeNil())
		result := <-updateDone
		Expect(result.err).To(BeNil())
		Expect(result.addon.LifecycleConfig.HibernationEnabled).To(BeTrue())
		Expect(result.addon.Instances.Count).To(Equal(2))
		Expect(result.addon.Status.State).To(Equal(models.PostgresAddonStatePending))

		restartedStore := pgstore.NewPostgresAddonStore(pgstore.PostgresAddonStoreSpec{SessionFactory: sessionFactory})
		pending, serr := restartedStore.InternalListIDs(ctx, "status->>'state' = ?", string(models.PostgresAddonStatePending))
		Expect(serr).To(BeNil())
		Expect(pending).To(HaveLen(1))
		Expect(pending[0]).To(Equal(addon.ID))
	})
})
