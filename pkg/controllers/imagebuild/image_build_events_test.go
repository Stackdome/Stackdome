package imagebuild

import (
	"context"
	"testing"

	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	ctrlreconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

const (
	testEventStackID      = "stack-1"
	testEventResourceName = "api"
	testEventBuildID      = "build-1"
	testEventNamespace    = "ns-1"
)

func newEventReconciler(t *testing.T) (*ImageBuildReconciler, *MockreleaseActiveChecker, *MockbuildEventRecorder) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
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

// TestRecordBuildEvent_phaseMapping covers the phase→event-type mapping and the
// best-effort active-release attribution when the build's revision does not
// match the active release's pins.
func TestRecordBuildEvent_phaseMapping(t *testing.T) {
	cases := []struct {
		name      string
		phase     buildsv1alpha1.BuildPhase
		eventType models.ReleaseEventType
	}{
		{"pending emits build_started", buildsv1alpha1.BuildPhasePending, models.ReleaseEventTypeBuildStarted},
		{"success emits build_succeeded", buildsv1alpha1.BuildPhaseSuccess, models.ReleaseEventTypeBuildSucceeded},
		{"failed emits build_failed", buildsv1alpha1.BuildPhaseFailed, models.ReleaseEventTypeBuildFailed},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, checker, recorder := newEventReconciler(t)
			active := &models.StackRelease{ID: "rel-1"}
			cr := eventBuildCR(tc.phase)
			build := gitBuildModel("deadbeef")

			checker.EXPECT().
				InternalGetActiveByStackID(gomock.Any(), testEventStackID).
				Return(active, nil)
			recorder.EXPECT().
				RecordBuildEvent(
					gomock.Any(),
					active,
					testEventResourceName,
					tc.eventType,
					testEventBuildID,
					models.ReleaseEventAttributionActiveRelease,
					"",
				).
				Return(nil)

			r.recordBuildEvent(context.Background(), testEventStackID, cr, build)
		})
	}
}

// TestRecordBuildEvent_failedIncludesReason asserts the failure reason is
// extracted from the build status the controller already reads.
func TestRecordBuildEvent_failedIncludesReason(t *testing.T) {
	r, checker, recorder := newEventReconciler(t)
	active := &models.StackRelease{ID: "rel-1"}
	cr := eventBuildCR(buildsv1alpha1.BuildPhaseFailed)
	cr.Status.LastBuildFailureDetail = &corev1alpha1.LastFailureDetail{
		LastTerminationReason:   "Error",
		LastTerminationMessage:  "COPY failed: file not found",
		LastTerminationExitCode: ptr.To(int32(1)),
	}
	build := gitBuildModel("deadbeef")

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
			"COPY failed: file not found",
		).
		Return(nil)

	r.recordBuildEvent(context.Background(), testEventStackID, cr, build)
}

// TestRecordBuildEvent_noEventPhases proves Cancelled and unknown phases emit
// nothing — not even an active-release lookup.
func TestRecordBuildEvent_noEventPhases(t *testing.T) {
	for _, phase := range []buildsv1alpha1.BuildPhase{buildsv1alpha1.BuildPhaseCancelled, buildsv1alpha1.BuildPhase("Weird")} {
		t.Run(string(phase), func(t *testing.T) {
			// No checker/recorder EXPECT: gomock fails the test on any call.
			r, _, _ := newEventReconciler(t)
			r.recordBuildEvent(context.Background(), testEventStackID, eventBuildCR(phase), gitBuildModel("deadbeef"))
		})
	}
}

// TestRecordBuildEvent_noActiveRelease proves that a nil active release (or a
// lookup error) skips the recorder call without failing.
func TestRecordBuildEvent_noActiveRelease(t *testing.T) {
	cases := []struct {
		name   string
		active *models.StackRelease
		err    *apperrors.ServiceError
	}{
		{"nil active release", nil, nil},
		{"lookup error", nil, apperrors.GeneralError("boom")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, checker, _ := newEventReconciler(t)
			checker.EXPECT().
				InternalGetActiveByStackID(gomock.Any(), testEventStackID).
				Return(tc.active, tc.err)
			// No recorder EXPECT: it must not be called.
			r.recordBuildEvent(context.Background(), testEventStackID, eventBuildCR(buildsv1alpha1.BuildPhasePending), gitBuildModel("deadbeef"))
		})
	}
}

