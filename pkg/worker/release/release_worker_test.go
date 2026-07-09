package release

import (
	"context"
	"testing"
	"time"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

func testLogger() logger.Logger {
	return logger.NewLoggerWithPrefix(context.Background(), "release-test")
}

func TestNewReleaseWorker_PanicsWithoutEventRecorder(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when EventRecorder is nil")
		}
	}()
	_ = NewReleaseWorker(ReleaseWorkerSpec{})
}

func TestGatekeeperReconciler_OlderSequence(t *testing.T) {
	ctrl := gomock.NewController(t)

	svc := NewMockreleaseService(ctrl)
	svc.EXPECT().InternalGetActiveByStackID(gomock.Any(), "stack-1").
		Return(&models.StackRelease{ID: "newer", Sequence: 2}, nil)
	svc.EXPECT().MarkSuperseded(gomock.Any(), "older", gomock.Any()).
		Return(true, nil)

	rec := NewMockeventRecorder(ctrl)
	rec.EXPECT().RecordReleaseTerminal(gomock.Any(), gomock.Any(), models.ReleaseStateSuperseded,
		"Release superseded by release #2").Return(nil)

	r := &gatekeeperReconciler{releaseService: svc, eventRecorder: rec, logger: testLogger()}

	release := &models.StackRelease{ID: "older", StackID: "stack-1", Sequence: 1, State: models.ReleaseStateInProgress}
	result, err := r.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultStop {
		t.Fatal("expected resultStop")
	}
}

func TestGatekeeperReconciler_LatestSequence(t *testing.T) {
	ctrl := gomock.NewController(t)

	release := &models.StackRelease{ID: "self", StackID: "stack-1", Sequence: 2, State: models.ReleaseStateInProgress}

	svc := NewMockreleaseService(ctrl)
	svc.EXPECT().InternalGetActiveByStackID(gomock.Any(), "stack-1").
		Return(release, nil)

	rec := NewMockeventRecorder(ctrl)

	r := &gatekeeperReconciler{releaseService: svc, eventRecorder: rec, logger: testLogger()}

	result, err := r.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultNil {
		t.Fatal("expected resultNil")
	}
}

func TestGatekeeperReconciler_PendingCASWon(t *testing.T) {
	ctrl := gomock.NewController(t)

	release := &models.StackRelease{ID: "self", StackID: "stack-1", Sequence: 1, State: models.ReleaseStatePending}

	svc := NewMockreleaseService(ctrl)
	svc.EXPECT().InternalGetActiveByStackID(gomock.Any(), "stack-1").
		Return(release, nil)
	svc.EXPECT().MarkInProgress(gomock.Any(), "self").
		Return(true, nil)

	rec := NewMockeventRecorder(ctrl)
	rec.EXPECT().RecordReleaseStarted(gomock.Any(), release).Return(nil)

	r := &gatekeeperReconciler{releaseService: svc, eventRecorder: rec, logger: testLogger()}

	result, err := r.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultNil {
		t.Fatal("expected resultNil")
	}
	if release.State != models.ReleaseStateInProgress {
		t.Fatalf("expected InProgress, got %s", release.State)
	}
}

func TestGatekeeperReconciler_PendingCASWon_RecorderErrorIsLogOnly(t *testing.T) {
	ctrl := gomock.NewController(t)

	release := &models.StackRelease{ID: "self", StackID: "stack-1", Sequence: 1, State: models.ReleaseStatePending}

	svc := NewMockreleaseService(ctrl)
	svc.EXPECT().InternalGetActiveByStackID(gomock.Any(), "stack-1").
		Return(release, nil)
	svc.EXPECT().MarkInProgress(gomock.Any(), "self").
		Return(true, nil)

	rec := NewMockeventRecorder(ctrl)
	rec.EXPECT().RecordReleaseStarted(gomock.Any(), release).
		Return(errors.GeneralError("event insert failed"))

	r := &gatekeeperReconciler{releaseService: svc, eventRecorder: rec, logger: testLogger()}

	result, err := r.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("recorder error must not fail the reconciler, got: %v", err)
	}
	if !result.resultNil {
		t.Fatal("expected resultNil")
	}
}

func TestGatekeeperReconciler_PendingCASLost(t *testing.T) {
	ctrl := gomock.NewController(t)

	release := &models.StackRelease{ID: "self", StackID: "stack-1", Sequence: 1, State: models.ReleaseStatePending}

	svc := NewMockreleaseService(ctrl)
	svc.EXPECT().InternalGetActiveByStackID(gomock.Any(), "stack-1").
		Return(release, nil)
	svc.EXPECT().MarkInProgress(gomock.Any(), "self").
		Return(false, nil)

	// CAS lost: recorder must never be called.
	rec := NewMockeventRecorder(ctrl)

	r := &gatekeeperReconciler{releaseService: svc, eventRecorder: rec, logger: testLogger()}

	result, err := r.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultStop {
		t.Fatal("expected resultStop")
	}
}

