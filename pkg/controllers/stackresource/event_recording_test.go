package workspaceresource

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
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

func TestResourceEventForState(t *testing.T) {
	cases := []struct {
		name         string
		state        models.StackResourceState
		conditions   []models.Condition
		statusReason string
		wantType     models.ReleaseEventType
		wantReason   string
		wantEmit     bool
	}{
		{
			name:  "pending with dependencies not ready records waiting with the condition message",
			state: models.StackResourcePhasePending,
			conditions: []models.Condition{
				{Type: string(corev1alpha1.StackResourceDependenciesReady), Status: string(models.ConditionFalse), Message: "waiting for mysql"},
			},
			wantType:   models.ReleaseEventTypeResourceWaiting,
			wantReason: "waiting for mysql",
			wantEmit:   true,
		},
		{
			name:  "pending while building suppresses the resource event",
			state: models.StackResourcePhasePending,
			conditions: []models.Condition{
				{Type: string(corev1alpha1.StackResourceBuildReady), Status: string(models.ConditionFalse), Message: "building"},
			},
			wantEmit: false,
		},
		{
			name:     "pending with no adverse conditions records deploying",
			state:    models.StackResourcePhasePending,
			wantType: models.ReleaseEventTypeResourceDeploying,
			wantEmit: true,
		},
		{
			name:  "dependencies-not-ready wins over build-not-ready",
			state: models.StackResourcePhasePending,
			conditions: []models.Condition{
				{Type: string(corev1alpha1.StackResourceBuildReady), Status: string(models.ConditionFalse), Message: "building"},
				{Type: string(corev1alpha1.StackResourceDependenciesReady), Status: string(models.ConditionFalse), Message: "waiting for redis"},
			},
			wantType:   models.ReleaseEventTypeResourceWaiting,
			wantReason: "waiting for redis",
			wantEmit:   true,
		},
		{
			name:     "ready records resource ready",
			state:    models.StackResourcePhaseReady,
			wantType: models.ReleaseEventTypeResourceReady,
			wantEmit: true,
		},
		{
			name:         "failed passes the status reason through",
			state:        models.StackResourcePhaseFailed,
			statusReason: "container exited",
			wantType:     models.ReleaseEventTypeResourceFailed,
			wantReason:   "container exited",
			wantEmit:     true,
		},
		{
			name:         "failed with a stalled condition prefers the stalled message",
			state:        models.StackResourcePhaseFailed,
			statusReason: "container exited",
			conditions: []models.Condition{
				{Type: string(corev1alpha1.StackResourceStalled), Status: string(models.ConditionTrue), Message: "image pull backoff"},
			},
			wantType:   models.ReleaseEventTypeResourceFailed,
			wantReason: "image pull backoff",
			wantEmit:   true,
		},
		{
			name:         "unknown maps to resource failed",
			state:        models.StackResourcePhaseUnknown,
			statusReason: "cluster unreachable",
			wantType:     models.ReleaseEventTypeResourceFailed,
			wantReason:   "cluster unreachable",
			wantEmit:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotType, gotReason, gotEmit := resourceEventForState(tc.state, tc.conditions, tc.statusReason)
			if gotEmit != tc.wantEmit {
				t.Fatalf("emit = %v, want %v", gotEmit, tc.wantEmit)
			}
			if !tc.wantEmit {
				return
			}
			if gotType != tc.wantType {
				t.Errorf("type = %q, want %q", gotType, tc.wantType)
			}
			if gotReason != tc.wantReason {
				t.Errorf("reason = %q, want %q", gotReason, tc.wantReason)
			}
		})
	}
}

func stackResourceTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("scheme setup failed: %v", err)
	}
	if err := corev1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("scheme setup failed: %v", err)
	}
	return scheme
}

func stackResourceCR(hash string, phase corev1alpha1.StackResourcePhase) *corev1alpha1.StackResource {
	return &corev1alpha1.StackResource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web",
			Namespace: "ns-1",
			Labels:    map[string]string{corev1alpha1.LabelStackID: "stack-1"},
		},
		Status: corev1alpha1.StackResourceStatus{
			Phase:      phase,
			StatusHash: hash,
		},
	}
}

