package controllers_test

import (
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/controllers"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"k8s.io/utils/ptr"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

func TestMapFailureType(t *testing.T) {
	cases := []struct {
		reason   string
		expected string
	}{
		{"CrashLoopBackOff", "crash_loop"},
		{"OOMKilled", "out_of_memory"},
		{"ImagePullBackOff", "image_pull_failed"},
		{"ErrImagePull", "image_pull_failed"},
		{"CreateContainerError", "create_container_error"},
		{"Error", "exit_error"},
		{"", "exit_error"},
		{"SomethingUnknown", "exit_error"},
	}
	for _, tc := range cases {
		t.Run(tc.reason, func(t *testing.T) {
			got := controllers.MapFailureType(tc.reason)
			if got != tc.expected {
				t.Errorf("MapFailureType(%q) = %q, want %q", tc.reason, got, tc.expected)
			}
		})
	}
}

func TestMapLastFailureDetails_nil(t *testing.T) {
	got := controllers.MapLastFailureDetails("web", nil)
	if got != nil {
		t.Errorf("expected nil for empty details, got %v", got)
	}
}

func TestMapLastFailureDetails_mainContainer(t *testing.T) {
	details := []corev1alpha1.LastFailureDetail{
		{
			ContainerName:           "web",
			RestartCount:            3,
			LastTerminationReason:   "CrashLoopBackOff",
			LastTerminationMessage:  "back-off restarting",
			LastTerminationExitCode: ptr.To(int32(1)),
		},
	}
	got := controllers.MapLastFailureDetails("web", details)
	if got == nil {
		t.Fatal("expected non-nil failure")
	}
	if got.Type != models.FailureTypeRuntimeCrash {
		t.Errorf("Type = %q, want %q", got.Type, models.FailureTypeRuntimeCrash)
	}
	if got.Container == nil {
		t.Fatal("expected Container to be set")
	}
	if got.Container.FailureType != "crash_loop" {
		t.Errorf("Container.FailureType = %q, want crash_loop", got.Container.FailureType)
	}
	if got.Container.RestartCount != 3 {
		t.Errorf("Container.RestartCount = %d, want 3", got.Container.RestartCount)
	}
	if got.InitContainer != nil {
		t.Errorf("expected InitContainer to be nil")
	}
}

func TestMapLastFailureDetails_initContainer(t *testing.T) {
	details := []corev1alpha1.LastFailureDetail{
		{
			ContainerName:           "web-init",
			RestartCount:            1,
			LastTerminationReason:   "Error",
			LastTerminationExitCode: ptr.To(int32(2)),
		},
	}
	got := controllers.MapLastFailureDetails("web", details)
	if got == nil {
		t.Fatal("expected non-nil failure")
	}
	if got.InitContainer == nil {
		t.Fatal("expected InitContainer to be set")
	}
	if got.InitContainer.FailureType != "exit_error" {
		t.Errorf("InitContainer.FailureType = %q, want exit_error", got.InitContainer.FailureType)
	}
	if got.Container != nil {
		t.Errorf("expected Container to be nil")
	}
}

func TestMapBuildFailureDetail_nil(t *testing.T) {
	got := controllers.MapBuildFailureDetail(nil)
	if got != nil {
		t.Errorf("expected nil for nil input")
	}
}

func TestMapBuildFailureDetail(t *testing.T) {
	detail := &corev1alpha1.LastFailureDetail{
		ContainerName:           "kaniko",
		RestartCount:            0,
		LastTerminationReason:   "Error",
		LastTerminationMessage:  "COPY failed: file not found",
		LastTerminationExitCode: ptr.To(int32(1)),
	}
	got := controllers.MapBuildFailureDetail(detail)
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	if got.FailureType != "exit_error" {
		t.Errorf("FailureType = %q, want exit_error", got.FailureType)
	}
	if got.Message != "COPY failed: file not found" {
		t.Errorf("Message = %q", got.Message)
	}
	if got.ExitCode == nil || *got.ExitCode != 1 {
		t.Errorf("ExitCode = %v, want 1", got.ExitCode)
	}
}
