package release

import (
	"context"
	stderrors "errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
)

func testLogger() logger.Logger {
	return logger.NewLoggerWithPrefix(context.Background(), "release-test")
}

var _ = Describe("GatekeeperReconciler", func() {
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	It("marks an older sequence superseded and stops", func() {
		svc := NewMockreleaseService(ctrl)
		svc.EXPECT().InternalGetLatestByStackID(gomock.Any(), "stack-1").
			Return(&models.StackRelease{ID: "newer", Sequence: 2}, nil)
		svc.EXPECT().MarkSuperseded(gomock.Any(), "older", "Release superseded by release #2").
			Return(true, nil)

		rec := NewMockeventRecorder(ctrl)

		r := &gatekeeperReconciler{releaseService: svc, eventRecorder: rec, logger: testLogger()}

		release := &models.StackRelease{ID: "older", StackID: "stack-1", Sequence: 1, State: models.ReleaseStateInProgress}
		result, err := r.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultStop).To(BeTrue())
	})

	It("continues when the release is the latest sequence", func() {
		release := &models.StackRelease{ID: "self", StackID: "stack-1", Sequence: 2, State: models.ReleaseStateInProgress}

		svc := NewMockreleaseService(ctrl)
		svc.EXPECT().InternalGetLatestByStackID(gomock.Any(), "stack-1").
			Return(release, nil)

		rec := NewMockeventRecorder(ctrl)

		r := &gatekeeperReconciler{releaseService: svc, eventRecorder: rec, logger: testLogger()}

		result, err := r.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultNil).To(BeTrue())
	})

	It("marks a pending release in progress when the CAS wins", func() {
		release := &models.StackRelease{ID: "self", StackID: "stack-1", Sequence: 1, State: models.ReleaseStatePending}

		svc := NewMockreleaseService(ctrl)
		svc.EXPECT().InternalGetLatestByStackID(gomock.Any(), "stack-1").
			Return(release, nil)
		svc.EXPECT().MarkInProgress(gomock.Any(), "self").
			Return(true, nil)

		rec := NewMockeventRecorder(ctrl)
		rec.EXPECT().RecordReleaseStarted(gomock.Any(), release).Return(nil)

		r := &gatekeeperReconciler{releaseService: svc, eventRecorder: rec, logger: testLogger()}

		result, err := r.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultNil).To(BeTrue())
		Expect(release.State).To(Equal(models.ReleaseStateInProgress))
	})

	It("treats a recorder error on release start as log-only", func() {
		release := &models.StackRelease{ID: "self", StackID: "stack-1", Sequence: 1, State: models.ReleaseStatePending}

		svc := NewMockreleaseService(ctrl)
		svc.EXPECT().InternalGetLatestByStackID(gomock.Any(), "stack-1").
			Return(release, nil)
		svc.EXPECT().MarkInProgress(gomock.Any(), "self").
			Return(true, nil)

		rec := NewMockeventRecorder(ctrl)
		rec.EXPECT().RecordReleaseStarted(gomock.Any(), release).
			Return(errors.GeneralError("event insert failed"))

		r := &gatekeeperReconciler{releaseService: svc, eventRecorder: rec, logger: testLogger()}

		result, err := r.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred(), "recorder error must not fail the reconciler")
		Expect(result.resultNil).To(BeTrue())
	})

	It("stops without recording when the pending CAS is lost", func() {
		release := &models.StackRelease{ID: "self", StackID: "stack-1", Sequence: 1, State: models.ReleaseStatePending}

		svc := NewMockreleaseService(ctrl)
		svc.EXPECT().InternalGetLatestByStackID(gomock.Any(), "stack-1").
			Return(release, nil)
		svc.EXPECT().MarkInProgress(gomock.Any(), "self").
			Return(false, nil)

		// CAS lost: recorder must never be called.
		rec := NewMockeventRecorder(ctrl)

		r := &gatekeeperReconciler{releaseService: svc, eventRecorder: rec, logger: testLogger()}

		result, err := r.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultStop).To(BeTrue())
	})
})

