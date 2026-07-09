package presenters_test

import (
	"testing"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"k8s.io/utils/ptr"
)

func TestPresentImageBuild_lastBuildFailureDetail(t *testing.T) {
	build := &models.ImageBuild{
		Status: &models.ImageBuildStatus{
			State: "Failed",
			LastBuildFailureDetail: &models.BuildFailureDetail{
				FailureType:  "exit_error",
				Reason:       "Error",
				Message:      "COPY failed: file not found",
				RestartCount: 0,
				ExitCode:     ptr.To(int32(1)),
			},
		},
	}

	out := presenters.PresentImageBuild(build)

	if out.Status == nil {
		t.Fatal("expected Status to be set")
	}
	if out.Status.LastBuildFailureDetail == nil {
		t.Fatal("expected LastBuildFailureDetail to be set")
	}
	if out.Status.LastBuildFailureDetail.FailureType == nil || *out.Status.LastBuildFailureDetail.FailureType != "exit_error" {
		t.Errorf("FailureType = %v, want exit_error", out.Status.LastBuildFailureDetail.FailureType)
	}
	if out.Status.LastBuildFailureDetail.Message == nil || *out.Status.LastBuildFailureDetail.Message != "COPY failed: file not found" {
		t.Errorf("Message = %v", out.Status.LastBuildFailureDetail.Message)
	}
}

func TestPresentImageBuild_noFailure(t *testing.T) {
	build := &models.ImageBuild{
		Status: &models.ImageBuildStatus{
			State:    "Success",
			ImageURL: "registry.io/img:tag",
		},
	}

	out := presenters.PresentImageBuild(build)

	if out.Status != nil && out.Status.LastBuildFailureDetail != nil {
		t.Errorf("expected LastBuildFailureDetail to be nil")
	}
}
