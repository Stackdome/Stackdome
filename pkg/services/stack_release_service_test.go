package services

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/auth"
	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
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

var _ = Describe("stackReleaseService release creation records release_created", func() {
	const (
		createEventsStackID = "stack-1"
		createEventsOrgID   = "org-1"
		createEventsTeamID  = "team-1"
		createEventsUserID  = "user-1"
		rollbackFromRelID   = "rel-src"
	)

	var (
		ctrl         *gomock.Controller
		releaseStore *mocks.MockStackReleaseStore
		stackSvc     *MockStackService
		perms        *mocks.MockPermissionService
		referenceSvc *mocks.MockReferenceService
		recorder     *MockReleaseEventRecorder
		enqueuer     *mocks.MockBackgroundJobEnqueuer
		svc          *stackReleaseService
		ctx          context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		releaseStore = mocks.NewMockStackReleaseStore(ctrl)
		stackSvc = NewMockStackService(ctrl)
		perms = mocks.NewMockPermissionService(ctrl)
		referenceSvc = mocks.NewMockReferenceService(ctrl)
		recorder = NewMockReleaseEventRecorder(ctrl)
		enqueuer = mocks.NewMockBackgroundJobEnqueuer(ctrl)
		svc = &stackReleaseService{
			store:            releaseStore,
			stackQuery:       stackSvc,
			permissions:      perms,
			referenceService: referenceSvc,
			eventRecorder:    recorder,
			BackgroundJobEnqueuerDep: BackgroundJobEnqueuerDep{
				BackgroundJobEnqueuer: enqueuer,
			},
		}
		ctx = context.Background()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	// runTxInline stubs WithTransaction to execute the closure inline with the
	// same context, so the ambient-transaction body is exercised directly.
	runTxInline := func() {
		releaseStore.EXPECT().WithTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			})
	}

	Describe("fresh create (createReleaseForStack)", func() {
		var stack *models.Stack

		BeforeEach(func() {
			stack = &models.Stack{ID: createEventsStackID, OrganisationID: createEventsOrgID}
		})

		It("records release_created inside the creation transaction as its final act", func() {
			created := &models.StackRelease{ID: "rel-1", StackID: createEventsStackID}
			runTxInline()
			releaseStore.EXPECT().Create(ctx, gomock.Any()).Return(created, nil)
			referenceSvc.EXPECT().ProjectRelease(ctx, created).Return(nil)
			recorder.EXPECT().RecordReleaseCreated(ctx, created).Return(nil)
			enqueuer.EXPECT().EnqueueAfterCommit(ctx, &models.StackRelease{ID: "rel-1"}).Return(nil)

			got, serr := svc.createReleaseForStack(ctx, stack, models.ReleaseCause{Kind: models.ReleaseCauseManual}, createEventsUserID)
			Expect(serr).To(BeNil())
			Expect(got).To(Equal(created))
		})

		It("fails creation and hands back no release when recording release_created fails", func() {
			created := &models.StackRelease{ID: "rel-1", StackID: createEventsStackID}
			runTxInline()
			releaseStore.EXPECT().Create(ctx, gomock.Any()).Return(created, nil)
			referenceSvc.EXPECT().ProjectRelease(ctx, created).Return(nil)
			recorder.EXPECT().RecordReleaseCreated(ctx, created).Return(errors.GeneralError("event insert failed"))
			// EnqueueAfterCommit must never run: the transaction rolled back.

			got, serr := svc.createReleaseForStack(ctx, stack, models.ReleaseCause{Kind: models.ReleaseCauseManual}, createEventsUserID)
			Expect(serr).ToNot(BeNil())
			Expect(got).To(BeNil())
		})
	})

	Describe("rollback (RollbackRelease)", func() {
		BeforeEach(func() {
			releaseStore.EXPECT().
				GetByID(ctx, rollbackFromRelID).
				Return(&models.StackRelease{ID: rollbackFromRelID, StackID: createEventsStackID, State: models.ReleaseStateReleased, Sequence: 7}, nil)
			stackSvc.EXPECT().
				GetStack(ctx, createEventsStackID).
				Return(&models.Stack{ID: createEventsStackID, TeamID: createEventsTeamID}, nil)
			perms.EXPECT().
				Check(ctx, createEventsTeamID, auth.ResourceStacks, createEventsStackID, auth.ActionWrite).
				Return(nil)
		})

		It("records release_created inside the rollback transaction", func() {
			created := &models.StackRelease{ID: "rel-2", StackID: createEventsStackID}
			runTxInline()
			releaseStore.EXPECT().Create(ctx, gomock.Any()).Return(created, nil)
			referenceSvc.EXPECT().ProjectRelease(ctx, created).Return(nil)
			recorder.EXPECT().RecordReleaseCreated(ctx, created).Return(nil)
			enqueuer.EXPECT().EnqueueAfterCommit(ctx, &models.StackRelease{ID: "rel-2"}).Return(nil)

			got, serr := svc.RollbackRelease(ctx, createEventsStackID, rollbackFromRelID)
			Expect(serr).To(BeNil())
			Expect(got).To(Equal(created))
		})
	})
})