var _ = Describe("RenderReconciler", func() {
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	It("continues when a manifest is already present", func() {
		r := &renderReconciler{releaseService: NewMockreleaseService(ctrl)}
		release := &models.StackRelease{Manifest: &models.ReleaseManifest{}}

		result, err := r.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultNil).To(BeTrue())
	})
})

var _ = Describe("ConvergeReconciler", func() {
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	It("marks released and stops when the stack matches the last converged record", func() {
		now := time.Now().UTC()

		release := &models.StackRelease{
			ID:               "rel-1",
			StackID:          "stack-1",
			ManifestRevision: "rev-1",
			RenderedAt:       &now,
			Manifest:         &models.ReleaseManifest{},
		}

		relSvc := NewMockreleaseService(ctrl)
		relSvc.EXPECT().MarkReleased(gomock.Any(), "rel-1", gomock.Any()).
			Return(true, nil)

		stackSvc := NewMockstackService(ctrl)
		stackSvc.EXPECT().InternalGetStack(gomock.Any(), "stack-1").
			Return(&models.Stack{
				ID: "stack-1",
				Status: &models.StackStatus{
					LastConverged: &models.StackConvergenceRecord{
						ReleaseID: "rel-1",
						Revision:  "rev-1",
					},
				},
			}, nil)

		r := &convergeReconciler{
			releaseService: relSvc,
			stackService:   stackSvc,
			logger:         testLogger(),
		}

		result, err := r.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultStop).To(BeTrue())
	})

	It("stops when the MarkReleased CAS fails", func() {
		now := time.Now().UTC()

		release := &models.StackRelease{
			ID:               "rel-1",
			StackID:          "stack-1",
			ManifestRevision: "rev-1",
			RenderedAt:       &now,
			Manifest:         &models.ReleaseManifest{},
		}

		relSvc := NewMockreleaseService(ctrl)
		relSvc.EXPECT().MarkReleased(gomock.Any(), "rel-1", gomock.Any()).
			Return(false, nil)

		stackSvc := NewMockstackService(ctrl)
		stackSvc.EXPECT().InternalGetStack(gomock.Any(), "stack-1").
			Return(&models.Stack{
				ID: "stack-1",
				Status: &models.StackStatus{
					LastConverged: &models.StackConvergenceRecord{
						ReleaseID: "rel-1",
						Revision:  "rev-1",
					},
				},
			}, nil)

		r := &convergeReconciler{
			releaseService: relSvc,
			stackService:   stackSvc,
			logger:         testLogger(),
		}

		result, err := r.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultStop).To(BeTrue(), "expected resultStop when CAS fails")
	})

	// There is no deploy timeout: a slow build or convergence never fails the
	// release on time alone. It keeps polling however old the render is.
	It("keeps polling a long-unconverged release instead of failing it", func() {
		past := time.Now().UTC().Add(-24 * time.Hour)

		release := &models.StackRelease{
			ID:               "rel-1",
			StackID:          "stack-1",
			ManifestRevision: "rev-1",
			RenderedAt:       &past,
			Manifest:         &models.ReleaseManifest{},
		}

		stackSvc := NewMockstackService(ctrl)
		stackSvc.EXPECT().InternalGetStack(gomock.Any(), "stack-1").
			Return(&models.Stack{ID: "stack-1"}, nil)

		buildSvc := NewMockimageBuildService(ctrl)
		buildSvc.EXPECT().ListByStackID(gomock.Any(), "stack-1").Return(nil, nil)

		r := &convergeReconciler{
			// No expectations: any MarkFailed call fails the test.
			releaseService:    NewMockreleaseService(ctrl),
			stackService:      stackSvc,
			imageBuildService: buildSvc,
			logger:            testLogger(),
		}

		result, err := r.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultRequeueAfter).ToNot(BeNil())
		Expect(*result.resultRequeueAfter).To(Equal(convergencePollInterval))
	})

	// The fail-fast that replaced the deploy timeout: a terminally failed build
	// this release triggered can never converge.
	It("marks failed and stops when this release's build terminally fails", func() {
		release := &models.StackRelease{
			ID:               "rel-1",
			StackID:          "stack-1",
			ManifestRevision: "rev-1",
			Manifest:         &models.ReleaseManifest{},
		}

		stackSvc := NewMockstackService(ctrl)
		stackSvc.EXPECT().InternalGetStack(gomock.Any(), "stack-1").
			Return(&models.Stack{ID: "stack-1"}, nil)

		buildSvc := NewMockimageBuildService(ctrl)
		buildSvc.EXPECT().ListByStackID(gomock.Any(), "stack-1").Return([]*models.ImageBuild{{
			StackResourceName: "web",
			Status: &models.ImageBuildStatus{
				State:                  string(buildsv1alpha1.BuildPhaseFailed),
				ReleaseID:              "rel-1",
				LastBuildFailureDetail: &models.BuildFailureDetail{Message: "COPY failed: file not found"},
			},
		}}, nil)

		relSvc := NewMockreleaseService(ctrl)
		relSvc.EXPECT().
			MarkFailed(gomock.Any(), "rel-1", "build failed for web: COPY failed: file not found", gomock.Any()).
			Return(true, nil)

		r := &convergeReconciler{
			releaseService:    relSvc,
			stackService:      stackSvc,
			imageBuildService: buildSvc,
			logger:            testLogger(),
		}

		result, err := r.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultStop).To(BeTrue())
	})

	It("marks failed and stops when this release's build is cancelled", func() {
		release := &models.StackRelease{
			ID:               "rel-1",
			StackID:          "stack-1",
			ManifestRevision: "rev-1",
			Manifest:         &models.ReleaseManifest{},
		}

		stackSvc := NewMockstackService(ctrl)
		stackSvc.EXPECT().InternalGetStack(gomock.Any(), "stack-1").
			Return(&models.Stack{ID: "stack-1"}, nil)

		buildSvc := NewMockimageBuildService(ctrl)
		buildSvc.EXPECT().ListByStackID(gomock.Any(), "stack-1").Return([]*models.ImageBuild{{
			StackResourceName: "web",
			Status: &models.ImageBuildStatus{
				State:     string(buildsv1alpha1.BuildPhaseCancelled),
				ReleaseID: "rel-1",
			},
		}}, nil)

		relSvc := NewMockreleaseService(ctrl)
		relSvc.EXPECT().
			MarkFailed(gomock.Any(), "rel-1", "build cancelled for web", gomock.Any()).
			Return(true, nil)

		r := &convergeReconciler{
			releaseService:    relSvc,
			stackService:      stackSvc,
			imageBuildService: buildSvc,
			logger:            testLogger(),
		}

		result, err := r.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultStop).To(BeTrue())
	})

	// Another release's failed build — or one from an agent that predates the
	// release-id annotation — says nothing about this release: keep polling.
	DescribeTable("keeps polling on a failed build that is not this release's",
		func(build *models.ImageBuild) {
			release := &models.StackRelease{
				ID:               "rel-2",
				StackID:          "stack-1",
				ManifestRevision: "rev-2",
				Manifest:         &models.ReleaseManifest{},
			}

			stackSvc := NewMockstackService(ctrl)
			stackSvc.EXPECT().InternalGetStack(gomock.Any(), "stack-1").
				Return(&models.Stack{ID: "stack-1"}, nil)

			buildSvc := NewMockimageBuildService(ctrl)
			buildSvc.EXPECT().ListByStackID(gomock.Any(), "stack-1").
				Return([]*models.ImageBuild{build}, nil)

			r := &convergeReconciler{
				releaseService:    NewMockreleaseService(ctrl),
				stackService:      stackSvc,
				imageBuildService: buildSvc,
				logger:            testLogger(),
			}

			result, err := r.Reconcile(context.Background(), release)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.resultRequeueAfter).ToNot(BeNil())
		},
		Entry("prior release's failure", &models.ImageBuild{
			StackResourceName: "web",
			Status: &models.ImageBuildStatus{
				State:     string(buildsv1alpha1.BuildPhaseFailed),
				ReleaseID: "rel-1",
			},
		}),
		Entry("no release id on the build", &models.ImageBuild{
			StackResourceName: "web",
			Status:            &models.ImageBuildStatus{State: string(buildsv1alpha1.BuildPhaseFailed)},
		}),
	)

	It("requeues at the convergence poll interval when not yet converged", func() {
		now := time.Now().UTC()

		release := &models.StackRelease{
			ID:               "rel-1",
			StackID:          "stack-1",
			ManifestRevision: "rev-1",
			RenderedAt:       &now,
			Manifest:         &models.ReleaseManifest{},
		}

		stackSvc := NewMockstackService(ctrl)
		stackSvc.EXPECT().InternalGetStack(gomock.Any(), "stack-1").
			Return(&models.Stack{ID: "stack-1"}, nil)

		buildSvc := NewMockimageBuildService(ctrl)
		buildSvc.EXPECT().ListByStackID(gomock.Any(), "stack-1").Return(nil, nil)

		r := &convergeReconciler{
			releaseService:    NewMockreleaseService(ctrl),
			stackService:      stackSvc,
			imageBuildService: buildSvc,
			logger:            testLogger(),
		}

		result, err := r.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultRequeueAfter).ToNot(BeNil())
		Expect(*result.resultRequeueAfter).To(Equal(convergencePollInterval))
	})
})

