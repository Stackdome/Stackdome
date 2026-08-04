package imagebuild

import (
	"context"
	"testing"

	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	ctrlreconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

func TestMapClusterStatusToServerStatus_noFailure(t *testing.T) {
	status := buildsv1alpha1.ImageBuildStatus{
		Phase:      buildsv1alpha1.BuildPhaseSuccess,
		StatusHash: "abc123",
		ImageUrl:   "registry.io/img:tag",
	}

	got := mapClusterStatusToServerStatus(status)

	if got.LastBuildFailureDetail != nil {
		t.Errorf("expected LastBuildFailureDetail to be nil, got %v", got.LastBuildFailureDetail)
	}
	if got.ImageURL != "registry.io/img:tag" {
		t.Errorf("ImageURL = %q", got.ImageURL)
	}
}

func TestMapClusterStatusToServerStatus_withFailure(t *testing.T) {
	status := buildsv1alpha1.ImageBuildStatus{
		Phase:      buildsv1alpha1.BuildPhaseFailed,
		StatusHash: "def456",
		LastBuildFailureDetail: &corev1alpha1.LastFailureDetail{
			ContainerName:           "kaniko",
			RestartCount:            0,
			LastTerminationReason:   "Error",
			LastTerminationMessage:  "COPY failed: file not found",
			LastTerminationExitCode: ptr.To(int32(1)),
		},
	}

	got := mapClusterStatusToServerStatus(status)

	if got.LastBuildFailureDetail == nil {
		t.Fatal("expected LastBuildFailureDetail to be set")
	}
	if got.LastBuildFailureDetail.FailureType != "exit_error" {
		t.Errorf("FailureType = %q, want exit_error", got.LastBuildFailureDetail.FailureType)
	}
	if got.LastBuildFailureDetail.Message != "COPY failed: file not found" {
		t.Errorf("Message = %q", got.LastBuildFailureDetail.Message)
	}
	if got.LastBuildFailureDetail.ExitCode == nil || *got.LastBuildFailureDetail.ExitCode != 1 {
		t.Errorf("ExitCode = %v, want 1", got.LastBuildFailureDetail.ExitCode)
	}
}

func TestBuildStackResourceFailureFromBuild_failure(t *testing.T) {
	detail := &corev1alpha1.LastFailureDetail{
		ContainerName:           "kaniko",
		LastTerminationReason:   "OOMKilled",
		LastTerminationExitCode: ptr.To(int32(137)),
	}

	got := buildStackResourceFailureFromBuild(detail)

	if got == nil {
		t.Fatal("expected non-nil failure")
	}
	if got.Type != models.FailureTypeBuildFailure {
		t.Errorf("Type = %q, want build_failure", got.Type)
	}
	if got.Build == nil {
		t.Fatal("expected Build to be set")
	}
	if got.Build.FailureType != "out_of_memory" {
		t.Errorf("Build.FailureType = %q, want out_of_memory", got.Build.FailureType)
	}
}

func TestBuildStackResourceFailureFromBuild_nil(t *testing.T) {
	got := buildStackResourceFailureFromBuild(nil)
	if got != nil {
		t.Errorf("expected nil for nil input")
	}
}

func TestComputeStackResourceStatusAfterBuild_buildFailure(t *testing.T) {
	current := models.StackResourceStatus{
		State: models.StackResourcePhaseReady,
	}
	buildStatus := buildsv1alpha1.ImageBuildStatus{
		Phase: buildsv1alpha1.BuildPhaseFailed,
		LastBuildFailureDetail: &corev1alpha1.LastFailureDetail{
			LastTerminationReason:   "Error",
			LastTerminationMessage:  "COPY failed",
			LastTerminationExitCode: ptr.To(int32(1)),
		},
	}

	got, changed := computeStackResourceStatusAfterBuild(current, buildStatus)

	if !changed {
		t.Fatal("expected changed=true for build failure")
	}
	if got == nil {
		t.Fatal("expected non-nil status")
	}
	if got.LastFailure == nil {
		t.Fatal("expected LastFailure to be set")
	}
	if got.LastFailure.Type != models.FailureTypeBuildFailure {
		t.Errorf("Type = %q, want build_failure", got.LastFailure.Type)
	}
	if got.LastFailure.Build == nil || got.LastFailure.Build.FailureType != "exit_error" {
		t.Errorf("Build.FailureType = %v, want exit_error", got.LastFailure.Build)
	}
}

func TestComputeStackResourceStatusAfterBuild_successClearsBuildFailure(t *testing.T) {
	current := models.StackResourceStatus{
		LastFailure: &models.StackResourceFailure{
			Type: models.FailureTypeBuildFailure,
		},
	}
	buildStatus := buildsv1alpha1.ImageBuildStatus{
		Phase: buildsv1alpha1.BuildPhaseSuccess,
	}

	got, changed := computeStackResourceStatusAfterBuild(current, buildStatus)

	if !changed {
		t.Fatal("expected changed=true when clearing build failure on success")
	}
	if got.LastFailure != nil {
		t.Errorf("expected LastFailure to be nil after build success, got %v", got.LastFailure)
	}
}

func TestComputeStackResourceStatusAfterBuild_successNoOpWhenNoBuildFailure(t *testing.T) {
	current := models.StackResourceStatus{
		LastFailure: &models.StackResourceFailure{
			Type: models.FailureTypeRuntimeCrash,
		},
	}
	buildStatus := buildsv1alpha1.ImageBuildStatus{
		Phase: buildsv1alpha1.BuildPhaseSuccess,
	}

	got, changed := computeStackResourceStatusAfterBuild(current, buildStatus)

	if changed {
		t.Errorf("expected changed=false when success but LastFailure is runtime_crash, got status=%v", got)
	}
}

func TestComputeStackResourceStatusAfterBuild_inProgressNoOp(t *testing.T) {
	current := models.StackResourceStatus{}
	buildStatus := buildsv1alpha1.ImageBuildStatus{
		Phase: buildsv1alpha1.BuildPhasePending,
	}

	_, changed := computeStackResourceStatusAfterBuild(current, buildStatus)

	if changed {
		t.Error("expected changed=false for in-progress build")
	}
}

func TestPropagateBuildFailure_nilStatusWithFailureReturnsError(t *testing.T) {
	r := &ImageBuildReconciler{}
	resource := &models.StackResource{ID: "res-1"}
	buildStatus := buildsv1alpha1.ImageBuildStatus{
		Phase: buildsv1alpha1.BuildPhaseFailed,
		LastBuildFailureDetail: &corev1alpha1.LastFailureDetail{
			LastTerminationReason:   "Error",
			LastTerminationExitCode: ptr.To(int32(1)),
		},
	}

	err := r.propagateBuildFailureToStackResource(context.Background(), resource, buildStatus)

	if err == nil {
		t.Fatal("expected an error so the reconcile requeues instead of dropping the build failure")
	}
}

func TestPropagateBuildFailure_nilStatusWithoutFailureIsNoOp(t *testing.T) {
	r := &ImageBuildReconciler{}
	resource := &models.StackResource{ID: "res-1"}
	buildStatus := buildsv1alpha1.ImageBuildStatus{
		Phase: buildsv1alpha1.BuildPhaseSuccess,
	}

	err := r.propagateBuildFailureToStackResource(context.Background(), resource, buildStatus)

	if err != nil {
		t.Fatalf("expected no error when there is no failure to propagate, got %v", err)
	}
}

const (
	testEventStackID        = "stack-1"
	testEventResourceName   = "api"
	testEventBuildID        = "build-1"
	testEventNamespace      = "ns-1"
	testEventClusterID      = "cluster-1"
	testEventOtherClusterID = "cluster-2"
)

func newEventReconciler() (*ImageBuildReconciler, *MockreleaseActiveChecker, *MockbuildEventRecorder) {
	ctrl := gomock.NewController(GinkgoT())
	checker := NewMockreleaseActiveChecker(ctrl)
	recorder := NewMockbuildEventRecorder(ctrl)
	r := &ImageBuildReconciler{
		Logger:         logger.NewLoggerWithPrefix(context.Background(), "event-test"),
		releaseChecker: checker,
		eventRecorder:  recorder,
	}
	return r, checker, recorder
}

func eventBuildCR(phase buildsv1alpha1.BuildPhase) *buildsv1alpha1.ImageBuild {
	return &buildsv1alpha1.ImageBuild{
		ObjectMeta: metav1.ObjectMeta{Name: testEventBuildID, Namespace: testEventNamespace},
		Spec:       buildsv1alpha1.ImageBuildSpec{ResourceName: testEventResourceName},
		Status:     buildsv1alpha1.ImageBuildStatus{Phase: phase},
	}
}

func gitBuildModel(commit string) *models.ImageBuild {
	return &models.ImageBuild{
		ID: testEventBuildID,
		Spec: models.BuildConfigSpec{
			SourceRevision: models.BuildSourceRevision{
				Git: &models.GitRevision{Commit: commit},
			},
		},
	}
}

var _ = Describe("recordBuildEvent", func() {
	DescribeTable("phase to event type mapping with best-effort active-release attribution",
		func(phase buildsv1alpha1.BuildPhase, eventType models.ReleaseEventType) {
			r, checker, recorder := newEventReconciler()
			active := &models.StackRelease{ID: "rel-1"}

			checker.EXPECT().
				InternalGetActiveByStackID(gomock.Any(), testEventStackID).
				Return(active, nil)
			recorder.EXPECT().
				RecordBuildEvent(
					gomock.Any(),
					active,
					testEventResourceName,
					eventType,
					testEventBuildID,
					models.ReleaseEventAttributionActiveRelease,
					nil,
				).
				Return(nil)

			r.recordBuildEvent(context.Background(), testEventStackID, eventBuildCR(phase), gitBuildModel("deadbeef"))
		},
		Entry("pending emits build_started", buildsv1alpha1.BuildPhasePending, models.ReleaseEventTypeBuildStarted),
		Entry("success emits build_succeeded", buildsv1alpha1.BuildPhaseSuccess, models.ReleaseEventTypeBuildSucceeded),
		Entry("failed emits build_failed", buildsv1alpha1.BuildPhaseFailed, models.ReleaseEventTypeBuildFailed),
	)

	It("includes the mapped failure detail on failed builds", func() {
		r, checker, recorder := newEventReconciler()
		active := &models.StackRelease{ID: "rel-1"}
		cr := eventBuildCR(buildsv1alpha1.BuildPhaseFailed)
		cr.Status.LastBuildFailureDetail = &corev1alpha1.LastFailureDetail{
			LastTerminationReason:   "Error",
			LastTerminationMessage:  "COPY failed: file not found",
			LastTerminationExitCode: ptr.To(int32(1)),
		}

		checker.EXPECT().
			InternalGetActiveByStackID(gomock.Any(), testEventStackID).
			Return(active, nil)
		recorder.EXPECT().
			RecordBuildEvent(
				gomock.Any(),
				active,
				testEventResourceName,
				models.ReleaseEventTypeBuildFailed,
				testEventBuildID,
				models.ReleaseEventAttributionActiveRelease,
				&models.BuildFailureDetail{
					FailureType: "exit_error",
					Reason:      "Error",
					Message:     "COPY failed: file not found",
					ExitCode:    ptr.To(int32(1)),
				},
			).
			Return(nil)

		r.recordBuildEvent(context.Background(), testEventStackID, cr, gitBuildModel("deadbeef"))
	})

	It("records build_attempt_failed when a pending build carries a failure detail", func() {
		r, checker, recorder := newEventReconciler()
		active := &models.StackRelease{ID: "rel-1"}
		cr := eventBuildCR(buildsv1alpha1.BuildPhasePending)
		cr.Status.LastBuildFailureDetail = &corev1alpha1.LastFailureDetail{
			LastTerminationReason:   "OOMKilled",
			LastTerminationMessage:  "build container OOMKilled",
			RestartCount:            2,
			LastTerminationExitCode: ptr.To(int32(137)),
		}

		checker.EXPECT().
			InternalGetActiveByStackID(gomock.Any(), testEventStackID).
			Return(active, nil)
		// build_attempt_failed only: gomock rejects a build_started call.
		recorder.EXPECT().
			RecordBuildEvent(
				gomock.Any(),
				active,
				testEventResourceName,
				models.ReleaseEventTypeBuildAttemptFailed,
				testEventBuildID,
				models.ReleaseEventAttributionActiveRelease,
				&models.BuildFailureDetail{
					FailureType:  "out_of_memory",
					Reason:       "OOMKilled",
					Message:      "build container OOMKilled",
					RestartCount: 2,
					ExitCode:     ptr.To(int32(137)),
				},
			).
			Return(nil)

		r.recordBuildEvent(context.Background(), testEventStackID, cr, gitBuildModel("deadbeef"))
	})

	DescribeTable("phases that emit nothing",
		func(phase buildsv1alpha1.BuildPhase) {
			// No checker/recorder EXPECT: gomock fails the spec on any call.
			r, _, _ := newEventReconciler()
			r.recordBuildEvent(context.Background(), testEventStackID, eventBuildCR(phase), gitBuildModel("deadbeef"))
		},
		Entry("cancelled", buildsv1alpha1.BuildPhaseCancelled),
		Entry("unknown phase", buildsv1alpha1.BuildPhase("Weird")),
	)

	DescribeTable("skips the recorder when there is no active release",
		func(active *models.StackRelease, serr *apperrors.ServiceError) {
			r, checker, _ := newEventReconciler()
			checker.EXPECT().
				InternalGetActiveByStackID(gomock.Any(), testEventStackID).
				Return(active, serr)
			// No recorder EXPECT: it must not be called.
			r.recordBuildEvent(context.Background(), testEventStackID, eventBuildCR(buildsv1alpha1.BuildPhasePending), gitBuildModel("deadbeef"))
		},
		Entry("nil active release", nil, nil),
		Entry("lookup error", nil, apperrors.GeneralError("boom")),
	)

	DescribeTable("deterministic pin match drops the attribution marker",
		func(pins models.ReleasePins, build *models.ImageBuild) {
			r, checker, recorder := newEventReconciler()
			active := &models.StackRelease{ID: "rel-1", Pins: pins}

			checker.EXPECT().
				InternalGetActiveByStackID(gomock.Any(), testEventStackID).
				Return(active, nil)
			recorder.EXPECT().
				RecordBuildEvent(
					gomock.Any(),
					active,
					testEventResourceName,
					models.ReleaseEventTypeBuildStarted,
					testEventBuildID,
					"", // deterministic pin match: no attribution marker
					nil,
				).
				Return(nil)

			r.recordBuildEvent(context.Background(), testEventStackID, eventBuildCR(buildsv1alpha1.BuildPhasePending), build)
		},
		Entry("git commit pin match",
			models.ReleasePins{Resources: map[string]models.ResourcePins{
				testEventResourceName: {GitSHA: "abc123"},
			}},
			gitBuildModel("abc123"),
		),
		Entry("volume hash pin match",
			models.ReleasePins{Resources: map[string]models.ResourcePins{
				testEventResourceName: {VolumeHash: "vh-9"},
			}},
			&models.ImageBuild{
				ID: testEventBuildID,
				Spec: models.BuildConfigSpec{
					SourceRevision: models.BuildSourceRevision{
						Volume: &models.VolumeRevision{CurrentVolumeHash: "vh-9"},
					},
				},
			},
		),
	)

	It("falls back to active-release attribution on a pin mismatch", func() {
		r, checker, recorder := newEventReconciler()
		active := &models.StackRelease{
			ID: "rel-1",
			Pins: models.ReleasePins{Resources: map[string]models.ResourcePins{
				testEventResourceName: {GitSHA: "abc123"},
			}},
		}

		checker.EXPECT().
			InternalGetActiveByStackID(gomock.Any(), testEventStackID).
			Return(active, nil)
		recorder.EXPECT().
			RecordBuildEvent(
				gomock.Any(), active, testEventResourceName,
				models.ReleaseEventTypeBuildStarted, testEventBuildID,
				models.ReleaseEventAttributionActiveRelease, nil,
			).
			Return(nil)

		r.recordBuildEvent(context.Background(), testEventStackID, eventBuildCR(buildsv1alpha1.BuildPhasePending), gitBuildModel("differentsha"))
	})

	// A terminal build failure whose pins deterministically belong to the active
	// release fails that release: it can never converge, and there is no deploy
	// timeout to catch it.
	It("fails the release when a pin-matched build terminally fails", func() {
		r, checker, recorder := newEventReconciler()
		active := &models.StackRelease{
			ID: "rel-1",
			Pins: models.ReleasePins{Resources: map[string]models.ResourcePins{
				testEventResourceName: {GitSHA: "abc123"},
			}},
		}
		cr := eventBuildCR(buildsv1alpha1.BuildPhaseFailed)
		cr.Status.LastBuildFailureDetail = &corev1alpha1.LastFailureDetail{
			LastTerminationReason:  "Error",
			LastTerminationMessage: "COPY failed: file not found",
		}

		checker.EXPECT().
			InternalGetActiveByStackID(gomock.Any(), testEventStackID).
			Return(active, nil)
		recorder.EXPECT().
			RecordBuildEvent(
				gomock.Any(), active, testEventResourceName,
				models.ReleaseEventTypeBuildFailed, testEventBuildID,
				"", gomock.Any(),
			).
			Return(nil)
		checker.EXPECT().
			MarkFailed(gomock.Any(), "rel-1", "build failed for "+testEventResourceName+": COPY failed: file not found", nil).
			Return(true, nil)

		r.recordBuildEvent(context.Background(), testEventStackID, cr, gitBuildModel("abc123"))
	})

	// Best-effort attribution is not enough to fail a release: the specs above
	// with BuildPhaseFailed and no pin match set no MarkFailed expectation, so
	// gomock rejects any call. A pin-matched but non-terminal attempt failure
	// must not fail the release either.
	It("does not fail the release on a pin-matched build attempt failure", func() {
		r, checker, recorder := newEventReconciler()
		active := &models.StackRelease{
			ID: "rel-1",
			Pins: models.ReleasePins{Resources: map[string]models.ResourcePins{
				testEventResourceName: {GitSHA: "abc123"},
			}},
		}
		cr := eventBuildCR(buildsv1alpha1.BuildPhasePending)
		cr.Status.LastBuildFailureDetail = &corev1alpha1.LastFailureDetail{
			LastTerminationReason: "Error",
		}

		checker.EXPECT().
			InternalGetActiveByStackID(gomock.Any(), testEventStackID).
			Return(active, nil)
		recorder.EXPECT().
			RecordBuildEvent(
				gomock.Any(), active, testEventResourceName,
				models.ReleaseEventTypeBuildAttemptFailed, testEventBuildID,
				"", gomock.Any(),
			).
			Return(nil)
		// No MarkFailed expectation: gomock fails the spec on any call.

		r.recordBuildEvent(context.Background(), testEventStackID, cr, gitBuildModel("abc123"))
	})

	It("swallows recorder errors", func() {
		r, checker, recorder := newEventReconciler()
		active := &models.StackRelease{ID: "rel-1"}

		checker.EXPECT().
			InternalGetActiveByStackID(gomock.Any(), testEventStackID).
			Return(active, nil)
		recorder.EXPECT().
			RecordBuildEvent(gomock.Any(), active, testEventResourceName, gomock.Any(), testEventBuildID, gomock.Any(), gomock.Any()).
			Return(apperrors.GeneralError("db down"))

		// Must not panic or otherwise surface the error.
		r.recordBuildEvent(context.Background(), testEventStackID, eventBuildCR(buildsv1alpha1.BuildPhaseSuccess), gitBuildModel("deadbeef"))
	})
})

func eventsTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
	Expect(buildsv1alpha1.AddToScheme(scheme)).To(Succeed())
	return scheme
}

