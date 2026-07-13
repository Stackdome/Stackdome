package pgstore_test

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PreviewStackStore", func() {
	var (
		store stores.PreviewStackStore
		sf    *sqliteSessionFactory
		ctx   context.Context
	)

	const (
		orgID      = "org-1"
		projectID  = "project-1"
		projectID2 = "project-2"
		userID     = "user-1"
		configA    = "config-a"
		configB    = "config-b"
	)

	BeforeEach(func() {
		sf = newSQLiteSessionFactory(`
			CREATE TABLE IF NOT EXISTS preview_stacks (
				id TEXT PRIMARY KEY,
				organisation_id TEXT NOT NULL,
				project_id TEXT NOT NULL,
				user_id TEXT NOT NULL,
				stack_preview_config_id TEXT NOT NULL,
				stack_id TEXT,
				active_release_id TEXT,
				name TEXT NOT NULL,
				pr_number TEXT NOT NULL,
				branch TEXT NOT NULL,
				commit_sha TEXT NOT NULL,
				source TEXT NOT NULL,
				image_overrides TEXT,
				labels TEXT,
				annotations TEXT,
				status TEXT NOT NULL,
				reconciler_status TEXT NOT NULL DEFAULT '{}',
				force_sync_requested_at DATETIME,
				stackfile_content TEXT,
				deletion_timestamp DATETIME,
				created_at DATETIME,
				updated_at DATETIME
			)
		`)
		store = pgstore.NewPreviewStackStore(pgstore.PreviewStackStoreSpec{SessionFactory: sf})
		ctx = context.Background()

		now := time.Now()
		previews := []models.PreviewStack{
			{
				ID:                   "ps-1",
				OrganisationID:       orgID,
				ProjectID:            projectID,
				UserID:               userID,
				StackPreviewConfigID: configA,
				Name:                 "preview-pr-10",
				PRNumber:             "10",
				Branch:               "feature/auth",
				CommitSHA:            "aaa1111",
				Source:               models.PreviewStackSourceWebhook,
				CreatedAt:            now.Add(-2 * time.Hour),
				UpdatedAt:            now.Add(-2 * time.Hour),
			},
			{
				ID:                   "ps-2",
				OrganisationID:       orgID,
				ProjectID:            projectID,
				UserID:               userID,
				StackPreviewConfigID: configB,
				Name:                 "preview-pr-20",
				PRNumber:             "20",
				Branch:               "feature/dashboard",
				CommitSHA:            "bbb2222",
				Source:               models.PreviewStackSourceManual,
				CreatedAt:            now.Add(-1 * time.Hour),
				UpdatedAt:            now.Add(-1 * time.Hour),
			},
			{
				ID:                   "ps-3",
				OrganisationID:       orgID,
				ProjectID:            projectID,
				UserID:               userID,
				StackPreviewConfigID: configA,
				Name:                 "preview-pr-30",
				PRNumber:             "30",
				Branch:               "fix/login",
				CommitSHA:            "ccc3333",
				Source:               models.PreviewStackSourceWebhook,
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				ID:                   "ps-4",
				OrganisationID:       orgID,
				ProjectID:            projectID2,
				UserID:               userID,
				StackPreviewConfigID: configA,
				Name:                 "preview-pr-40",
				PRNumber:             "40",
				Branch:               "feature/api",
				CommitSHA:            "ddd4444",
				Source:               models.PreviewStackSourceWebhook,
				CreatedAt:            now,
				UpdatedAt:            now,
			},
		}
		for i := range previews {
			Expect(sf.db.Create(&previews[i]).Error).NotTo(HaveOccurred())
		}
	})

	Describe("ListByProjectID", func() {
		Context("with stack_preview_config_id filter", func() {
			It("returns only previews matching the config", func() {
				params := stores.ListParams{
					Filters: []stores.Filter{{Field: "stack_preview_config_id", Value: configA}},
				}
				result, err := store.ListByProjectID(ctx, projectID, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Items).To(HaveLen(2))
				for _, item := range result.Items {
					Expect(item.StackPreviewConfigID).To(Equal(configA))
					Expect(item.ProjectID).To(Equal(projectID))
				}
			})

			It("returns single result when only one matches", func() {
				params := stores.ListParams{
					Filters: []stores.Filter{{Field: "stack_preview_config_id", Value: configB}},
				}
				result, err := store.ListByProjectID(ctx, projectID, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Items).To(HaveLen(1))
				Expect(result.Items[0].StackPreviewConfigID).To(Equal(configB))
				Expect(result.Items[0].Name).To(Equal("preview-pr-20"))
			})
		})

		Context("without filter", func() {
			It("returns all previews for the project", func() {
				result, err := store.ListByProjectID(ctx, projectID, stores.ListParams{})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Items).To(HaveLen(3))
				Expect(result.Total).To(Equal(int64(3)))
			})

			It("does not return previews from other projects", func() {
				result, err := store.ListByProjectID(ctx, projectID2, stores.ListParams{})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Items).To(HaveLen(1))
				Expect(result.Items[0].ID).To(Equal("ps-4"))
			})
		})

		Context("with non-existent config_id", func() {
			It("returns empty result", func() {
				params := stores.ListParams{
					Filters: []stores.Filter{{Field: "stack_preview_config_id", Value: "config-nonexistent"}},
				}
				result, err := store.ListByProjectID(ctx, projectID, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Items).To(BeEmpty())
				Expect(result.Total).To(Equal(int64(0)))
			})
		})

		Context("with name filter", func() {
			It("returns only the preview with the matching name", func() {
				params := stores.ListParams{
					Filters: []stores.Filter{{Field: "name", Value: "preview-pr-10"}},
				}
				result, err := store.ListByProjectID(ctx, projectID, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Items).To(HaveLen(1))
				Expect(result.Items[0].ID).To(Equal("ps-1"))
			})
		})

		Context("with multiple filters", func() {
			It("returns only previews matching all filters", func() {
				params := stores.ListParams{
					Filters: []stores.Filter{
						{Field: "stack_preview_config_id", Value: configA},
						{Field: "name", Value: "preview-pr-10"},
					},
				}
				result, err := store.ListByProjectID(ctx, projectID, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Items).To(HaveLen(1))
				Expect(result.Items[0].ID).To(Equal("ps-1"))
			})
		})

		Context("pagination metadata", func() {
			It("populates total and page info", func() {
				result, err := store.ListByProjectID(ctx, projectID, stores.ListParams{Page: 1, PageSize: 2})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Items).To(HaveLen(2))
				Expect(result.Total).To(Equal(int64(3)))
				Expect(result.Page).To(Equal(1))
				Expect(result.PageSize).To(Equal(2))
				Expect(result.TotalPages).To(Equal(2))
			})
		})
	})

	Describe("ListByConfigID", func() {
		Context("without filter", func() {
			It("returns all previews for the config across projects", func() {
				result, err := store.ListByConfigID(ctx, configA, stores.ListParams{})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Items).To(HaveLen(3))
			})
		})

		Context("with name filter", func() {
			It("returns only matching preview", func() {
				params := stores.ListParams{
					Filters: []stores.Filter{{Field: "name", Value: "preview-pr-30"}},
				}
				result, err := store.ListByConfigID(ctx, configA, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Items).To(HaveLen(1))
				Expect(result.Items[0].PRNumber).To(Equal("30"))
			})
		})
	})
})
