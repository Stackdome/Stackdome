package preview

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"k8s.io/utils/ptr"
)

var _ = Describe("ProvisionReconciler", func() {
	var (
		ctrl           *gomock.Controller
		previewStore   *MockpreviewStackStore
		previewService *MockpreviewStackService
		cfgStore       *MockconfigStore
		stackSvc       *MockstackService
		releaseSvc     *MockreleaseService
		reconciler     *provisionReconciler
		ctx            context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		previewStore = NewMockpreviewStackStore(ctrl)
		previewService = NewMockpreviewStackService(ctrl)
		cfgStore = NewMockconfigStore(ctrl)
		stackSvc = NewMockstackService(ctrl)
		releaseSvc = NewMockreleaseService(ctrl)
		ctx = context.Background()

		cache := &sync.Map{}
		cacheKeys := &sync.Map{}
		reconciler = &provisionReconciler{
			previewStackService: previewService,
			previewStackStore:   previewStore,
			configStore:         cfgStore,
			stackService:        stackSvc,
			releaseService:      releaseSvc,
			stackfileCache:      cache,
			previewCacheKeys:    cacheKeys,
			logger:              logger.NewLoggerWithPrefix(ctx, "test"),
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("create path (StackID is nil)", func() {
		var (
			preview *models.PreviewStack
			config  *models.StackPreviewConfig
		)

		BeforeEach(func() {
			preview = &models.PreviewStack{
				ID:                   "p-1",
				StackPreviewConfigID: "cfg-1",
				PRNumber:             "42",
				CommitSHA:            "abc1234def5678",
				Status:               models.PreviewStackStatus{Phase: models.PreviewStackPhaseProvisioning},
			}
			config = &models.StackPreviewConfig{
				ID:   "cfg-1",
				Name: "test-config",
			}
		})

		It("creates stack and release, sets Deploying phase", func() {
			stackfileContent := []byte("name: my-app")
			stackfileHash := "hash-abc"
			builtStack := &models.Stack{Name: "my-app"}
			createdStack := &models.Stack{ID: "stack-1", Name: "my-app"}
			createdRelease := &models.StackRelease{ID: "rel-1"}

			cfgStore.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)
			previewService.EXPECT().InternalFetchStackfile(gomock.Any(), config, preview.CommitSHA).
				Return(stackfileContent, stackfileHash, nil)
			previewService.EXPECT().InternalBuildStackFromContent(gomock.Any(), config, preview, stackfileContent).
				Return(builtStack, nil)

			previewStore.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
					return fn(ctx)
				})
			stackSvc.EXPECT().InternalCreateStack(gomock.Any(), builtStack).Return(createdStack, nil)
			releaseSvc.EXPECT().InternalCreateRelease(gomock.Any(), "stack-1", gomock.Any()).Return(createdRelease, nil)
			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.StackID).ToNot(BeNil())
					Expect(*p.StackID).To(Equal("stack-1"))
					Expect(p.ActiveReleaseID).ToNot(BeNil())
					Expect(*p.ActiveReleaseID).To(Equal("rel-1"))
					Expect(p.ReconcilerStatus.LastAppliedStackfileHash).To(Equal(stackfileHash))
					Expect(p.ReconcilerStatus.LastAppliedCommitSHA).To(Equal(preview.CommitSHA))

					Expect(p.Status.Phase).To(Equal(models.PreviewStackPhaseDeploying))
					Expect(p.Status.Reason).To(Equal("StackCreated"))
					return p, nil
				})

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultRequeueAfter(convergePollInterval)))
		})

		It("fails permanently when stackfile is not found", func() {
			cfgStore.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)
			previewService.EXPECT().InternalFetchStackfile(gomock.Any(), config, preview.CommitSHA).
				Return(nil, "", errors.Permanent("StackfileNotFound", "stackfile.yaml not found"))

			// fail() calls Update to set Failed status
			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.Status.Phase).To(Equal(models.PreviewStackPhaseFailed))
					Expect(p.Status.Reason).To(Equal("StackfileNotFound"))
					return p, nil
				})

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultStop))
		})

		It("returns error for retry on transient fetch error", func() {
			cfgStore.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)
			transientErr := errors.Transient("FetchFailed", "network timeout")
			previewService.EXPECT().InternalFetchStackfile(gomock.Any(), config, preview.CommitSHA).
				Return(nil, "", transientErr)

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).To(MatchError(transientErr))
			Expect(result).To(Equal(resultNil))
		})

		It("fails permanently when stackfile is invalid (parse error)", func() {
			stackfileContent := []byte("invalid: {{")
			stackfileHash := "hash-bad"

			cfgStore.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)
			previewService.EXPECT().InternalFetchStackfile(gomock.Any(), config, preview.CommitSHA).
				Return(stackfileContent, stackfileHash, nil)
			previewService.EXPECT().InternalBuildStackFromContent(gomock.Any(), config, preview, stackfileContent).
				Return(nil, errors.Permanent("InvalidStackfile", "failed to parse stackfile"))

			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.Status.Phase).To(Equal(models.PreviewStackPhaseFailed))
					Expect(p.Status.Reason).To(Equal("InvalidStackfile"))
					return p, nil
				})

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultStop))
		})

		It("returns error for retry when transaction fails", func() {
			stackfileContent := []byte("name: my-app")
			stackfileHash := "hash-abc"
			builtStack := &models.Stack{Name: "my-app"}

			cfgStore.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)
			previewService.EXPECT().InternalFetchStackfile(gomock.Any(), config, preview.CommitSHA).
				Return(stackfileContent, stackfileHash, nil)
			previewService.EXPECT().InternalBuildStackFromContent(gomock.Any(), config, preview, stackfileContent).
				Return(builtStack, nil)

			txErr := errors.GeneralError("database connection lost")
			previewStore.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).Return(txErr)

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("provision transaction failed"))
			Expect(result).To(Equal(resultNil))
		})

		It("uses inline StackfileContent instead of fetching, creates stack and release", func() {
			preview.StackfileContent = ptr.To("name: inline-app")
			inlineContent := []byte("name: inline-app")
			h := sha256.Sum256(inlineContent)
			expectedHash := hex.EncodeToString(h[:])

			builtStack := &models.Stack{Name: "inline-app"}
			createdStack := &models.Stack{ID: "stack-1", Name: "inline-app"}
			createdRelease := &models.StackRelease{ID: "rel-1"}

			cfgStore.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)
			// InternalFetchStackfile should NOT be called
			previewService.EXPECT().InternalBuildStackFromContent(gomock.Any(), config, preview, inlineContent).
				Return(builtStack, nil)

			previewStore.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
					return fn(ctx)
				})
			stackSvc.EXPECT().InternalCreateStack(gomock.Any(), builtStack).Return(createdStack, nil)
			releaseSvc.EXPECT().InternalCreateRelease(gomock.Any(), "stack-1", gomock.Any()).Return(createdRelease, nil)
			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.StackID).ToNot(BeNil())
					Expect(*p.StackID).To(Equal("stack-1"))
					Expect(p.ActiveReleaseID).ToNot(BeNil())
					Expect(*p.ActiveReleaseID).To(Equal("rel-1"))
					Expect(p.ReconcilerStatus.LastAppliedStackfileHash).To(Equal(expectedHash))
					Expect(p.ReconcilerStatus.LastAppliedCommitSHA).To(Equal(preview.CommitSHA))

					Expect(p.Status.Phase).To(Equal(models.PreviewStackPhaseDeploying))
					Expect(p.Status.Reason).To(Equal("StackCreated"))
					return p, nil
				})

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultRequeueAfter(convergePollInterval)))
		})
	})

	Context("sync path (StackID is set)", func() {
		var (
			preview *models.PreviewStack
			config  *models.StackPreviewConfig
			stackID string
		)

		BeforeEach(func() {
			stackID = "stack-existing"
			preview = &models.PreviewStack{
				ID:                   "p-2",
				StackPreviewConfigID: "cfg-1",
				StackID:              &stackID,
				PRNumber:             "42",
				CommitSHA:            "abc1234def5678",
				ReconcilerStatus: models.PreviewReconcilerStatus{
					LastAppliedStackfileHash: "old-hash",
					LastAppliedCommitSHA:     "abc1234def5678",
				},
				Status: models.PreviewStackStatus{Phase: models.PreviewStackPhaseProvisioning},
			}
			config = &models.StackPreviewConfig{
				ID:   "cfg-1",
				Name: "test-config",
			}
		})

		It("returns resultNil when stackfile hash and lastAppliedCommit is unchanged and no image overrides (idempotent noop)", func() {
			stackfileContent := []byte("name: my-app")
			sameHash := "old-hash"

			cfgStore.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)
			previewService.EXPECT().InternalFetchStackfile(gomock.Any(), config, preview.CommitSHA).
				Return(stackfileContent, sameHash, nil)

			// No InternalBuildStackFromContent, InternalUpdateStack, transaction, or release expected

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultNil))
		})

		It("calls UpdateStack when stackfile hash changed", func() {
			stackfileContent := []byte("name: updated-app")
			newHash := "new-hash"
			builtStack := &models.Stack{Name: "updated-app"}
			updatedStack := &models.Stack{ID: stackID, Name: "updated-app"}
			createdRelease := &models.StackRelease{ID: "rel-3"}

			cfgStore.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)
			previewService.EXPECT().InternalFetchStackfile(gomock.Any(), config, preview.CommitSHA).
				Return(stackfileContent, newHash, nil)
			previewService.EXPECT().InternalBuildStackFromContent(gomock.Any(), config, preview, stackfileContent).
				Return(builtStack, nil)

			previewStore.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
					return fn(ctx)
				})
			stackSvc.EXPECT().InternalUpdateStack(gomock.Any(), stackID, builtStack).Return(updatedStack, nil)
			releaseSvc.EXPECT().InternalCreateRelease(gomock.Any(), stackID, gomock.Any()).Return(createdRelease, nil)
			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.ReconcilerStatus.LastAppliedStackfileHash).To(Equal(newHash))
					Expect(p.ReconcilerStatus.LastAppliedCommitSHA).To(Equal(preview.CommitSHA))
					Expect(p.ReconcilerStatus.LastAppliedOverridesHash).To(Equal("")) // no overrides

					Expect(p.Status.Phase).To(Equal(models.PreviewStackPhaseDeploying))
					return p, nil
				})

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultRequeueAfter(convergePollInterval)))
		})

		It("calls UpdateStack when image overrides are present", func() {
			stackfileContent := []byte("name: my-app")
			sameHash := "old-hash"
			preview.ImageOverrides = models.ImageOverrides{"web": "myapp:pr-42"}
			// LastAppliedOverridesHash defaults to "" which differs from the non-empty overrides hash

			builtStack := &models.Stack{Name: "my-app"}
			updatedStack := &models.Stack{ID: stackID, Name: "my-app"}
			createdRelease := &models.StackRelease{ID: "rel-4"}

			cfgStore.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)
			previewService.EXPECT().InternalFetchStackfile(gomock.Any(), config, preview.CommitSHA).
				Return(stackfileContent, sameHash, nil)
			previewService.EXPECT().InternalBuildStackFromContent(gomock.Any(), config, preview, stackfileContent).
				Return(builtStack, nil)

			previewStore.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
					return fn(ctx)
				})
			stackSvc.EXPECT().InternalUpdateStack(gomock.Any(), stackID, builtStack).Return(updatedStack, nil)
			releaseSvc.EXPECT().InternalCreateRelease(gomock.Any(), stackID, gomock.Any()).Return(createdRelease, nil)
			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.ReconcilerStatus.LastAppliedCommitSHA).To(Equal(preview.CommitSHA))
					Expect(p.ReconcilerStatus.LastAppliedOverridesHash).ToNot(BeEmpty())
					Expect(p.Status.Phase).To(Equal(models.PreviewStackPhaseDeploying))
					return p, nil
				})

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultRequeueAfter(convergePollInterval)))
		})

		It("calls UpdateStack with inline StackfileContent when hash changed", func() {
			preview.StackfileContent = ptr.To("name: updated-inline-app")
			inlineContent := []byte("name: updated-inline-app")
			h := sha256.Sum256(inlineContent)
			newHash := hex.EncodeToString(h[:])

			builtStack := &models.Stack{Name: "updated-inline-app"}
			updatedStack := &models.Stack{ID: stackID, Name: "updated-inline-app"}
			createdRelease := &models.StackRelease{ID: "rel-5"}

			cfgStore.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)
			// InternalFetchStackfile should NOT be called
			previewService.EXPECT().InternalBuildStackFromContent(gomock.Any(), config, preview, inlineContent).
				Return(builtStack, nil)

			previewStore.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
					return fn(ctx)
				})
			stackSvc.EXPECT().InternalUpdateStack(gomock.Any(), stackID, builtStack).Return(updatedStack, nil)
			releaseSvc.EXPECT().InternalCreateRelease(gomock.Any(), stackID, gomock.Any()).Return(createdRelease, nil)
			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.ReconcilerStatus.LastAppliedStackfileHash).To(Equal(newHash))
					Expect(p.ReconcilerStatus.LastAppliedCommitSHA).To(Equal(preview.CommitSHA))
					Expect(p.ReconcilerStatus.LastAppliedOverridesHash).To(Equal("")) // no overrides

					Expect(p.Status.Phase).To(Equal(models.PreviewStackPhaseDeploying))
					return p, nil
				})

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultRequeueAfter(convergePollInterval)))
		})

		It("returns resultNil with inline StackfileContent when hash is unchanged (idempotent noop)", func() {
			inlineContent := []byte("name: same-content")
			h := sha256.Sum256(inlineContent)
			unchangedHash := hex.EncodeToString(h[:])

			preview.ReconcilerStatus.LastAppliedStackfileHash = unchangedHash
			preview.StackfileContent = ptr.To("name: same-content")

			cfgStore.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)
			// InternalFetchStackfile should NOT be called (inline content used)
			// No InternalBuildStackFromContent, transaction, or release expected

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultNil))
		})

		It("calls UpdateStack when commit SHA changed but stackfile hash unchanged", func() {
			// Set up: stackfile hash matches, but commit SHA differs from last applied
			preview.CommitSHA = "newcommit1234567"
			preview.ReconcilerStatus.LastAppliedCommitSHA = "abc1234def5678"
			preview.ReconcilerStatus.LastAppliedStackfileHash = "old-hash"

			stackfileContent := []byte("name: my-app")
			sameHash := "old-hash" // same as StackfileHash
			builtStack := &models.Stack{Name: "my-app"}
			updatedStack := &models.Stack{ID: stackID, Name: "my-app"}
			createdRelease := &models.StackRelease{ID: "rel-commit"}

			cfgStore.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)
			previewService.EXPECT().InternalFetchStackfile(gomock.Any(), config, preview.CommitSHA).
				Return(stackfileContent, sameHash, nil)
			previewService.EXPECT().InternalBuildStackFromContent(gomock.Any(), config, preview, stackfileContent).
				Return(builtStack, nil)

			previewStore.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
					return fn(ctx)
				})
			stackSvc.EXPECT().InternalUpdateStack(gomock.Any(), stackID, builtStack).Return(updatedStack, nil)
			releaseSvc.EXPECT().InternalCreateRelease(gomock.Any(), stackID, gomock.Any()).Return(createdRelease, nil)
			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.ReconcilerStatus.LastAppliedCommitSHA).To(Equal("newcommit1234567"))
					Expect(p.ReconcilerStatus.LastAppliedOverridesHash).To(Equal("")) // no overrides

					Expect(p.Status.Phase).To(Equal(models.PreviewStackPhaseDeploying))
					return p, nil
				})

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultRequeueAfter(convergePollInterval)))
		})

		It("returns resultNil when commit, hash, and overrides are all unchanged (idempotent noop)", func() {
			// Everything matches - needsUpdate=false, ForceSync=false
			preview.CommitSHA = "abc1234def5678"
			preview.ReconcilerStatus.LastAppliedCommitSHA = "abc1234def5678"
			preview.ReconcilerStatus.LastAppliedStackfileHash = "old-hash"
			// No image overrides, LastAppliedOverridesHash defaults to ""

			stackfileContent := []byte("name: my-app")
			sameHash := "old-hash"

			cfgStore.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)
			previewService.EXPECT().InternalFetchStackfile(gomock.Any(), config, preview.CommitSHA).
				Return(stackfileContent, sameHash, nil)

			// No InternalBuildStackFromContent, transaction, or release expected

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultNil))
		})

		It("uses cached stackfile on second reconcile without duplicate fetch", func() {
			stackfileContent := []byte("name: my-app")
			stackfileHash := "old-hash"

			config.GitRepository = models.PreviewGitRepository{RepoURL: "https://github.com/test/repo"}
			config.StackfilePath = "stackfile.yaml"

			// Pre-populate the cache with the same key resolveStackfileContent would use
			cacheKey := config.GitRepository.RepoURL + ":" + preview.CommitSHA + ":" + config.StackfilePath
			reconciler.stackfileCache.Store(cacheKey, stackfileCacheEntry{content: stackfileContent, hash: stackfileHash})

			// InternalFetchStackfile should NOT be called — cache hit
			// needsUpdate = false (hash matches), needsRelease = false → noop
			cfgStore.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultNil))
		})

		It("creates release when ForceSyncRequestedAt is pending despite no changes", func() {
			preview.CommitSHA = "abc1234def5678"
			preview.ReconcilerStatus.LastAppliedCommitSHA = "abc1234def5678"
			preview.ReconcilerStatus.LastAppliedStackfileHash = "old-hash"
			preview.ForceSyncRequestedAt = ptr.To(time.Now().UTC())

			stackfileContent := []byte("name: my-app")
			sameHash := "old-hash"
			createdRelease := &models.StackRelease{ID: "rel-force"}

			cfgStore.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)
			previewService.EXPECT().InternalFetchStackfile(gomock.Any(), config, preview.CommitSHA).
				Return(stackfileContent, sameHash, nil)

			// needsUpdate=false so no InternalBuildStackFromContent or InternalUpdateStack,
			// but needsRelease=true (ForceSync) so transaction + release are created

			previewStore.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
					return fn(ctx)
				})
			// No InternalUpdateStack expected (needsUpdate is false)
			releaseSvc.EXPECT().InternalCreateRelease(gomock.Any(), stackID, gomock.Any()).Return(createdRelease, nil)
			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.ActiveReleaseID).ToNot(BeNil())
					Expect(*p.ActiveReleaseID).To(Equal("rel-force"))

					Expect(p.ReconcilerStatus.LastAppliedCommitSHA).To(Equal("abc1234def5678"))
					Expect(p.ReconcilerStatus.LastAppliedOverridesHash).To(Equal(""))
					Expect(p.ReconcilerStatus.LastProcessedForceSyncAt).To(Equal(*preview.ForceSyncRequestedAt))
					Expect(p.Status.Phase).To(Equal(models.PreviewStackPhaseDeploying))
					Expect(p.Status.Reason).To(Equal("SyncTriggered"))
					return p, nil
				})

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultRequeueAfter(convergePollInterval)))
		})
	})

	It("returns Name() as provision", func() {
		Expect(reconciler.Name()).To(Equal("provision"))
	})

	Describe("fail helper", func() {
		It("sets preview to Failed and returns resultStop", func() {
			preview := &models.PreviewStack{
				ID:     "p-fail",
				Status: models.PreviewStackStatus{Phase: models.PreviewStackPhaseProvisioning},
			}

			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.Status.Phase).To(Equal(models.PreviewStackPhaseFailed))
					Expect(p.Status.Reason).To(Equal("TestReason"))
					Expect(p.Status.Message).To(Equal("something broke"))
					return p, nil
				})

			result, err := reconciler.fail(ctx, preview, "TestReason", "something broke")
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultStop))
		})
	})

	Describe("config store error", func() {
		It("returns error when config store fails", func() {
			preview := &models.PreviewStack{
				ID:                   "p-err",
				StackPreviewConfigID: "cfg-missing",
				Status:               models.PreviewStackStatus{Phase: models.PreviewStackPhaseProvisioning},
			}

			cfgStore.EXPECT().GetByID(gomock.Any(), "cfg-missing").
				Return(nil, errors.NotFound("config not found"))

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get config"))
			Expect(result).To(Equal(resultNil))
		})
	})
})

// helper to construct a ReleaseCause matcher; we use gomock.Any() instead for simplicity.
func newPreviewStackWithPhase(id string, phase models.PreviewStackPhase) *models.PreviewStack {
	return &models.PreviewStack{
		ID:     id,
		Status: models.PreviewStackStatus{Phase: phase},
	}
}