func reconcileCR(phase buildsv1alpha1.BuildPhase, statusHash string) *buildsv1alpha1.ImageBuild {
	cr := eventBuildCR(phase)
	cr.Labels = map[string]string{corev1alpha1.LabelStackID: testEventStackID}
	cr.Status.StatusHash = statusHash
	return cr
}

func newReconcileHarness(cr *buildsv1alpha1.ImageBuild) (
	*ImageBuildReconciler, *mocks.MockStackResourceService, *mocks.MockImageBuildService, *MockreleaseActiveChecker, *MockbuildEventRecorder, *MockstackClusterResolver,
) {
	ctrl := gomock.NewController(GinkgoT())

	resources := mocks.NewMockStackResourceService(ctrl)
	builds := mocks.NewMockImageBuildService(ctrl)
	checker := NewMockreleaseActiveChecker(ctrl)
	recorder := NewMockbuildEventRecorder(ctrl)
	resolver := NewMockstackClusterResolver(ctrl)

	fakeClient := fake.NewClientBuilder().WithScheme(eventsTestScheme()).WithObjects(cr).Build()

	r := &ImageBuildReconciler{
		Client:              fakeClient,
		ClusterID:           testEventClusterID,
		DBResourceService:   resources,
		DBImageBuildService: builds,
		Logger:              logger.NewLoggerWithPrefix(context.Background(), "reconcile-test"),
		releaseChecker:      checker,
		eventRecorder:       recorder,
		stackResolver:       resolver,
	}
	return r, resources, builds, checker, recorder, resolver
}

