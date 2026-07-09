package workspaceresource

import (
	"testing"
	"time"

	"github.com/Stackdome/stackdome/pkg/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

func TestMapClusterStatusToServerStatus_withRuntimeCrash(t *testing.T) {
	cr := &corev1alpha1.StackResource{}
	cr.Name = "web"
	cr.Status = corev1alpha1.StackResourceStatus{
		Phase:      corev1alpha1.StackResourcePhaseFailed,
		StatusHash: "abc123",
		LastFailureDetails: []corev1alpha1.LastFailureDetail{
			{
				ContainerName:           "web",
				RestartCount:            5,
				LastTerminationReason:   "CrashLoopBackOff",
				LastTerminationMessage:  "back-off restarting",
				LastTerminationExitCode: ptr.To(int32(1)),
			},
		},
	}

	got := mapClusterStatusToServerStatus(cr)

	if got.LastFailure == nil {
		t.Fatal("expected LastFailure to be set")
	}
	if got.LastFailure.Type != models.FailureTypeRuntimeCrash {
		t.Errorf("Type = %q, want runtime_crash", got.LastFailure.Type)
	}
	if got.LastFailure.Container == nil {
		t.Fatal("expected Container to be set")
	}
	if got.LastFailure.Container.FailureType != "crash_loop" {
		t.Errorf("Container.FailureType = %q, want crash_loop", got.LastFailure.Container.FailureType)
	}
	if got.LastFailure.Container.RestartCount != 5 {
		t.Errorf("Container.RestartCount = %d, want 5", got.LastFailure.Container.RestartCount)
	}
	if got.LastFailure.InitContainer != nil {
		t.Error("expected InitContainer to be nil")
	}
}

func TestMapClusterStatusToServerStatus_withInitContainerCrash(t *testing.T) {
	cr := &corev1alpha1.StackResource{}
	cr.Name = "api"
	cr.Status = corev1alpha1.StackResourceStatus{
		Phase:      corev1alpha1.StackResourcePhaseFailed,
		StatusHash: "def456",
		LastFailureDetails: []corev1alpha1.LastFailureDetail{
			{
				ContainerName:           "api-init",
				RestartCount:            0,
				LastTerminationReason:   "Error",
				LastTerminationExitCode: ptr.To(int32(2)),
			},
		},
	}

	got := mapClusterStatusToServerStatus(cr)

	if got.LastFailure == nil {
		t.Fatal("expected LastFailure to be set")
	}
	if got.LastFailure.InitContainer == nil {
		t.Fatal("expected InitContainer to be set")
	}
	if got.LastFailure.InitContainer.FailureType != "exit_error" {
		t.Errorf("InitContainer.FailureType = %q, want exit_error", got.LastFailure.InitContainer.FailureType)
	}
	if got.LastFailure.Container != nil {
		t.Error("expected Container to be nil")
	}
}

func TestMapClusterStatusToServerStatus_noFailure(t *testing.T) {
	cr := &corev1alpha1.StackResource{}
	cr.Name = "web"
	cr.Status = corev1alpha1.StackResourceStatus{
		Phase:      corev1alpha1.StackResourcePhaseReady,
		StatusHash: "ghi789",
	}

	got := mapClusterStatusToServerStatus(cr)

	if got.LastFailure != nil {
		t.Errorf("expected LastFailure to be nil for ready resource, got %v", got.LastFailure)
	}
}

func TestComputeStatusRewrite_preservesBuildFailureWhenCRHasNoDetails(t *testing.T) {
	buildFailure := &models.StackResourceFailure{Type: models.FailureTypeBuildFailure}
	current := &models.StackResourceStatus{
		LastObservedStatusHash: "old-hash",
		LastFailure:            buildFailure,
	}
	cr := &corev1alpha1.StackResource{}
	cr.Name = "web"
	cr.Status = corev1alpha1.StackResourceStatus{
		Phase:      corev1alpha1.StackResourcePhaseFailed,
		StatusHash: "new-hash",
	}

	got := computeStatusRewrite(current, cr)

	if got.LastFailure != buildFailure {
		t.Fatalf("expected build failure to be carried over the rewrite, got %v", got.LastFailure)
	}
	if got.LastObservedStatusHash != "new-hash" {
		t.Errorf("LastObservedStatusHash = %q, want new-hash", got.LastObservedStatusHash)
	}
}

func TestComputeStatusRewrite_crFailureDetailsOverwriteBuildFailure(t *testing.T) {
	current := &models.StackResourceStatus{
		LastFailure: &models.StackResourceFailure{Type: models.FailureTypeBuildFailure},
	}
	cr := &corev1alpha1.StackResource{}
	cr.Name = "web"
	cr.Status = corev1alpha1.StackResourceStatus{
		Phase: corev1alpha1.StackResourcePhaseFailed,
		LastFailureDetails: []corev1alpha1.LastFailureDetail{
			{
				ContainerName:           "web",
				RestartCount:            3,
				LastTerminationReason:   "CrashLoopBackOff",
				LastTerminationExitCode: ptr.To(int32(1)),
			},
		},
	}

	got := computeStatusRewrite(current, cr)

	if got.LastFailure == nil {
		t.Fatal("expected LastFailure to be set")
	}
	if got.LastFailure.Type != models.FailureTypeRuntimeCrash {
		t.Errorf("Type = %q, want %q", got.LastFailure.Type, models.FailureTypeRuntimeCrash)
	}
}

func TestComputeStatusRewrite_doesNotCarryNonBuildFailure(t *testing.T) {
	current := &models.StackResourceStatus{
		LastFailure: &models.StackResourceFailure{Type: models.FailureTypeRuntimeCrash},
	}
	cr := &corev1alpha1.StackResource{}
	cr.Name = "web"
	cr.Status = corev1alpha1.StackResourceStatus{
		Phase: corev1alpha1.StackResourcePhaseReady,
	}

	got := computeStatusRewrite(current, cr)

	if got.LastFailure != nil {
		t.Errorf("expected runtime failure to be cleared by the CR rewrite, got %v", got.LastFailure)
	}
}

func TestComputeStatusRewrite_clearedBuildFailureStaysCleared(t *testing.T) {
	current := &models.StackResourceStatus{}
	cr := &corev1alpha1.StackResource{}
	cr.Name = "web"
	cr.Status = corev1alpha1.StackResourceStatus{
		Phase: corev1alpha1.StackResourcePhaseReady,
	}

	got := computeStatusRewrite(current, cr)

	if got.LastFailure != nil {
		t.Errorf("expected LastFailure to stay nil after the imagebuild controller cleared it, got %v", got.LastFailure)
	}
}

func TestComputeStatusRewrite_nilCurrentStatus(t *testing.T) {
	cr := &corev1alpha1.StackResource{}
	cr.Name = "web"
	cr.Status = corev1alpha1.StackResourceStatus{
		Phase: corev1alpha1.StackResourcePhaseReady,
	}

	got := computeStatusRewrite(nil, cr)

	if got == nil {
		t.Fatal("expected a status")
	}
	if got.LastFailure != nil {
		t.Errorf("expected LastFailure to be nil, got %v", got.LastFailure)
	}
}

func TestMapClusterStatusToServerStatus_lastRestartTime(t *testing.T) {
	now := metav1.NewTime(time.Now())
	cr := &corev1alpha1.StackResource{}
	cr.Name = "web"
	cr.Status = corev1alpha1.StackResourceStatus{
		LastRestartRequestProcessedAt: &now,
	}

	got := mapClusterStatusToServerStatus(cr)

	if got.LastRestartRequestProcessedAt == nil {
		t.Error("expected LastRestartRequestProcessedAt to be set")
	}
}
