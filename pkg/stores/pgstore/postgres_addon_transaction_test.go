package pgstore_test

import (
	"context"

	serviceerrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm/clause"
)

const postgresAddonsDDL = `
	CREATE TABLE postgres_addons (
		id text PRIMARY KEY,
		organisation_id text NOT NULL,
		project_id text,
		user_id text NOT NULL,
		cluster_id text NOT NULL,
		name text NOT NULL,
		namespace_id text NOT NULL,
		namespace text NOT NULL,
		labels jsonb,
		annotations jsonb,
		revision text,
		postgres_version jsonb,
		instances jsonb,
		resources jsonb,
		storage jsonb,
		configuration jsonb,
		initialization jsonb,
		backup_config jsonb,
		backup_requested_at datetime,
		lifecycle_config jsonb,
		deletion_timestamp datetime,
		status jsonb,
		created_at datetime,
		updated_at datetime
	)
`

var _ = Describe("PostgresAddonStore transaction boundary", func() {
	It("preserves lifecycle state during an ordinary configuration update", func() {
		ctx := context.Background()
		sf := newSQLiteSessionFactory(postgresAddonsDDL)
		store := pgstore.NewPostgresAddonStore(pgstore.PostgresAddonStoreSpec{SessionFactory: sf})
		addon := &models.PostgresAddon{ID: "addon-1", OrganisationID: "org-1", ProjectID: "project-1", UserID: "user-1", ClusterID: "cluster-1", Name: "database", NamespaceID: "namespace-1", Namespace: "database", LifecycleConfig: models.PostgresLifecycleConfig{HibernationEnabled: true}}
		Expect(sf.New(ctx).Omit(clause.Associations).Create(addon).Error).To(Succeed())

		var updated *models.PostgresAddon
		serr := store.WithTransaction(ctx, func(txCtx context.Context) *serviceerrors.ServiceError {
			request := *addon
			request.LifecycleConfig = models.PostgresLifecycleConfig{}
			request.Instances = models.PostgresInstances{Count: 3}
			var updateErr *serviceerrors.ServiceError
			updated, updateErr = store.UpdateWithTx(txCtx, &request)
			return updateErr
		})

		Expect(serr).To(BeNil())
		var persisted models.PostgresAddon
		Expect(sf.New(ctx).First(&persisted, "id = ?", addon.ID).Error).To(Succeed())
		Expect(persisted.Instances.Count).To(Equal(3))
		Expect(persisted.LifecycleConfig.HibernationEnabled).To(BeTrue())
		Expect(updated.LifecycleConfig.HibernationEnabled).To(BeTrue())
	})

	It("changes one lifecycle flag without erasing the other and marks reconciliation pending", func() {
		ctx := context.Background()
		sf := newSQLiteSessionFactory(postgresAddonsDDL)
		store := pgstore.NewPostgresAddonStore(pgstore.PostgresAddonStoreSpec{SessionFactory: sf})
		addon := &models.PostgresAddon{
			ID: "addon-1", OrganisationID: "org-1", ProjectID: "project-1", UserID: "user-1", ClusterID: "cluster-1", Name: "database", NamespaceID: "namespace-1", Namespace: "database",
			LifecycleConfig: models.PostgresLifecycleConfig{FencingEnabled: true},
			Status:          models.PostgresAddonStatus{State: models.PostgresAddonStateReady},
		}
		Expect(sf.New(ctx).Omit(clause.Associations).Create(addon).Error).To(Succeed())

		serr := store.WithTransaction(ctx, func(txCtx context.Context) *serviceerrors.ServiceError {
			_, updateErr := store.SetHibernationWithTx(txCtx, addon.ID, true)
			return updateErr
		})

		Expect(serr).To(BeNil())
		var persisted models.PostgresAddon
		Expect(sf.New(ctx).First(&persisted, "id = ?", addon.ID).Error).To(Succeed())
		Expect(persisted.LifecycleConfig.HibernationEnabled).To(BeTrue())
		Expect(persisted.LifecycleConfig.FencingEnabled).To(BeTrue())
		Expect(persisted.Status.State).To(Equal(models.PostgresAddonStatePending))
	})
})