func expectOwnedStack(resolver *MockstackClusterResolver) {
	resolver.EXPECT().
		InternalGetStack(gomock.Any(), testEventStackID).
		Return(&models.Stack{ID: testEventStackID, ClusterID: testEventClusterID}, nil)
}

var _ = Describe("Reconcile build event hook", func() {
	// The hook lives strictly inside the StatusHash-change gate.
	It("records no event when the status hash is unchanged", func() {
		cr := reconcileCR(buildsv1alpha1.BuildPhasePending, "same-hash")
		r, resources, builds, _, _, resolver := newReconcileHarness(cr)

		expectOwnedStack(resolver)
		resources.EXPECT().
			InternalGetByStackIDAndResourceName(gomock.Any(), testEventStackID, testEventResourceName).
			Return(&models.StackResource{ID: "res-1", StackID: testEventStackID, Name: testEventResourceName}, nil)
		builds.EXPECT().
			InternalGetByID(gomock.Any(), testEventBuildID).
			Return(&models.ImageBuild{
				ID:     testEventBuildID,
				Status: &models.ImageBuildStatus{LastObservedStatusHash: "same-hash"},
			}, nil)
		// No InternalUpdateStatus, no checker, no recorder: the hash gate is closed.

		_, err := r.Reconcile(context.Background(), ctrlreconcile.Request{
			NamespacedName: client.ObjectKey{Name: testEventBuildID, Namespace: testEventNamespace},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("updates status and skips the recorder when the hash changed but no release is active", func() {
		cr := reconcileCR(buildsv1alpha1.BuildPhasePending, "new-hash")
		r, resources, builds, checker, _, resolver := newReconcileHarness(cr)

		expectOwnedStack(resolver)
		resources.EXPECT().
			InternalGetByStackIDAndResourceName(gomock.Any(), testEventStackID, testEventResourceName).
			Return(&models.StackResource{ID: "res-1", StackID: testEventStackID, Name: testEventResourceName}, nil)
		builds.EXPECT().
			InternalGetByID(gomock.Any(), testEventBuildID).
			Return(&models.ImageBuild{
				ID:     testEventBuildID,
				Status: &models.ImageBuildStatus{LastObservedStatusHash: "old-hash"},
			}, nil)
		builds.EXPECT().
			InternalUpdateStatus(gomock.Any(), testEventBuildID, gomock.Any()).
			Return(nil)
		checker.EXPECT().
			InternalGetActiveByStackID(gomock.Any(), testEventStackID).
			Return(nil, nil)
		// No recorder EXPECT: no active release means no event.

		_, err := r.Reconcile(context.Background(), ctrlreconcile.Request{
			NamespacedName: client.ObjectKey{Name: testEventBuildID, Namespace: testEventNamespace},
		})
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("Reconcile cluster ownership guard", func() {
	It("skips builds owned by another cluster without touching services", func() {
		cr := reconcileCR(buildsv1alpha1.BuildPhasePending, "any-hash")
		r, _, _, _, _, resolver := newReconcileHarness(cr)

		resolver.EXPECT().
			InternalGetStack(gomock.Any(), testEventStackID).
			Return(&models.Stack{ID: testEventStackID, ClusterID: testEventOtherClusterID}, nil)
		// No resource/build/checker/recorder EXPECTs: strict mocks prove the skip.

		result, err := r.Reconcile(context.Background(), ctrlreconcile.Request{
			NamespacedName: client.ObjectKey{Name: testEventBuildID, Namespace: testEventNamespace},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(ctrlreconcile.Result{}))
	})

	It("ignores the build when the stack is not found", func() {
		cr := reconcileCR(buildsv1alpha1.BuildPhasePending, "any-hash")
		r, _, _, _, _, resolver := newReconcileHarness(cr)

		resolver.EXPECT().
			InternalGetStack(gomock.Any(), testEventStackID).
			Return(nil, apperrors.NotFound("stack not found"))

		result, err := r.Reconcile(context.Background(), ctrlreconcile.Request{
			NamespacedName: client.ObjectKey{Name: testEventBuildID, Namespace: testEventNamespace},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(ctrlreconcile.Result{}))
	})

	It("surfaces resolver errors", func() {
		cr := reconcileCR(buildsv1alpha1.BuildPhasePending, "any-hash")
		r, _, _, _, _, resolver := newReconcileHarness(cr)

		resolver.EXPECT().
			InternalGetStack(gomock.Any(), testEventStackID).
			Return(nil, apperrors.GeneralError("db down"))

		_, err := r.Reconcile(context.Background(), ctrlreconcile.Request{
			NamespacedName: client.ObjectKey{Name: testEventBuildID, Namespace: testEventNamespace},
		})
		Expect(err).To(HaveOccurred())
	})
})
