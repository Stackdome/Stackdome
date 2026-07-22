package pgstore_test

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StackPreviewConfigStore", func() {
	var (
		store stores.StackPreviewConfigStore
		ctx   context.Context
	)

	const (
		orgID       = "org-1"
		otherOrgID  = "org-2"
		projectID   = "project-1"
		userID      = "user-1"
		repoURL     = "https://github.com/acme/app.git"
		repoURLSlug = "https://github.com/acme/app/"
	)

	BeforeEach(func() {
		sf := newSQLiteSessionFactory(`
			CREATE TABLE IF NOT EXISTS stack_preview_configs (
				id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
				organisation_id TEXT NOT NULL,
				project_id TEXT NOT NULL,
				user_id TEXT NOT NULL,
				name TEXT NOT NULL,
				description TEXT,
				git_repository TEXT NOT NULL,
				stackfile_path TEXT NOT NULL DEFAULT 'stackfile.yaml',
				max_active_previews INTEGER NOT NULL DEFAULT 10,
				env TEXT,
				repo_url_normalized TEXT NOT NULL DEFAULT '',
				labels TEXT,
				annotations TEXT,
				created_at DATETIME,
				updated_at DATETIME,
				UNIQUE (project_id, name),
				UNIQUE (organisation_id, repo_url_normalized)
			)
		`)
		store = pgstore.NewStackPreviewConfigStore(pgstore.StackPreviewConfigStoreSpec{SessionFactory: sf})
		ctx = context.Background()
	})

	Describe("GetByOrgAndRepo", func() {
		BeforeEach(func() {
			_, err := store.Create(ctx, &models.StackPreviewConfig{
				OrganisationID: orgID,
				ProjectID:      projectID,
				UserID:         userID,
				Name:           "config-a",
				GitRepository: models.PreviewGitRepository{
					RepoURL:    repoURL,
					BaseBranch: models.DefaultBaseBranch,
				},
			})
			Expect(err).To(BeNil())
		})

		It("finds the config by its normalized repo URL, regardless of .git/trailing-slash differences", func() {
			got, err := store.GetByOrgAndRepo(ctx, orgID, models.NormalizeRepoURL(repoURLSlug))
			Expect(err).To(BeNil())
			Expect(got.Name).To(Equal("config-a"))
			Expect(got.RepoURLNormalized).To(Equal(models.NormalizeRepoURL(repoURL)))
		})

		It("returns not found for a different organisation", func() {
			_, err := store.GetByOrgAndRepo(ctx, otherOrgID, models.NormalizeRepoURL(repoURL))
			Expect(err).ToNot(BeNil())
			Expect(err.Is404()).To(BeTrue())
		})

		It("returns not found when no config matches the repo", func() {
			_, err := store.GetByOrgAndRepo(ctx, orgID, models.NormalizeRepoURL("https://github.com/acme/other.git"))
			Expect(err).ToNot(BeNil())
			Expect(err.Is404()).To(BeTrue())
		})

		It("rejects a second config in the same org with an equivalent repo URL", func() {
			_, err := store.Create(ctx, &models.StackPreviewConfig{
				OrganisationID: orgID,
				ProjectID:      projectID,
				UserID:         userID,
				Name:           "config-b",
				GitRepository: models.PreviewGitRepository{
					RepoURL:    repoURLSlug,
					BaseBranch: models.DefaultBaseBranch,
				},
			})
			Expect(err).ToNot(BeNil())
		})
	})
})