var _ = Describe("ApplyReconciler", func() {
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	// The apply-time cluster-CR supersede path mirrors the gatekeeper supersede
	// invariant: whether the CAS is won or lost, supersededByClusterCR reports
	// the release as superseded.
	DescribeTable("supersededByClusterCR",
		func(casWon bool) {
			relSvc := NewMockreleaseService(ctrl)
			relSvc.EXPECT().InternalGet(gomock.Any(), "higher-release").
				Return(&models.StackRelease{ID: "higher-release", Sequence: 5}, nil)
			relSvc.EXPECT().MarkSuperseded(gomock.Any(), "self", "Release superseded by release #5").
				Return(casWon, nil)

			r := &applyReconciler{
				releaseService: relSvc,
				logger:         testLogger(),
			}

			existing := &corev1alpha1.Stack{}
			existing.SetAnnotations(map[string]string{
				corev1alpha1.ReleaseIDAnnotation: "higher-release",
			})

			release := &models.StackRelease{ID: "self", Sequence: 3}

			superseded, err := r.supersededByClusterCR(context.Background(), existing, release)
			Expect(err).ToNot(HaveOccurred())
			Expect(superseded).To(BeTrue(), "expected release to be superseded")
		},
		Entry("CAS win", true),
		Entry("CAS loss", false),
	)

	It("does not supersede when the cluster CR carries a lower sequence", func() {
		relSvc := NewMockreleaseService(ctrl)
		relSvc.EXPECT().InternalGet(gomock.Any(), "lower-release").
			Return(&models.StackRelease{ID: "lower-release", Sequence: 1}, nil)

		r := &applyReconciler{
			releaseService: relSvc,
			logger:         testLogger(),
		}

		existing := &corev1alpha1.Stack{}
		existing.SetAnnotations(map[string]string{
			corev1alpha1.ReleaseIDAnnotation: "lower-release",
		})

		release := &models.StackRelease{ID: "self", Sequence: 3}

		superseded, err := r.supersededByClusterCR(context.Background(), existing, release)
		Expect(err).ToNot(HaveOccurred())
		Expect(superseded).To(BeFalse(), "expected release NOT to be superseded by lower sequence")
	})

	It("continues when the manifest is nil", func() {
		r := &applyReconciler{logger: testLogger()}

		release := &models.StackRelease{ID: "self", Manifest: nil}
		result, err := r.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultNil).To(BeTrue(), "expected resultNil when manifest is nil")
	})

	It("marks failed and stops when the stack is being deleted", func() {
		now := time.Now().UTC()

		relSvc := NewMockreleaseService(ctrl)
		relSvc.EXPECT().MarkFailed(gomock.Any(), "self", gomock.Any(), gomock.Any()).
			Return(true, nil)

		stackSvc := NewMockstackService(ctrl)
		stackSvc.EXPECT().InternalGetStack(gomock.Any(), "stack-1").
			Return(&models.Stack{
				ID:                "stack-1",
				DeletionTimestamp: &now,
			}, nil)

		r := &applyReconciler{
			releaseService: relSvc,
			stackService:   stackSvc,
			logger:         testLogger(),
		}

		release := &models.StackRelease{
			ID:       "self",
			StackID:  "stack-1",
			Manifest: &models.ReleaseManifest{},
		}
		result, err := r.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultStop).To(BeTrue(), "expected resultStop when stack is being deleted")
	})

	It("rejects a cluster mutation after a newer terminal release is persisted", func() {
		release := &models.StackRelease{ID: "release-a", StackID: "stack-1", Sequence: 1, State: models.ReleaseStateInProgress}
		stack := &models.Stack{ID: "stack-1", OrganisationID: "org-1"}
		newer := &models.StackRelease{ID: "release-b", StackID: "stack-1", Sequence: 2, State: models.ReleaseStateFailed}
		relSvc := NewMockreleaseService(ctrl)
		relSvc.EXPECT().InternalGetLatestByStackID(gomock.Any(), release.StackID).Return(newer, nil)
		relSvc.EXPECT().MarkSuperseded(gomock.Any(), release.ID, "Release superseded by release #2").Return(true, nil)
		stackSvc := NewMockstackService(ctrl)
		stackSvc.EXPECT().InternalGetStack(gomock.Any(), stack.ID).Return(stack, nil)
		r := &applyReconciler{releaseService: relSvc, stackService: stackSvc}

		err := r.authorizeMutation(release)(context.Background())

		Expect(stderrors.Is(err, errReleaseSuperseded)).To(BeTrue())
	})

	It("rejects a cluster mutation when stack deletion starts after apply authorization", func() {
		now := time.Now().UTC()
		release := &models.StackRelease{ID: "release-a", StackID: "stack-1", State: models.ReleaseStateInProgress}
		stackSvc := NewMockstackService(ctrl)
		stackSvc.EXPECT().InternalGetStack(gomock.Any(), release.StackID).Return(&models.Stack{
			ID: release.StackID, DeletionTimestamp: &now,
		}, nil)
		r := &applyReconciler{stackService: stackSvc}

		err := r.authorizeMutation(release)(context.Background())

		Expect(stderrors.Is(err, errReleaseSuperseded)).To(BeTrue())
	})

	It("continues a cluster mutation after its release was admitted", func() {
		stack := &models.Stack{ID: "stack-1", OrganisationID: "org-1"}
		release := &models.StackRelease{
			ID: "release-a", StackID: stack.ID, State: models.ReleaseStateInProgress,
			Snapshot: models.StackSnapshot{Stack: models.StackShellSnapshot{ID: stack.ID, OrganisationID: stack.OrganisationID}},
		}
		stackSvc := NewMockstackService(ctrl)
		stackSvc.EXPECT().InternalGetStack(gomock.Any(), stack.ID).Return(stack, nil)
		relSvc := NewMockreleaseService(ctrl)
		relSvc.EXPECT().InternalGetLatestByStackID(gomock.Any(), stack.ID).Return(release, nil)
		r := &applyReconciler{releaseService: relSvc, stackService: stackSvc}

		err := r.authorizeMutation(release)(context.Background())

		Expect(err).NotTo(HaveOccurred())
	})
})
