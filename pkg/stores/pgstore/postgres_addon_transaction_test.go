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
	It("rolls back UpdateWithTx when the surrounding admission transaction fails", func() {
		ctx := context.Background()
		sf := newSQLiteSessionFactory(postgresAddonsDDL)
		store := pgstore.NewPostgresAddonStore(pgstore.PostgresAddonStoreSpec{SessionFactory: sf})
		addon := &models.PostgresAddon{ID: "addon-1", OrganisationID: "org-1", ProjectID: "project-1", UserID: "user-1", ClusterID: "cluster-1", Name: "database", NamespaceID: "namespace-1", Namespace: "database"}
		Expect(sf.New(ctx).Omit(clause.Associations).Create(addon).Error).To(Succeed())

		serr := store.WithTransaction(ctx, func(txCtx context.Context) *serviceerrors.ServiceError {
			addon.LifecycleConfig.HibernationEnabled = true
			if _, updateErr := store.UpdateWithTx(txCtx, addon); updateErr != nil {
				return updateErr
			}
			return serviceerrors.GeneralError("force rollback")
		})

		Expect(serr.Reason).To(Equal("force rollback"))
		var persisted models.PostgresAddon
		Expect(sf.New(ctx).First(&persisted, "id = ?", addon.ID).Error).To(Succeed())
		Expect(persisted.LifecycleConfig.HibernationEnabled).To(BeFalse())
	})
})
