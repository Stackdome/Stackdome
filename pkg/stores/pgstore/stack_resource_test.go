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

const stackResourcesDDL = `
	CREATE TABLE stack_resources (
		id text PRIMARY KEY,
		user_id text NOT NULL,
		stack_id text NOT NULL,
		name text NOT NULL,
		namespace text,
		labels jsonb,
		annotations jsonb,
		version bigint DEFAULT 1,
		build_config jsonb,
		image_config jsonb,
		init jsonb,
		execution_config jsonb,
		depends_on jsonb,
		lifecycle_config jsonb,
		ports jsonb,
		status jsonb,
		workload_type text NOT NULL DEFAULT 'Service',
		schedule text,
		replicas integer,
		created_at datetime,
		updated_at datetime
	)
`

var _ = Describe("StackResourceStore", func() {
	var (
		sf    *sqliteSessionFactory
		store stores.StackResourceStore
		ctx   context.Context
	)

	seedResource := func() *models.StackResource {
		seeded := &models.StackResource{
			ID:           "res-1",
			UserID:       "user-1",
			StackID:      "stack-1",
			Name:         "test",
			Namespace:    "ns-1",
			ImageConfig:  &models.ImageConfigSpec{Image: "nginx"},
			WorkloadType: models.WorkloadTypeCronJob,
			Schedule:     "*/5 * * * *",
			Status:       &models.StackResourceStatus{State: models.StackResourcePhaseReady},
		}
		Expect(sf.New(ctx).Create(seeded).Error).NotTo(HaveOccurred())
		return seeded
	}

	gitUpdateSpec := func() *models.StackResource {
		return &models.StackResource{
			ID:      "res-1",
			UserID:  "user-1",
			StackID: "stack-1",
			Name:    "test",
			BuildConfig: &models.BuildConfigSpec{
				SourceContext: models.BuildContextSource{
					Git: &models.GitBuildSource{RepoURL: "https://github.com/Stackdome/test-repo"},
				},
				DockerfilePath:          "Dockerfile",
				ContextPathWithinSource: ".",
			},
			WorkloadType: models.WorkloadTypeService,
		}
	}

	BeforeEach(func() {
		sf = newSQLiteSessionFactory(stackResourcesDDL)
		store = pgstore.NewStackResourceStore(pgstore.StackResourceStoreSpec{SessionFactory: sf})
		ctx = context.Background()
	})

	Describe("UpdateWithTx", func() {
		runUpdate := func(spec *models.StackResource) *models.StackResource {
			tx := sf.New(ctx).Begin()
			txCtx := db.CtxWithTransaction(ctx, tx)
			_, serr := store.UpdateWithTx(txCtx, "res-1", spec, &models.Stack{ID: "stack-1"})
			Expect(serr).To(BeNil())
			Expect(tx.Commit().Error).NotTo(HaveOccurred())

			fetched, gerr := store.GetByID(ctx, "res-1")
			Expect(gerr).To(BeNil())
			return fetched
		}

		It("clears fields that are zero in the update spec", func() {
			seedResource()

			fetched := runUpdate(gitUpdateSpec())

			Expect(fetched.BuildConfig).NotTo(BeNil())
			Expect(fetched.BuildConfig.SourceContext.Git).NotTo(BeNil())
			Expect(fetched.BuildConfig.SourceContext.Git.RepoURL).To(Equal("https://github.com/Stackdome/test-repo"))
			Expect(fetched.ImageConfig).To(BeNil())
			Expect(fetched.Schedule).To(BeEmpty())
			Expect(fetched.WorkloadType).To(Equal(models.WorkloadTypeService))
		})

		It("preserves status and created-only fields", func() {
			seeded := seedResource()

			fetched := runUpdate(gitUpdateSpec())

			Expect(fetched.Status).NotTo(BeNil())
			Expect(fetched.Status.State).To(Equal(models.StackResourcePhaseReady))
			Expect(fetched.Name).To(Equal(seeded.Name))
			Expect(fetched.Namespace).To(Equal(seeded.Namespace))
			Expect(fetched.CreatedAt).To(BeTemporally("~", seeded.CreatedAt, 0))
		})
	})

	Describe("Update", func() {
		It("clears fields that are zero in the update spec", func() {
			seedResource()

			_, serr := store.Update(ctx, "res-1", gitUpdateSpec(), &models.Stack{ID: "stack-1"})
			Expect(serr).To(BeNil())

			fetched, gerr := store.GetByID(ctx, "res-1")
			Expect(gerr).To(BeNil())
			Expect(fetched.BuildConfig).NotTo(BeNil())
			Expect(fetched.ImageConfig).To(BeNil())
			Expect(fetched.Schedule).To(BeEmpty())
			Expect(fetched.Status).NotTo(BeNil())
		})
	})
})
