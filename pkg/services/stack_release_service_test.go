package services

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

// Ginkgo supports exactly one RunSpecs call per test binary; this package's
// suite is bootstrapped by TestAESEncryptionService in
// encryption_service_test.go, which discovers every Describe below.

const (
	resolvePinsTestOrgID   = "org-1"
	resolvePinsTestRepoURL = "https://github.com/acme/web.git"
)

// stackWithGitResource builds a single-resource stack whose "web" resource's
// git source revision is rev.
func stackWithGitResource(rev models.GitRevision) *models.Stack {
	return &models.Stack{
		ID:             "stack-1",
		OrganisationID: resolvePinsTestOrgID,
		StackResources: []*models.StackResource{
			{
				Name: "web",
				BuildConfig: &models.BuildConfigSpec{
					SourceContext: models.BuildContextSource{
						Git: &models.GitBuildSource{RepoURL: resolvePinsTestRepoURL},
					},
					SourceRevision: models.BuildSourceRevision{
						Git: &rev,
					},
				},
			},
		},
	}
}

var _ = Describe("stackReleaseService.resolvePins", func() {
	var (
		ctrl         *gomock.Controller
		credResolver *mocks.MockCredentialResolver
		gitClients   *MocksourceGitClientProvider
		gitClient    *mocks.MockGitClient
		svc          *stackReleaseService
		ctx          context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		credResolver = mocks.NewMockCredentialResolver(ctrl)
		gitClients = NewMocksourceGitClientProvider(ctrl)
		gitClient = mocks.NewMockGitClient(ctrl)
		svc = &stackReleaseService{
			credentialResolver: credResolver,
			gitClients:         gitClients,
		}
		ctx = context.Background()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	// expectAnonymousResolution stubs the credential resolution + client
	// construction steps that every non-commit-pinned resolution goes through.
	expectAnonymousResolution := func() {
		credResolver.EXPECT().
			GitCredentials(ctx, resolvePinsTestOrgID, resolvePinsTestRepoURL, credentials.GitAuthSelector{}).
			Return(&credentials.ResolvedGitCredential{Source: credentials.SourceAnonymous}, nil)
		gitClients.EXPECT().
			ClientFor(resolvePinsTestRepoURL, gitclient.GitCredentials{}).
			Return(gitClient, nil)
	}

	It("resolves a branch pin to the branch head SHA", func() {
		stack := stackWithGitResource(models.GitRevision{Branch: "main"})
		expectAnonymousResolution()
		gitClient.EXPECT().
			GetBranchHeadSHA(ctx, resolvePinsTestRepoURL, "main").
			Return(&gitclient.RepoResult{HeadSHA: "sha-main"}, nil)

		pins, serr := svc.resolvePins(ctx, stack)
		Expect(serr).To(BeNil())
		Expect(pins.Resources["web"].GitSHA).To(Equal("sha-main"))
		Expect(pins.Resources["web"].Branch).To(BeEmpty())
	})

	It("falls back to the repository default branch and writes it back to the snapshot", func() {
		stack := stackWithGitResource(models.GitRevision{})
		expectAnonymousResolution()
		gitClient.EXPECT().
			GetDefaultBranch(ctx, resolvePinsTestRepoURL).
			Return("trunk", nil)
		gitClient.EXPECT().
			GetBranchHeadSHA(ctx, resolvePinsTestRepoURL, "trunk").
			Return(&gitclient.RepoResult{HeadSHA: "sha-trunk"}, nil)

		pins, serr := svc.resolvePins(ctx, stack)
		Expect(serr).To(BeNil())
		Expect(pins.Resources["web"].GitSHA).To(Equal("sha-trunk"))
		Expect(pins.Resources["web"].Branch).To(Equal("trunk"))

		snapshot, err := models.NewStackSnapshot(stack)
		Expect(err).ToNot(HaveOccurred())
		applyPinsToSnapshot(&snapshot, pins)

		gitRev := snapshot.Resources[0].BuildConfig.SourceRevision.Git
		Expect(gitRev.Branch).To(Equal("trunk"))
		Expect(gitRev.Commit).To(Equal("sha-trunk"))
	})

	It("returns a structured VErrGitAuthFailed error when branch resolution fails auth", func() {
		stack := stackWithGitResource(models.GitRevision{Branch: "main"})
		expectAnonymousResolution()
		gitClient.EXPECT().
			GetBranchHeadSHA(ctx, resolvePinsTestRepoURL, "main").
			Return(nil, fmt.Errorf("wrapped: %w", gitclient.ErrAuthFailed))

		pins, serr := svc.resolvePins(ctx, stack)
		Expect(pins.Resources).To(BeEmpty())
		Expect(serr).ToNot(BeNil())
		Expect(serr.Code).To(Equal(errors.ErrorValidation))
		details, ok := serr.Details.(errors.ValidationErrorDetails)
		Expect(ok).To(BeTrue())
		Expect(details.Errors).To(HaveLen(1))
		Expect(details.Errors[0].Code).To(Equal(errors.VErrGitAuthFailed))
		Expect(details.Errors[0].Field).To(Equal("resources[web].source.git.branch"))
	})

	It("returns VErrGitBranchNotFound for a non-auth, non-rate-limit branch failure", func() {
		stack := stackWithGitResource(models.GitRevision{Branch: "does-not-exist"})
		expectAnonymousResolution()
		gitClient.EXPECT().
			GetBranchHeadSHA(ctx, resolvePinsTestRepoURL, "does-not-exist").
			Return(nil, gitclient.ErrNotFound)

		_, serr := svc.resolvePins(ctx, stack)
		Expect(serr).ToNot(BeNil())
		Expect(serr.Code).To(Equal(errors.ErrorValidation))
		details := serr.Details.(errors.ValidationErrorDetails)
		Expect(details.Errors[0].Code).To(Equal(errors.VErrGitBranchNotFound))
		Expect(details.Errors[0].Field).To(Equal("resources[web].source.git.branch"))
	})

	It("returns VErrGitRateLimited when the git host rate-limits the branch lookup", func() {
		stack := stackWithGitResource(models.GitRevision{Branch: "main"})
		expectAnonymousResolution()
		gitClient.EXPECT().
			GetBranchHeadSHA(ctx, resolvePinsTestRepoURL, "main").
			Return(nil, fmt.Errorf("wrapped: %w", gitclient.ErrRateLimited))

		_, serr := svc.resolvePins(ctx, stack)
		Expect(serr).ToNot(BeNil())
		Expect(serr.Code).To(Equal(errors.ErrorValidation))
		details := serr.Details.(errors.ValidationErrorDetails)
		Expect(details.Errors[0].Code).To(Equal(errors.VErrGitRateLimited))
	})

	It("returns VErrGitTagNotFound when the pinned tag cannot be resolved", func() {
		stack := stackWithGitResource(models.GitRevision{Tag: "v1.0.0"})
		expectAnonymousResolution()
		gitClient.EXPECT().
			GetTagSHA(ctx, resolvePinsTestRepoURL, "v1.0.0").
			Return("", gitclient.ErrNotFound)

		_, serr := svc.resolvePins(ctx, stack)
		Expect(serr).ToNot(BeNil())
		Expect(serr.Code).To(Equal(errors.ErrorValidation))
		details := serr.Details.(errors.ValidationErrorDetails)
		Expect(details.Errors[0].Code).To(Equal(errors.VErrGitTagNotFound))
		Expect(details.Errors[0].Field).To(Equal("resources[web].source.git.tag"))
	})
})
