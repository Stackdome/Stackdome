package postgresaddon

import (
	"context"
	"database/sql"
	stderrors "errors"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/Stackdome/stackdome/pkg/worker"
	"github.com/glebarez/sqlite"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type lifecycleSQLiteSessionFactory struct{ db *gorm.DB }

func newLifecycleSQLiteSessionFactory() *lifecycleSQLiteSessionFactory {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	Expect(err).NotTo(HaveOccurred())
	for _, ddl := range []string{lifecyclePostgresAddonsDDL, lifecyclePostgresDatabasesDDL, lifecyclePostgresBackupsDDL} {
		Expect(database.Exec(ddl).Error).To(Succeed())
	}
	return &lifecycleSQLiteSessionFactory{db: database}
}

const lifecyclePostgresAddonsDDL = `
	CREATE TABLE postgres_addons (
		id text PRIMARY KEY, organisation_id text NOT NULL, project_id text, user_id text NOT NULL,
		cluster_id text NOT NULL, name text NOT NULL, namespace_id text NOT NULL, namespace text NOT NULL,
		labels jsonb, annotations jsonb, revision text, postgres_version jsonb, instances jsonb,
		resources jsonb, storage jsonb, configuration jsonb, initialization jsonb, backup_config jsonb,
		backup_requested_at datetime, lifecycle_config jsonb, deletion_timestamp datetime, status jsonb,
		created_at datetime, updated_at datetime
	)`

const lifecyclePostgresDatabasesDDL = `
	CREATE TABLE postgres_addon_databases (
		id text PRIMARY KEY, postgres_addon_id text NOT NULL, name text NOT NULL,
		extensions jsonb, created_at datetime, updated_at datetime
	)`

const lifecyclePostgresBackupsDDL = `
	CREATE TABLE postgres_backups (
		id text PRIMARY KEY, postgres_addon_id text NOT NULL, name text, description text, type text,
		phase text, started_at datetime, completed_at datetime, error text, size_bytes integer,
		created_at datetime, updated_at datetime
	)`

func (f *lifecycleSQLiteSessionFactory) Init(*config.DatabaseConfig) {}
func (f *lifecycleSQLiteSessionFactory) DirectDB() *sql.DB {
	direct, _ := f.db.DB()
	return direct
}
func (f *lifecycleSQLiteSessionFactory) New(ctx context.Context) *gorm.DB {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	return f.db.Session(&gorm.Session{})
}
func (f *lifecycleSQLiteSessionFactory) CheckConnection() error { return nil }
func (f *lifecycleSQLiteSessionFactory) Close() error {
	direct, _ := f.db.DB()
	return direct.Close()
}

var _ = Describe("Postgres lifecycle restart recovery", func() {
	It("discovers and applies a committed lifecycle change when enqueue is lost", func() {
		ctx := context.Background()
		ctrl := gomock.NewController(GinkgoT())
		sessionFactory := newLifecycleSQLiteSessionFactory()
		permissions := mocks.NewMockPermissionService(ctrl)
		enqueuer := mocks.NewMockBackgroundJobEnqueuer(ctrl)
		service := services.NewPostgresAddonService(services.PostgresAddonServiceSpec{
			SessionFactory: sessionFactory,
			Logger:         logger.NewLogger(),
			Permissions:    permissions,
			RuntimePolicy:  &activeAddonRuntimePolicy{},
		})
		service.InjectBackgroundJobEnqueuer(services.BackgroundJobEnqueuerDep{BackgroundJobEnqueuer: enqueuer})
		addon := &models.PostgresAddon{
			ID: "addon-1", OrganisationID: "org-1", ProjectID: "project-1", UserID: "user-1",
			ClusterID: "cluster-1", Name: "database", NamespaceID: "namespace-1", Namespace: "database",
			Status: models.PostgresAddonStatus{State: models.PostgresAddonStateReady},
		}
		Expect(sessionFactory.New(ctx).Omit(clause.Associations).Create(addon).Error).To(Succeed())
		permissions.EXPECT().Check(ctx, "project-1", auth.ResourceAddonsPostgres, "addon-1", auth.ActionWrite).Return(nil)
		enqueuer.EXPECT().Enqueue(models.PostgresAddonOperand{ID: "addon-1"}).Return(stderrors.New("process stopped before enqueue completed"))

		serr := service.TriggerHibernate(ctx, "addon-1", true)

		Expect(serr).NotTo(BeNil())
		persisted, getErr := service.InternalGetPostgresAddon(ctx, "addon-1")
		Expect(getErr).To(BeNil())
		Expect(persisted.LifecycleConfig.HibernationEnabled).To(BeTrue())
		Expect(persisted.Status.State).To(Equal(models.PostgresAddonStatePending))

		reconciler := &recordingLifecycleReconciler{}
		references := NewMockreferenceService(ctrl)
		releases := NewMockreleaseService(ctrl)
		stacks := NewMockstackService(ctrl)
		releaseID := "release-1"
		release := &models.StackRelease{
			ID: releaseID, StackID: "stack-1", State: models.ReleaseStateReleased,
			Snapshot: models.StackSnapshot{
				Stack:       models.StackShellSnapshot{ID: "stack-1", OrganisationID: "org-1"},
				Connections: models.StackConnections{{From: models.TopologyNodeRef{Type: models.TopologyNodeTypePostgresAddon, Id: "addon-1"}}},
			},
		}
		references.EXPECT().IsReferentInUse(ctx, models.ReferentPostgresAddon, "addon-1").Return(true, []models.ResourceReference{{StackID: "stack-1", ReleaseID: &releaseID}}, nil)
		releases.EXPECT().InternalGet(ctx, releaseID).Return(release, nil).Times(2)
		stacks.EXPECT().InternalGetStack(ctx, "stack-1").Return(&models.Stack{
			ID: "stack-1", OrganisationID: "org-1", Status: &models.StackStatus{LastConverged: &models.StackConvergenceRecord{ReleaseID: releaseID}},
		}, nil).Times(2)
		restartedWorker := &postgresAddonWorker{
			postgresAddonService: service,
			runtimePolicy:        &activeAddonRuntimePolicy{},
			referenceService:     references,
			releaseService:       releases,
			stackService:         stacks,
			subReconcilers:       []subReconciler{reconciler},
			BaseWorker:           worker.NewBaseWorker(WorkerName, "test"),
		}
		operands, inputErr := restartedWorker.GetInput(ctx)
		Expect(inputErr).To(BeNil())
		Expect(operands).To(ConsistOf(models.PostgresAddonOperand{ID: "addon-1"}))

		_, executeErr := restartedWorker.Execute(ctx, operands[0])

		Expect(executeErr).To(BeNil())
		Expect(reconciler.hibernationEnabled).To(BeTrue())
	})
})

type recordingLifecycleReconciler struct{ hibernationEnabled bool }

func (*recordingLifecycleReconciler) Name() string { return "record-lifecycle" }
func (r *recordingLifecycleReconciler) Reconcile(_ context.Context, addon *models.PostgresAddon) (subReconcilerResult, error) {
	r.hibernationEnabled = addon.LifecycleConfig.HibernationEnabled
	return resultNil, nil
}
