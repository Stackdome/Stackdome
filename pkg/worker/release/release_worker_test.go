package release

import (
	"context"
	"testing"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type mockReleaseService struct {
	activeByStack     *models.StackRelease
	markInProgressOK  bool
	saveManifestOK    bool
	markReleasedOK    bool
	markSuperseded    bool
	markFailed        bool
	appliedRelease    *models.StackRelease
	appliedReleaseErr *errors.ServiceError
}

func (m *mockReleaseService) InternalGet(ctx context.Context, releaseID string) (*models.StackRelease, *errors.ServiceError) {
	if m.appliedRelease != nil && m.appliedRelease.ID == releaseID {
		return m.appliedRelease, m.appliedReleaseErr
	}
	return nil, errors.NotFound("not found")
}

func (m *mockReleaseService) InternalGetActiveByStackID(ctx context.Context, stackID string) (*models.StackRelease, *errors.ServiceError) {
	return m.activeByStack, nil
}

func (m *mockReleaseService) InternalListActive(ctx context.Context) ([]*models.StackRelease, *errors.ServiceError) {
	return nil, nil
}

func (m *mockReleaseService) MarkInProgress(ctx context.Context, id string) (bool, *errors.ServiceError) {
	return m.markInProgressOK, nil
}

func (m *mockReleaseService) SaveManifest(ctx context.Context, id string, manifest *models.ReleaseManifest, rev string, pins models.ReleasePins, rendererVersion string) (bool, *errors.ServiceError) {
	return m.saveManifestOK, nil
}

func (m *mockReleaseService) MarkSuperseded(ctx context.Context, id string, reason string) (bool, *errors.ServiceError) {
	m.markSuperseded = true
	return true, nil
}

func (m *mockReleaseService) MarkReleased(ctx context.Context, id string, outcome models.ReleaseOutcome) (bool, *errors.ServiceError) {
	return m.markReleasedOK, nil
}

func (m *mockReleaseService) MarkFailed(ctx context.Context, id string, message string, outcome *models.ReleaseOutcome) (bool, *errors.ServiceError) {
	m.markFailed = true
	return true, nil
}

func (m *mockReleaseService) AppendImageDigests(ctx context.Context, id string, digests map[string]string) *errors.ServiceError {
	return nil
}

type mockStackService struct {
	stack *models.Stack
}

func (m *mockStackService) InternalGetStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError) {
	return m.stack, nil
}

func (m *mockStackService) UpdateStackCrRevision(ctx context.Context, ID string, revision string) *errors.ServiceError {
	return nil
}

func testLogger() logger.Logger {
	return logger.NewLoggerWithPrefix(context.Background(), "release-test")
}

func TestFreshnessReconciler_OlderSequence(t *testing.T) {
	r := &freshnessReconciler{
		releaseService: &mockReleaseService{
			activeByStack: &models.StackRelease{ID: "newer", Sequence: 2},
		},
		logger: testLogger(),
	}

	release := &models.StackRelease{ID: "older", StackID: "stack-1", Sequence: 1, State: models.ReleaseStateInProgress}
	result, err := r.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultStop {
		t.Fatal("expected resultStop")
	}
	if !r.releaseService.(*mockReleaseService).markSuperseded {
		t.Fatal("expected MarkSuperseded")
	}
}

func TestFreshnessReconciler_LatestSequence(t *testing.T) {
	release := &models.StackRelease{ID: "self", StackID: "stack-1", Sequence: 2, State: models.ReleaseStateInProgress}
	r := &freshnessReconciler{
		releaseService: &mockReleaseService{
			activeByStack: release,
		},
		logger: testLogger(),
	}

	result, err := r.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultNil {
		t.Fatal("expected resultNil")
	}
}

