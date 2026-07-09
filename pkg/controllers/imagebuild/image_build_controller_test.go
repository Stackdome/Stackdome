package imagebuild

import (
	"context"
	"testing"

	"github.com/Stackdome/stackdome/pkg/models"
	"k8s.io/utils/ptr"
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
