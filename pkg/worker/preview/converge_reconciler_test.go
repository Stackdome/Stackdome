package preview

import (
	"context"
	stderrors "errors"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"k8s.io/utils/ptr"
)

var _ = Describe("ConvergeReconciler", func() {
	var (
		ctrl         *gomock.Controller
		previewStore   *MockpreviewStackStore
		stackSvc       *MockstackService
		releaseSvc     *MockreleaseService
		commentService *MockpreviewCommentService
		reconciler     *convergeReconciler
		ctx            context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		previewStore = NewMockpreviewStackStore(ctrl)
		stackSvc = NewMockstackService(ctrl)
		releaseSvc = NewMockreleaseService(ctrl)
		commentService = NewMockpreviewCommentService(ctrl)
		ctx = context.Background()

		reconciler = &convergeReconciler{
			releaseService:    releaseSvc,
			stackService:      stackSvc,
			previewStackStore: previewStore,
			commentService:    commentService,
			logger:            logger.NewLoggerWithPrefix(ctx, "test"),
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("noop cases", func() {
		It("returns resultNil when ActiveReleaseID is nil", func() {
			preview := &models.PreviewStack{
				ID:     "p-2",
				Status: models.PreviewStackStatus{Phase: models.PreviewStackPhaseDeploying},
			}
			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultNil))
		})

		It("returns resultNil with no DB update when release is Released and phase is already Ready", func() {
			releaseID := "rel-ready"
			preview := &models.PreviewStack{
				ID:              "p-ready",
				StackID:         ptr.To("stack-1"),
				ActiveReleaseID: &releaseID,
				Status:          models.PreviewStackStatus{Phase: models.PreviewStackPhaseReady},
			}

			releaseSvc.EXPECT().InternalGet(gomock.Any(), releaseID).
				Return(&models.StackRelease{ID: releaseID, State: models.ReleaseStateReleased}, nil)

			// No previewStore.Update expected — gomock strict mode enforces this

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultNil))
		})

		It("retries a pending comment on a Ready preview and clears the flag on success", func() {
			releaseID := "rel-ready"
			preview := &models.PreviewStack{
				ID:                   "p-ready",
				StackID:              ptr.To("stack-1"),
				ActiveReleaseID:      &releaseID,
				GitHubCommentPending: true,
				Status:               models.PreviewStackStatus{Phase: models.PreviewStackPhaseReady},
			}

			releaseSvc.EXPECT().InternalGet(gomock.Any(), releaseID).
				Return(&models.StackRelease{ID: releaseID, State: models.ReleaseStateReleased}, nil)
			stackSvc.EXPECT().InternalGetStack(gomock.Any(), "stack-1").
				Return(&models.Stack{ID: "stack-1"}, nil)
			commentService.EXPECT().InternalUpsertComment(gomock.Any(), preview).Return(nil)
			previewStore.EXPECT().Update(gomock.Any(), preview).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.GitHubCommentPending).To(BeFalse())
					return p, nil
				})

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultStop))
		})

		It("keeps the pending flag without a DB update when the comment retry fails again", func() {
			releaseID := "rel-ready"
			preview := &models.PreviewStack{
				ID:                   "p-ready",
				StackID:              ptr.To("stack-1"),
				ActiveReleaseID:      &releaseID,
				GitHubCommentPending: true,
				Status:               models.PreviewStackStatus{Phase: models.PreviewStackPhaseReady},
			}

			releaseSvc.EXPECT().InternalGet(gomock.Any(), releaseID).
				Return(&models.StackRelease{ID: releaseID, State: models.ReleaseStateReleased}, nil)
			stackSvc.EXPECT().InternalGetStack(gomock.Any(), "stack-1").
				Return(&models.Stack{ID: "stack-1"}, nil)
			commentService.EXPECT().InternalUpsertComment(gomock.Any(), preview).
				Return(stderrors.New("github down"))
			previewStore.EXPECT().Update(gomock.Any(), preview).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.GitHubCommentPending).To(BeTrue())
					return p, nil
				})

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultStop))
			Expect(preview.GitHubCommentPending).To(BeTrue())
		})
	})

	Context("pending public URLs", func() {
		var (
			preview   *models.PreviewStack
			stackID   string
			releaseID string
			release   *models.StackRelease
		)

		exposedResource := func(status *models.StackResourceStatus) *models.StackResource {
			return &models.StackResource{
				Name:   "web",
				Ports:  models.Ports{{Name: "http", Number: 8080, ExposedToPublic: true}},
				Status: status,
			}
		}

		BeforeEach(func() {
			stackID = "stack-1"
			releaseID = "rel-1"
			release = &models.StackRelease{ID: releaseID, State: models.ReleaseStateReleased}
			preview = &models.PreviewStack{
				ID:              "p-urls",
				StackID:         &stackID,
				ActiveReleaseID: &releaseID,
				CommitSHA:       "abc1234",
				Status:          models.PreviewStackStatus{Phase: models.PreviewStackPhaseDeploying},
			}
		})

		It("requeues and keeps the comment pending while an exposed resource has no URL", func() {
			stack := &models.Stack{
				ID:             stackID,
				StackResources: []*models.StackResource{exposedResource(&models.StackResourceStatus{})},
			}

			releaseSvc.EXPECT().InternalGet(gomock.Any(), releaseID).Return(release, nil)
			stackSvc.EXPECT().InternalGetStack(gomock.Any(), stackID).Return(stack, nil)
			commentService.EXPECT().InternalUpsertComment(gomock.Any(), gomock.Any()).Return(nil)
			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.Status.Phase).To(Equal(models.PreviewStackPhaseReady))
					Expect(p.Status.Outputs.URLs).To(BeEmpty())
					Expect(p.Status.Outputs.URLsPending).To(BeTrue())
					Expect(p.GitHubCommentPending).To(BeTrue())
					return p, nil
				})

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultRequeueAfter(convergePollInterval)))
		})

		It("treats an ingress with an empty URL as still pending", func() {
			stack := &models.Stack{
				ID: stackID,
				StackResources: []*models.StackResource{
					exposedResource(&models.StackResourceStatus{PublicIngresses: []models.Ingress{{TargetPort: 8080}}}),
				},
			}

			releaseSvc.EXPECT().InternalGet(gomock.Any(), releaseID).Return(release, nil)
			stackSvc.EXPECT().InternalGetStack(gomock.Any(), stackID).Return(stack, nil)
			commentService.EXPECT().InternalUpsertComment(gomock.Any(), gomock.Any()).Return(nil)
			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.Status.Outputs.URLs).To(BeEmpty())
					Expect(p.Status.Outputs.URLsPending).To(BeTrue())
					return p, nil
				})

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultRequeueAfter(convergePollInterval)))
		})

		It("finalises the comment on a later pass once the http fallback URL lands", func() {
			preview.Status = models.PreviewStackStatus{
				Phase:   models.PreviewStackPhaseReady,
				Reason:  "ReleaseConverged",
				Outputs: &models.PreviewStackOutputs{CommitSHA: "abc1234", URLsPending: true},
			}
			preview.GitHubCommentPending = true
			preview.GitHubCommentID = 4242

			stack := &models.Stack{
				ID: stackID,
				StackResources: []*models.StackResource{
					exposedResource(&models.StackResourceStatus{
						PublicIngresses: []models.Ingress{{URL: "http://pr-1-web.example.com", TargetPort: 8080}},
					}),
				},
			}

			releaseSvc.EXPECT().InternalGet(gomock.Any(), releaseID).Return(release, nil)
			stackSvc.EXPECT().InternalGetStack(gomock.Any(), stackID).Return(stack, nil)
			commentService.EXPECT().InternalUpsertComment(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) error {
					Expect(p.Status.Outputs.URLs).To(HaveLen(1))
					Expect(p.Status.Outputs.URLs[0].URL).To(Equal("http://pr-1-web.example.com"))
					return nil
				})
			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.Status.Outputs.URLsPending).To(BeFalse())
					Expect(p.GitHubCommentPending).To(BeFalse())
					return p, nil
				})

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultStop))
		})

		It("does not wait on resources with no publicly exposed port", func() {
			stack := &models.Stack{
				ID: stackID,
				StackResources: []*models.StackResource{
					{Name: "worker", Ports: models.Ports{{Name: "metrics", Number: 9090}}, Status: &models.StackResourceStatus{}},
				},
			}

			releaseSvc.EXPECT().InternalGet(gomock.Any(), releaseID).Return(release, nil)
			stackSvc.EXPECT().InternalGetStack(gomock.Any(), stackID).Return(stack, nil)
			commentService.EXPECT().InternalUpsertComment(gomock.Any(), gomock.Any()).Return(nil)
			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.Status.Outputs.URLsPending).To(BeFalse())
					Expect(p.GitHubCommentPending).To(BeFalse())
					return p, nil
				})

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultStop))
		})
	})

	Context("terminal cases", func() {
		var (
			preview   *models.PreviewStack
			stackID   string
			releaseID string
		)

		BeforeEach(func() {
			stackID = "stack-1"
			releaseID = "rel-1"
			preview = &models.PreviewStack{
				ID:              "p-3",
				StackID:         &stackID,
				ActiveReleaseID: &releaseID,
				CommitSHA:       "abc1234",
				Status:          models.PreviewStackStatus{Phase: models.PreviewStackPhaseDeploying},
			}
		})

		It("sets Ready phase with outputs when release is Released", func() {
			release := &models.StackRelease{
				ID:    releaseID,
				State: models.ReleaseStateReleased,
			}
			stack := &models.Stack{
				ID: stackID,
				StackResources: []*models.StackResource{
					{
						Name: "web",
						Status: &models.StackResourceStatus{
							PublicIngresses: []models.Ingress{
								{URL: "https://pr-1-web.example.com", TargetPort: 8080},
							},
						},
					},
					{
						Name: "api",
						Status: &models.StackResourceStatus{
							PublicIngresses: []models.Ingress{
								{URL: "https://pr-1-api.example.com", TargetPort: 3000},
							},
						},
					},
					{
						Name:   "worker",
						Status: nil, // no status, should be skipped
					},
				},
			}

			releaseSvc.EXPECT().InternalGet(gomock.Any(), releaseID).Return(release, nil)
			stackSvc.EXPECT().InternalGetStack(gomock.Any(), stackID).Return(stack, nil)
			commentService.EXPECT().InternalUpsertComment(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) error {
					p.GitHubCommentID = 4242
					return nil
				})
			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.GitHubCommentID).To(Equal(int64(4242)))
					Expect(p.GitHubCommentPending).To(BeFalse())
					Expect(p.Status.Phase).To(Equal(models.PreviewStackPhaseReady))
					Expect(p.Status.Reason).To(Equal("ReleaseConverged"))
					Expect(p.Status.Outputs).ToNot(BeNil())
					Expect(p.Status.Outputs.CommitSHA).To(Equal("abc1234"))
					Expect(p.Status.Outputs.URLs).To(HaveLen(2))
					Expect(p.Status.Outputs.URLs[0].Resource).To(Equal("web"))
					Expect(p.Status.Outputs.URLs[0].URL).To(Equal("https://pr-1-web.example.com"))
					Expect(p.Status.Outputs.URLs[1].Resource).To(Equal("api"))
					Expect(p.Status.Outputs.URLs[1].URL).To(Equal("https://pr-1-api.example.com"))
					return p, nil
				})

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultStop))
		})

		It("persists Ready with the comment marked pending when the upsert fails", func() {
			release := &models.StackRelease{
				ID:    releaseID,
				State: models.ReleaseStateReleased,
			}
			stack := &models.Stack{ID: stackID}

			releaseSvc.EXPECT().InternalGet(gomock.Any(), releaseID).Return(release, nil)
			stackSvc.EXPECT().InternalGetStack(gomock.Any(), stackID).Return(stack, nil)
			commentService.EXPECT().InternalUpsertComment(gomock.Any(), gomock.Any()).
				Return(stderrors.New("github down"))
			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.Status.Phase).To(Equal(models.PreviewStackPhaseReady))
					Expect(p.GitHubCommentPending).To(BeTrue())
					return p, nil
				})

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultStop))
		})

		It("sets Failed phase when release is Failed", func() {
			release := &models.StackRelease{
				ID:      releaseID,
				State:   models.ReleaseStateFailed,
				Message: "image pull backoff",
			}

			releaseSvc.EXPECT().InternalGet(gomock.Any(), releaseID).Return(release, nil)
			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.Status.Phase).To(Equal(models.PreviewStackPhaseFailed))
					Expect(p.Status.Reason).To(Equal("ReleaseFailed"))
					Expect(p.Status.Message).To(Equal("image pull backoff"))
					return p, nil
				})
			commentService.EXPECT().InternalUpsertComment(gomock.Any(), gomock.Any()).Return(nil)

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultStop))
		})

		It("preserves last-known outputs when release is Failed", func() {
			preview.Status.Outputs = &models.PreviewStackOutputs{
				CommitSHA: "abc1234",
				URLs:      []models.PreviewURL{{Resource: "web", URL: "https://pr-1-web.example.com"}},
			}
			release := &models.StackRelease{
				ID:      releaseID,
				State:   models.ReleaseStateFailed,
				Message: "image pull backoff",
			}

			releaseSvc.EXPECT().InternalGet(gomock.Any(), releaseID).Return(release, nil)
			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.Status.Phase).To(Equal(models.PreviewStackPhaseFailed))
					Expect(p.Status.Outputs).ToNot(BeNil())
					Expect(p.Status.Outputs.CommitSHA).To(Equal("abc1234"))
					Expect(p.Status.Outputs.URLs).To(HaveLen(1))
					Expect(p.Status.Outputs.URLs[0].URL).To(Equal("https://pr-1-web.example.com"))
					return p, nil
				})
			commentService.EXPECT().InternalUpsertComment(gomock.Any(), gomock.Any()).Return(nil)

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultStop))
		})

		It("sets Failed phase with ReleaseCancelled reason when release is Cancelled", func() {
			release := &models.StackRelease{
				ID:    releaseID,
				State: models.ReleaseStateCancelled,
			}

			releaseSvc.EXPECT().InternalGet(gomock.Any(), releaseID).Return(release, nil)
			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
					Expect(p.Status.Phase).To(Equal(models.PreviewStackPhaseFailed))
					Expect(p.Status.Reason).To(Equal("ReleaseCancelled"))
					Expect(p.Status.Message).To(Equal("release was cancelled"))
					return p, nil
				})

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultStop))
		})
	})

	Context("non-terminal cases", func() {
		var (
			preview   *models.PreviewStack
			stackID   string
			releaseID string
		)

		BeforeEach(func() {
			stackID = "stack-1"
			releaseID = "rel-1"
			preview = &models.PreviewStack{
				ID:              "p-4",
				StackID:         &stackID,
				ActiveReleaseID: &releaseID,
				CommitSHA:       "abc1234",
				Status:          models.PreviewStackStatus{Phase: models.PreviewStackPhaseDeploying},
			}
		})

		It("requeues when release is Superseded", func() {
			release := &models.StackRelease{
				ID:    releaseID,
				State: models.ReleaseStateSuperseded,
			}

			releaseSvc.EXPECT().InternalGet(gomock.Any(), releaseID).Return(release, nil)

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultRequeueAfter(convergePollInterval)))
		})

		It("requeues when release is InProgress", func() {
			release := &models.StackRelease{
				ID:    releaseID,
				State: models.ReleaseStateInProgress,
			}

			releaseSvc.EXPECT().InternalGet(gomock.Any(), releaseID).Return(release, nil)

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultRequeueAfter(convergePollInterval)))
		})

		It("requeues when release is Pending", func() {
			release := &models.StackRelease{
				ID:    releaseID,
				State: models.ReleaseStatePending,
			}

			releaseSvc.EXPECT().InternalGet(gomock.Any(), releaseID).Return(release, nil)

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultRequeueAfter(convergePollInterval)))
		})
	})

	Context("error cases", func() {
		It("returns error when release service fails", func() {
			releaseID := "rel-err"
			preview := &models.PreviewStack{
				ID:              "p-5",
				ActiveReleaseID: &releaseID,
				Status:          models.PreviewStackStatus{Phase: models.PreviewStackPhaseDeploying},
			}

			releaseSvc.EXPECT().InternalGet(gomock.Any(), releaseID).
				Return(nil, errors.GeneralError("database error"))

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get release"))
			Expect(result).To(Equal(resultNil))
		})

		It("returns error when stack service fails on Released path", func() {
			stackID := "stack-err"
			releaseID := "rel-released"
			preview := &models.PreviewStack{
				ID:              "p-6",
				StackID:         &stackID,
				ActiveReleaseID: &releaseID,
				CommitSHA:       "abc1234",
				Status:          models.PreviewStackStatus{Phase: models.PreviewStackPhaseDeploying},
			}
			release := &models.StackRelease{
				ID:    releaseID,
				State: models.ReleaseStateReleased,
			}

			releaseSvc.EXPECT().InternalGet(gomock.Any(), releaseID).Return(release, nil)
			stackSvc.EXPECT().InternalGetStack(gomock.Any(), stackID).
				Return(nil, errors.GeneralError("stack fetch error"))

			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get stack"))
			Expect(result).To(Equal(resultNil))
		})
	})

	It("returns Name() as converge", func() {
		Expect(reconciler.Name()).To(Equal("converge"))
	})
})
