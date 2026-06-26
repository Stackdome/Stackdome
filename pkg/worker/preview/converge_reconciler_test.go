package preview

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("ConvergeReconciler", func() {
	var (
		ctrl         *gomock.Controller
		previewStore *MockpreviewStackStore
		stackSvc     *MockstackService
		releaseSvc   *MockreleaseService
		reconciler   *convergeReconciler
		ctx          context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		previewStore = NewMockpreviewStackStore(ctrl)
		stackSvc = NewMockstackService(ctrl)
		releaseSvc = NewMockreleaseService(ctrl)
		ctx = context.Background()

		reconciler = &convergeReconciler{
			releaseService:    releaseSvc,
			stackService:      stackSvc,
			previewStackStore: previewStore,
			logger:            logger.NewLoggerWithPrefix(ctx, "test"),
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("noop cases", func() {
		It("returns resultNil when phase is not Deploying", func() {
			preview := &models.PreviewStack{
				ID:     "p-1",
				Status: models.PreviewStackStatus{Phase: models.PreviewStackPhaseReady},
			}
			result, err := reconciler.Reconcile(ctx, preview)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultNil))
		})

		It("returns resultStop when ActiveReleaseID is nil", func() {
			preview := &models.PreviewStack{
				ID:     "p-2",
				Status: models.PreviewStackStatus{Phase: models.PreviewStackPhaseDeploying},
			}
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
			previewStore.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, p *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
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
