package presenters_test

import (
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"k8s.io/utils/ptr"
)

func TestPresentStackResource_lastFailure_runtimeCrash(t *testing.T) {
	resource := &models.StackResource{
		Name: "web",
		Status: &models.StackResourceStatus{
			State: models.StackResourcePhaseReady,
			LastFailure: &models.StackResourceFailure{
				Type: models.FailureTypeRuntimeCrash,
				Container: &models.ContainerFailureDetail{
					FailureType:  "crash_loop",
					Reason:       "CrashLoopBackOff",
					Message:      "back-off restarting",
					RestartCount: 5,
					ExitCode:     ptr.To(int32(1)),
				},
			},
		},
	}

	out := presenters.PresentStackResource(resource)

	if out.Status == nil {
		t.Fatal("expected Status to be set")
	}
	if out.Status.LastFailure == nil {
		t.Fatal("expected LastFailure to be set")
	}
	if out.Status.LastFailure.Type == nil || *out.Status.LastFailure.Type != "runtime_crash" {
		t.Errorf("LastFailure.Type = %v, want runtime_crash", out.Status.LastFailure.Type)
	}
	if out.Status.LastFailure.Container == nil {
		t.Fatal("expected Container to be set")
	}
	if out.Status.LastFailure.Container.FailureType == nil || *out.Status.LastFailure.Container.FailureType != "crash_loop" {
		t.Errorf("Container.FailureType = %v, want crash_loop", out.Status.LastFailure.Container.FailureType)
	}
	if out.Status.LastFailure.Container.RestartCount == nil || *out.Status.LastFailure.Container.RestartCount != 5 {
		t.Errorf("Container.RestartCount = %v, want 5", out.Status.LastFailure.Container.RestartCount)
	}
}

func TestPresentStackResource_lastFailure_buildFailure(t *testing.T) {
	resource := &models.StackResource{
		Name: "web",
		Status: &models.StackResourceStatus{
			State: models.StackResourcePhaseFailed,
			LastFailure: &models.StackResourceFailure{
				Type: models.FailureTypeBuildFailure,
				Build: &models.BuildFailureDetail{
					FailureType:  "exit_error",
					Reason:       "Error",
					Message:      "COPY failed",
					RestartCount: 0,
					ExitCode:     ptr.To(int32(2)),
				},
			},
		},
	}

	out := presenters.PresentStackResource(resource)

	if out.Status.LastFailure == nil {
		t.Fatal("expected LastFailure to be set")
	}
	if out.Status.LastFailure.Type == nil || *out.Status.LastFailure.Type != "build_failure" {
		t.Errorf("LastFailure.Type = %v, want build_failure", out.Status.LastFailure.Type)
	}
	if out.Status.LastFailure.Build == nil {
		t.Fatal("expected Build to be set")
	}
	if out.Status.LastFailure.Build.FailureType == nil || *out.Status.LastFailure.Build.FailureType != "exit_error" {
		t.Errorf("Build.FailureType = %v, want exit_error", out.Status.LastFailure.Build.FailureType)
	}
}

func TestPresentStackResource_noFailure(t *testing.T) {
	resource := &models.StackResource{
		Name: "web",
		Status: &models.StackResourceStatus{
			State: models.StackResourcePhaseReady,
		},
	}

	out := presenters.PresentStackResource(resource)

	if out.Status != nil && out.Status.LastFailure != nil {
		t.Errorf("expected LastFailure to be nil")
	}
}