// TestRecordBuildEvent_pinMatchDeterministicAttribution proves that when the
// active release's pins deterministically identify the build's revision, the
// attribution marker is dropped (empty).
func TestRecordBuildEvent_pinMatchDeterministicAttribution(t *testing.T) {
	cases := []struct {
		name  string
		pins  models.ReleasePins
		build *models.ImageBuild
	}{
		{
			name: "git commit pin match",
			pins: models.ReleasePins{Resources: map[string]models.ResourcePins{
				testEventResourceName: {GitSHA: "abc123"},
			}},
			build: gitBuildModel("abc123"),
		},
		{
			name: "volume hash pin match",
			pins: models.ReleasePins{Resources: map[string]models.ResourcePins{
				testEventResourceName: {VolumeHash: "vh-9"},
			}},
			build: &models.ImageBuild{
				ID: testEventBuildID,
				Spec: models.BuildConfigSpec{
					SourceRevision: models.BuildSourceRevision{
						Volume: &models.VolumeRevision{CurrentVolumeHash: "vh-9"},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, checker, recorder := newEventReconciler(t)
			active := &models.StackRelease{ID: "rel-1", Pins: tc.pins}

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
					"",
				).
				Return(nil)

			r.recordBuildEvent(context.Background(), testEventStackID, eventBuildCR(buildsv1alpha1.BuildPhasePending), tc.build)
		})
	}
}

// TestRecordBuildEvent_pinMismatchFallsBack proves that a pin that exists but
// does not match the build revision falls back to active-release attribution.
func TestRecordBuildEvent_pinMismatchFallsBack(t *testing.T) {
	r, checker, recorder := newEventReconciler(t)
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
			models.ReleaseEventAttributionActiveRelease, "",
		).
		Return(nil)

	r.recordBuildEvent(context.Background(), testEventStackID, eventBuildCR(buildsv1alpha1.BuildPhasePending), gitBuildModel("differentsha"))
}

// TestRecordBuildEvent_recorderErrorIsSwallowed proves a recorder failure is
// logged only and never propagated out of the helper.
func TestRecordBuildEvent_recorderErrorIsSwallowed(t *testing.T) {
	r, checker, recorder := newEventReconciler(t)
	active := &models.StackRelease{ID: "rel-1"}

	checker.EXPECT().
		InternalGetActiveByStackID(gomock.Any(), testEventStackID).
		Return(active, nil)
	recorder.EXPECT().
		RecordBuildEvent(gomock.Any(), active, testEventResourceName, gomock.Any(), testEventBuildID, gomock.Any(), gomock.Any()).
		Return(apperrors.GeneralError("db down"))

	// Must not panic or otherwise surface the error.
	r.recordBuildEvent(context.Background(), testEventStackID, eventBuildCR(buildsv1alpha1.BuildPhaseSuccess), gitBuildModel("deadbeef"))
}

// --- Reconcile-level integration of the event hook ---

func eventsTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("client-go scheme: %v", err)
	}
	if err := buildsv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("builds scheme: %v", err)
	}
	return scheme
}

func reconcileCR(phase buildsv1alpha1.BuildPhase, statusHash string) *buildsv1alpha1.ImageBuild {
	cr := eventBuildCR(phase)
	cr.Labels = map[string]string{corev1alpha1.LabelStackID: testEventStackID}
	cr.Status.StatusHash = statusHash
	return cr
}

func newReconcileHarness(t *testing.T, cr *buildsv1alpha1.ImageBuild) (
	*ImageBuildReconciler, *mocks.MockStackResourceService, *mocks.MockImageBuildService, *MockreleaseActiveChecker, *MockbuildEventRecorder,
) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	resources := mocks.NewMockStackResourceService(ctrl)
	builds := mocks.NewMockImageBuildService(ctrl)
	checker := NewMockreleaseActiveChecker(ctrl)
	recorder := NewMockbuildEventRecorder(ctrl)

	fakeClient := fake.NewClientBuilder().WithScheme(eventsTestScheme(t)).WithObjects(cr).Build()

	r := &ImageBuildReconciler{
		Client:              fakeClient,
		DBResourceService:   resources,
		DBImageBuildService: builds,
		Logger:              logger.NewLoggerWithPrefix(context.Background(), "reconcile-test"),
		releaseChecker:      checker,
		eventRecorder:       recorder,
	}
	return r, resources, builds, checker, recorder
}

// TestReconcile_unchangedStatusHashSkipsEvent proves the hook lives strictly
// inside the StatusHash-change gate: an unchanged hash records no event and the
// reconcile still succeeds.
func TestReconcile_unchangedStatusHashSkipsEvent(t *testing.T) {
	cr := reconcileCR(buildsv1alpha1.BuildPhasePending, "same-hash")
	r, resources, builds, _, _ := newReconcileHarness(t, cr)

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
	if err != nil {
		t.Fatalf("unexpected reconcile error: %v", err)
	}
}

// TestReconcile_changedHashNoActiveReleaseSucceeds drives the full hook path:
// a changed hash updates status and consults the release checker, but with no
// active release the recorder is never called and the reconcile succeeds.
func TestReconcile_changedHashNoActiveReleaseSucceeds(t *testing.T) {
	cr := reconcileCR(buildsv1alpha1.BuildPhasePending, "new-hash")
	r, resources, builds, checker, _ := newReconcileHarness(t, cr)

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
	if err != nil {
		t.Fatalf("unexpected reconcile error: %v", err)
	}
}