var _ = Describe("stackReleaseService.CancelRelease records release_cancelled", func() {
	const (
		cancelStackID   = "stack-1"
		cancelReleaseID = "rel-1"
		cancelTeamID    = "team-1"
	)

	var (
		ctrl         *gomock.Controller
		releaseStore *mocks.MockStackReleaseStore
		stackSvc     *MockStackService
		perms        *mocks.MockPermissionService
		recorder     *MockReleaseEventRecorder
		svc          *stackReleaseService
		ctx          context.Context
		rel          *models.StackRelease
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		releaseStore = mocks.NewMockStackReleaseStore(ctrl)
		stackSvc = NewMockStackService(ctrl)
		perms = mocks.NewMockPermissionService(ctrl)
		recorder = NewMockReleaseEventRecorder(ctrl)
		svc = &stackReleaseService{
			store:         releaseStore,
			stackQuery:    stackSvc,
			permissions:   perms,
			eventRecorder: recorder,
			logger:        logger.NewLoggerWithPrefix(context.Background(), "stack-release-service-test"),
		}
		ctx = context.Background()
		rel = &models.StackRelease{ID: cancelReleaseID, StackID: cancelStackID, Sequence: 3, State: models.ReleaseStatePending}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	// expectLookupAndPermission stubs the release lookup, stack lookup, and a
	// passing write-permission check that CancelRelease performs up front.
	expectLookupAndPermission := func() {
		releaseStore.EXPECT().GetByID(ctx, cancelReleaseID).Return(rel, nil)
		stackSvc.EXPECT().GetStack(ctx, cancelStackID).Return(&models.Stack{ID: cancelStackID, TeamID: cancelTeamID}, nil)
		perms.EXPECT().Check(ctx, cancelTeamID, auth.ResourceStacks, cancelStackID, auth.ActionWrite).Return(nil)
	}

	It("records release_cancelled when the cancel CAS is won", func() {
		expectLookupAndPermission()
		releaseStore.EXPECT().Cancel(ctx, cancelReleaseID).Return(true, nil)
		recorder.EXPECT().RecordReleaseTerminal(ctx, rel, models.ReleaseStateCancelled, releaseCancelledMessage).Return(nil)

		serr := svc.CancelRelease(ctx, cancelReleaseID)
		Expect(serr).To(BeNil())
	})

	It("records nothing when the cancel CAS is lost", func() {
		expectLookupAndPermission()
		releaseStore.EXPECT().Cancel(ctx, cancelReleaseID).Return(false, nil)
		// recorder must never be called: no EXPECT registered.

		serr := svc.CancelRelease(ctx, cancelReleaseID)
		Expect(serr).ToNot(BeNil())
		Expect(serr.Code).To(Equal(errors.ErrorConflict))
	})

	It("does not fail the cancel when recording the event errors (log-only)", func() {
		expectLookupAndPermission()
		releaseStore.EXPECT().Cancel(ctx, cancelReleaseID).Return(true, nil)
		recorder.EXPECT().RecordReleaseTerminal(ctx, rel, models.ReleaseStateCancelled, releaseCancelledMessage).
			Return(errors.GeneralError("event insert failed"))

		serr := svc.CancelRelease(ctx, cancelReleaseID)
		Expect(serr).To(BeNil())
	})
})

var _ = Describe("stackReleaseService.ListReleaseEvents", func() {
	const (
		listEventsStackID   = "stack-1"
		listEventsReleaseID = "rel-1"
		listEventsTeamID    = "team-1"
	)

	var (
		ctrl         *gomock.Controller
		releaseStore *mocks.MockStackReleaseStore
		stackSvc     *MockStackService
		perms        *mocks.MockPermissionService
		eventStore   *mocks.MockReleaseEventStore
		svc          *stackReleaseService
		ctx          context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		releaseStore = mocks.NewMockStackReleaseStore(ctrl)
		stackSvc = NewMockStackService(ctrl)
		perms = mocks.NewMockPermissionService(ctrl)
		eventStore = mocks.NewMockReleaseEventStore(ctrl)
		svc = &stackReleaseService{
			store:       releaseStore,
			stackQuery:  stackSvc,
			permissions: perms,
			eventStore:  eventStore,
		}
		ctx = context.Background()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	// expectOwnershipAndReadPermission stubs a release owned by listEventsStackID
	// and a passing read-permission check on that stack's team.
	expectOwnershipAndReadPermission := func() {
		releaseStore.EXPECT().
			GetByID(ctx, listEventsReleaseID).
			Return(&models.StackRelease{ID: listEventsReleaseID, StackID: listEventsStackID}, nil)
		stackSvc.EXPECT().
			GetStack(ctx, listEventsStackID).
			Return(&models.Stack{ID: listEventsStackID, TeamID: listEventsTeamID}, nil)
		perms.EXPECT().
			Check(ctx, listEventsTeamID, auth.ResourceStacks, listEventsStackID, auth.ActionRead).
			Return(nil)
	}

	It("returns NotFound and never touches the stack or event store when the release belongs to another stack", func() {
		releaseStore.EXPECT().
			GetByID(ctx, listEventsReleaseID).
			Return(&models.StackRelease{ID: listEventsReleaseID, StackID: "other-stack"}, nil)

		page, serr := svc.ListReleaseEvents(ctx, listEventsStackID, listEventsReleaseID, 0, 0)
		Expect(page).To(BeNil())
		Expect(serr).ToNot(BeNil())
		Expect(serr.Code).To(Equal(errors.ErrorNotFound))
	})

	It("checks read permission on the owning stack's team and stops on denial", func() {
		releaseStore.EXPECT().
			GetByID(ctx, listEventsReleaseID).
			Return(&models.StackRelease{ID: listEventsReleaseID, StackID: listEventsStackID}, nil)
		stackSvc.EXPECT().
			GetStack(ctx, listEventsStackID).
			Return(&models.Stack{ID: listEventsStackID, TeamID: listEventsTeamID}, nil)
		perms.EXPECT().
			Check(ctx, listEventsTeamID, auth.ResourceStacks, listEventsStackID, auth.ActionRead).
			Return(errors.Forbidden("nope"))

		page, serr := svc.ListReleaseEvents(ctx, listEventsStackID, listEventsReleaseID, 0, 0)
		Expect(page).To(BeNil())
		Expect(serr).ToNot(BeNil())
		Expect(serr.Code).To(Equal(errors.ErrorForbidden))
	})

	It("clamps a zero limit to the default and derives the cursor from the last event", func() {
		expectOwnershipAndReadPermission()
		eventStore.EXPECT().
			ListByReleaseID(ctx, listEventsReleaseID, 0, releaseEventsDefaultLimit).
			Return([]*models.ReleaseEvent{{Sequence: 3}, {Sequence: 4}}, nil)

		page, serr := svc.ListReleaseEvents(ctx, listEventsStackID, listEventsReleaseID, 0, 0)
		Expect(serr).To(BeNil())
		Expect(page.Events).To(HaveLen(2))
		Expect(page.NextAfterSequence).To(Equal(4))
	})

	It("clamps an oversized limit to the maximum", func() {
		expectOwnershipAndReadPermission()
		eventStore.EXPECT().
			ListByReleaseID(ctx, listEventsReleaseID, 0, releaseEventsMaxLimit).
			Return(nil, nil)

		page, serr := svc.ListReleaseEvents(ctx, listEventsStackID, listEventsReleaseID, 0, 9999)
		Expect(serr).To(BeNil())
		Expect(page.NextAfterSequence).To(Equal(0))
	})

	It("keeps the requested afterSequence as the cursor when the page is empty", func() {
		expectOwnershipAndReadPermission()
		eventStore.EXPECT().
			ListByReleaseID(ctx, listEventsReleaseID, 12, releaseEventsDefaultLimit).
			Return([]*models.ReleaseEvent{}, nil)

		page, serr := svc.ListReleaseEvents(ctx, listEventsStackID, listEventsReleaseID, 12, 0)
		Expect(serr).To(BeNil())
		Expect(page.NextAfterSequence).To(Equal(12))
	})
})