func newReconcilerForTest(t *testing.T, cr *corev1alpha1.StackResource) (
	*stackResourceReconciler,
	*mocks.MockStackResourceService,
	*MockreleaseActiveChecker,
	*MockresourceEventRecorder,
) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	svc := mocks.NewMockStackResourceService(ctrl)
	checker := NewMockreleaseActiveChecker(ctrl)
	recorder := NewMockresourceEventRecorder(ctrl)
	r := &stackResourceReconciler{
		client:               fake.NewClientBuilder().WithScheme(stackResourceTestScheme(t)).WithObjects(cr).Build(),
		stackResourceService: svc,
		releaseChecker:       checker,
		eventRecorder:        recorder,
		logger:               logger.NewLoggerWithPrefix(context.Background(), "sr-test"),
	}
	return r, svc, checker, recorder
}

func reconcileRequest() ctrl.Request {
	return ctrl.Request{NamespacedName: types.NamespacedName{Name: "web", Namespace: "ns-1"}}
}

// Unchanged StatusHash keeps the controller out of the status-rewrite gate, so
// nothing downstream — status update, active-release lookup, event record — runs.
func TestReconcile_unchangedStatusHashRecordsNothing(t *testing.T) {
	cr := stackResourceCR("h1", corev1alpha1.StackResourcePhaseReady)
	r, svc, _, _ := newReconcilerForTest(t, cr)

	db := &models.StackResource{ID: "sr-1", Name: "web", Status: &models.StackResourceStatus{LastObservedStatusHash: "h1"}}
	svc.EXPECT().InternalGetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(db, nil)
	// No UpdateStatus / checker / recorder expectations: gomock fails on any call.

	if _, err := r.Reconcile(context.Background(), reconcileRequest()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// No active release means the resource event is silently skipped, and the
// reconcile still succeeds because the status update already persisted.
func TestReconcile_noActiveReleaseSkipsEventAndSucceeds(t *testing.T) {
	cr := stackResourceCR("h2", corev1alpha1.StackResourcePhaseReady)
	r, svc, checker, _ := newReconcilerForTest(t, cr)

	db := &models.StackResource{ID: "sr-1", Name: "web", Status: nil}
	svc.EXPECT().InternalGetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(db, nil)
	svc.EXPECT().UpdateStatus(gomock.Any(), "sr-1", gomock.Any()).Return(nil)
	checker.EXPECT().InternalGetActiveByStackID(gomock.Any(), "stack-1").Return(nil, nil)
	// recorder must not be called.

	if _, err := r.Reconcile(context.Background(), reconcileRequest()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// A recorder failure is non-fatal: the reconcile still returns no error.
func TestReconcile_recorderErrorDoesNotFailReconcile(t *testing.T) {
	cr := stackResourceCR("h3", corev1alpha1.StackResourcePhaseReady)
	r, svc, checker, recorder := newReconcilerForTest(t, cr)

	release := &models.StackRelease{ID: "rel-1"}
	db := &models.StackResource{ID: "sr-1", Name: "web", Status: &models.StackResourceStatus{LastObservedStatusHash: "old"}}
	svc.EXPECT().InternalGetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(db, nil)
	svc.EXPECT().UpdateStatus(gomock.Any(), "sr-1", gomock.Any()).Return(nil)
	checker.EXPECT().InternalGetActiveByStackID(gomock.Any(), "stack-1").Return(release, nil)
	recorder.EXPECT().
		RecordResourceEvent(gomock.Any(), release, "web", models.ReleaseEventTypeResourceReady, "").
		Return(apperrors.GeneralError("boom"))

	if _, err := r.Reconcile(context.Background(), reconcileRequest()); err != nil {
		t.Fatalf("expected recorder error to be swallowed, got %v", err)
	}
}

// Happy path: a changed StatusHash on an active release records the mapped event.
func TestReconcile_recordsResourceEventForActiveRelease(t *testing.T) {
	cr := stackResourceCR("h4", corev1alpha1.StackResourcePhaseReady)
	r, svc, checker, recorder := newReconcilerForTest(t, cr)

	release := &models.StackRelease{ID: "rel-1"}
	db := &models.StackResource{ID: "sr-1", Name: "web", Status: &models.StackResourceStatus{LastObservedStatusHash: "old"}}
	svc.EXPECT().InternalGetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(db, nil)
	svc.EXPECT().UpdateStatus(gomock.Any(), "sr-1", gomock.Any()).Return(nil)
	checker.EXPECT().InternalGetActiveByStackID(gomock.Any(), "stack-1").Return(release, nil)
	recorder.EXPECT().
		RecordResourceEvent(gomock.Any(), release, "web", models.ReleaseEventTypeResourceReady, "").
		Return(nil)

	if _, err := r.Reconcile(context.Background(), reconcileRequest()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