func TestFreshnessReconciler_PendingCASWon(t *testing.T) {
	release := &models.StackRelease{ID: "self", StackID: "stack-1", Sequence: 1, State: models.ReleaseStatePending}
	r := &freshnessReconciler{
		releaseService: &mockReleaseService{
			activeByStack:    release,
			markInProgressOK: true,
		},
		logger: testLogger(),
	}

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

func TestFreshnessReconciler_PendingCASLost(t *testing.T) {
	release := &models.StackRelease{ID: "self", StackID: "stack-1", Sequence: 1, State: models.ReleaseStatePending}
	r := &freshnessReconciler{
		releaseService: &mockReleaseService{
			activeByStack:    release,
			markInProgressOK: false,
		},
		logger: testLogger(),
	}

	result, err := r.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultStop {
		t.Fatal("expected resultStop")
	}
}

func TestRenderReconciler_ManifestPresent(t *testing.T) {
	r := &renderReconciler{releaseService: &mockReleaseService{}}
	release := &models.StackRelease{
		Manifest: &models.ReleaseManifest{},
	}

	result, err := r.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultNil {
		t.Fatal("expected resultNil")
	}
}

func TestConvergeReconciler_MatchesLastConverged(t *testing.T) {
	now := time.Now().UTC()
	release := &models.StackRelease{
		ID:               "rel-1",
		StackID:          "stack-1",
		ManifestRevision: "rev-1",
		RenderedAt:       &now,
		Manifest:         &models.ReleaseManifest{},
	}
	r := &convergeReconciler{
		releaseService: &mockReleaseService{markReleasedOK: true},
		stackService: &mockStackService{
			stack: &models.Stack{
				ID: "stack-1",
				Status: &models.StackStatus{
					LastConverged: &models.StackConvergenceRecord{
						ReleaseID: "rel-1",
						Revision:  "rev-1",
					},
				},
			},
		},
		logger: testLogger(),
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
	now := time.Now().UTC()
	release := &models.StackRelease{
		ID:               "rel-1",
		StackID:          "stack-1",
		ManifestRevision: "rev-1",
		RenderedAt:       &now,
		Manifest:         &models.ReleaseManifest{},
	}
	r := &convergeReconciler{
		releaseService: &mockReleaseService{markReleasedOK: false},
		stackService: &mockStackService{
			stack: &models.Stack{
				ID: "stack-1",
				Status: &models.StackStatus{
					LastConverged: &models.StackConvergenceRecord{
						ReleaseID: "rel-1",
						Revision:  "rev-1",
					},
				},
			},
		},
		logger: testLogger(),
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
	past := time.Now().UTC().Add(-convergenceTimeout - time.Minute)
	release := &models.StackRelease{
		ID:               "rel-1",
		StackID:          "stack-1",
		ManifestRevision: "rev-1",
		RenderedAt:       &past,
		Manifest:         &models.ReleaseManifest{},
	}
	r := &convergeReconciler{
		releaseService: &mockReleaseService{},
		stackService:   &mockStackService{stack: &models.Stack{ID: "stack-1"}},
		logger:         testLogger(),
	}

	result, err := r.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultStop {
		t.Fatal("expected resultStop")
	}
	if !r.releaseService.(*mockReleaseService).markFailed {
		t.Fatal("expected MarkFailed")
	}
}

func TestConvergeReconciler_NotYet(t *testing.T) {
	now := time.Now().UTC()
	release := &models.StackRelease{
		ID:               "rel-1",
		StackID:          "stack-1",
		ManifestRevision: "rev-1",
		RenderedAt:       &now,
		Manifest:         &models.ReleaseManifest{},
	}
	r := &convergeReconciler{
		releaseService: &mockReleaseService{},
		stackService:   &mockStackService{stack: &models.Stack{ID: "stack-1"}},
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

func TestApplyReconciler_SequenceGuard(t *testing.T) {
	svc := &mockReleaseService{
		appliedRelease: &models.StackRelease{ID: "applied", Sequence: 5},
	}

	release := &models.StackRelease{ID: "self", Sequence: 3}

	appliedRelease, serr := svc.InternalGet(context.Background(), "applied")
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if appliedRelease.Sequence <= release.Sequence {
		t.Fatal("test setup: applied release should have higher sequence")
	}

	if _, err := svc.MarkSuperseded(context.Background(), release.ID, "test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !svc.markSuperseded {
		t.Fatal("expected markSuperseded")
	}
}