func TestRenderReconciler_ManifestPresent(t *testing.T) {
	ctrl := gomock.NewController(t)

	r := &renderReconciler{releaseService: NewMockreleaseService(ctrl)}
	release := &models.StackRelease{Manifest: &models.ReleaseManifest{}}

	result, err := r.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultNil {
		t.Fatal("expected resultNil")
	}
}

func TestConvergeReconciler_MatchesLastConverged(t *testing.T) {
	ctrl := gomock.NewController(t)
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

	rec := NewMockeventRecorder(ctrl)
	rec.EXPECT().RecordReleaseTerminal(gomock.Any(), release, models.ReleaseStateReleased, "Release is live").
		Return(nil)

	r := &convergeReconciler{
		releaseService: relSvc,
		stackService:   stackSvc,
		eventRecorder:  rec,
		logger:         testLogger(),
	}

	result, err := r.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultStop {
		t.Fatal("expected resultStop")
	}
}

func TestConvergeReconciler_MarkReleasedCASFails(t *testing.T) {
	ctrl := gomock.NewController(t)
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

	// CAS lost: recorder must never be called.
	rec := NewMockeventRecorder(ctrl)

	r := &convergeReconciler{
		releaseService: relSvc,
		stackService:   stackSvc,
		eventRecorder:  rec,
		logger:         testLogger(),
	}

	result, err := r.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultStop {
		t.Fatal("expected resultStop when CAS fails")
	}
}

func TestConvergeReconciler_Timeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	past := time.Now().UTC().Add(-time.Duration(models.DefaultDeployTimeoutMinutes)*time.Minute - time.Minute)

	release := &models.StackRelease{
		ID:               "rel-1",
		StackID:          "stack-1",
		ManifestRevision: "rev-1",
		RenderedAt:       &past,
		Manifest:         &models.ReleaseManifest{},
	}

	relSvc := NewMockreleaseService(ctrl)
	relSvc.EXPECT().MarkFailed(gomock.Any(), "rel-1", gomock.Any(), gomock.Any()).
		Return(true, nil)

	stackSvc := NewMockstackService(ctrl)
	stackSvc.EXPECT().InternalGetStack(gomock.Any(), "stack-1").
		Return(&models.Stack{ID: "stack-1"}, nil)

	rec := NewMockeventRecorder(ctrl)
	rec.EXPECT().RecordReleaseTerminal(gomock.Any(), release, models.ReleaseStateFailed, gomock.Any()).
		Return(nil)

	r := &convergeReconciler{
		releaseService: relSvc,
		stackService:   stackSvc,
		eventRecorder:  rec,
		logger:         testLogger(),
	}

	result, err := r.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultStop {
		t.Fatal("expected resultStop")
	}
}

func TestConvergeReconciler_NotYet(t *testing.T) {
	ctrl := gomock.NewController(t)
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

	r := &convergeReconciler{
		releaseService: NewMockreleaseService(ctrl),
		stackService:   stackSvc,
		eventRecorder:  NewMockeventRecorder(ctrl),
		logger:         testLogger(),
	}

	result, err := r.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.resultRequeueAfter == nil || *result.resultRequeueAfter != convergencePollInterval {
		t.Fatalf("expected requeue after %s, got %v", convergencePollInterval, result.resultRequeueAfter)
	}
}

func TestApplyReconciler_SupersededByClusterCR(t *testing.T) {
	ctrl := gomock.NewController(t)

	relSvc := NewMockreleaseService(ctrl)
	relSvc.EXPECT().InternalGet(gomock.Any(), "higher-release").
		Return(&models.StackRelease{ID: "higher-release", Sequence: 5}, nil)
	relSvc.EXPECT().MarkSuperseded(gomock.Any(), "self", gomock.Any()).
		Return(true, nil)

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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !superseded {
		t.Fatal("expected release to be superseded")
	}
}

func TestApplyReconciler_NotSupersededByLowerSequence(t *testing.T) {
	ctrl := gomock.NewController(t)

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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if superseded {
		t.Fatal("expected release NOT to be superseded by lower sequence")
	}
}

func TestApplyReconciler_NoManifestSkips(t *testing.T) {
	r := &applyReconciler{logger: testLogger()}

	release := &models.StackRelease{ID: "self", Manifest: nil}
	result, err := r.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultNil {
		t.Fatal("expected resultNil when manifest is nil")
	}
}

func TestApplyReconciler_StackDeletedFails(t *testing.T) {
	ctrl := gomock.NewController(t)
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

	rec := NewMockeventRecorder(ctrl)
	rec.EXPECT().RecordReleaseTerminal(gomock.Any(), gomock.Any(), models.ReleaseStateFailed, gomock.Any()).
		Return(nil)

	r := &applyReconciler{
		releaseService: relSvc,
		stackService:   stackSvc,
		eventRecorder:  rec,
		logger:         testLogger(),
	}

	release := &models.StackRelease{
		ID:       "self",
		StackID:  "stack-1",
		Manifest: &models.ReleaseManifest{},
	}
	result, err := r.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultStop {
		t.Fatal("expected resultStop when stack is being deleted")
	}
}
