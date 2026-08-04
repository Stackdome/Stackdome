package controllers_test

import (
	"testing"

	"github.com/Stackdome/stackdome/pkg/controllers"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

func TestMapFailureType(t *testing.T) {
	cases := []struct {
		reason   string
		expected string
	}{
		{controllers.ReasonCrashLoopBackOff, controllers.FailureTypeCrashLoop},
		{controllers.ReasonOOMKilled, controllers.FailureTypeOutOfMemory},
		{controllers.ReasonImagePullBackOff, controllers.FailureTypeImagePullFailed},
		{controllers.ReasonErrImagePull, controllers.FailureTypeImagePullFailed},
		{controllers.ReasonCreateContainerError, controllers.FailureTypeCreateContainerError},
		{controllers.ReasonPortNotListening, controllers.FailureTypePortNotListening},
		{"Error", controllers.FailureTypeExitError},
		{"", controllers.FailureTypeExitError},
		{"SomethingUnknown", controllers.FailureTypeExitError},
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
	got := controllers.MapLastFailureDetails("web", "", nil)
	if got != nil {
		t.Errorf("expected nil for empty details, got %v", got)
	}
}

func TestMapLastFailureDetails_mainContainer(t *testing.T) {
	details := []corev1alpha1.LastFailureDetail{
		{
			ContainerName:           "web",
			RestartCount:            3,
			LastTerminationReason:   controllers.ReasonCrashLoopBackOff,
			LastTerminationMessage:  "back-off restarting",
			LastTerminationExitCode: ptr.To(int32(1)),
		},
	}
	got := controllers.MapLastFailureDetails("web", "", details)
	if got == nil {
		t.Fatal("expected non-nil failure")
	}
	if got.Type != models.FailureTypeRuntimeCrash {
		t.Errorf("Type = %q, want %q", got.Type, models.FailureTypeRuntimeCrash)
	}
	if got.Container == nil {
		t.Fatal("expected Container to be set")
	}
	if got.Container.FailureType != controllers.FailureTypeCrashLoop {
		t.Errorf("Container.FailureType = %q, want %q", got.Container.FailureType, controllers.FailureTypeCrashLoop)
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
	got := controllers.MapLastFailureDetails("web", "", details)
	if got == nil {
		t.Fatal("expected non-nil failure")
	}
	if got.InitContainer == nil {
		t.Fatal("expected InitContainer to be set")
	}
	if got.InitContainer.FailureType != controllers.FailureTypeExitError {
		t.Errorf("InitContainer.FailureType = %q, want %q", got.InitContainer.FailureType, controllers.FailureTypeExitError)
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
	if got.FailureType != controllers.FailureTypeExitError {
		t.Errorf("FailureType = %q, want %q", got.FailureType, controllers.FailureTypeExitError)
	}
	if got.Message != "COPY failed: file not found" {
		t.Errorf("Message = %q", got.Message)
	}
	if got.ExitCode == nil || *got.ExitCode != 1 {
		t.Errorf("ExitCode = %v, want 1", got.ExitCode)
	}
}

var _ = Describe("MapLastFailureDetails", func() {
	const resourceName = "web"

	It("types a readiness failure from the agent's detail type", func() {
		failure := controllers.MapLastFailureDetails(resourceName, "", []corev1alpha1.LastFailureDetail{{
			Type:                   corev1alpha1.FailureTypeReadinessFailure,
			ContainerName:          resourceName,
			LastTerminationReason:  controllers.ReasonPortNotListening,
			LastTerminationMessage: "readiness check failed: nothing listening on port 8080",
		}})

		Expect(failure.Type).To(Equal(models.FailureTypeReadinessFailure))
		Expect(failure.Container).NotTo(BeNil())
		Expect(failure.Container.FailureType).To(Equal(controllers.FailureTypePortNotListening))
		Expect(failure.Container.Reason).To(Equal(controllers.ReasonPortNotListening))
		Expect(failure.Container.Message).To(ContainSubstring("nothing listening on port 8080"))
		Expect(failure.Container.ExitCode).To(BeNil())
		Expect(failure.Container.RestartCount).To(BeZero())
	})

	It("treats an empty detail type as a runtime crash, for older agents", func() {
		failure := controllers.MapLastFailureDetails(resourceName, "", []corev1alpha1.LastFailureDetail{{
			ContainerName:         resourceName,
			LastTerminationReason: controllers.ReasonCrashLoopBackOff,
			RestartCount:          5,
		}})

		Expect(failure.Type).To(Equal(models.FailureTypeRuntimeCrash))
		Expect(failure.Container.FailureType).To(Equal(controllers.FailureTypeCrashLoop))
	})

	It("drops details stamped for a different release", func() {
		failure := controllers.MapLastFailureDetails(resourceName, "release-new", []corev1alpha1.LastFailureDetail{{
			ContainerName:         resourceName,
			LastTerminationReason: controllers.ReasonCrashLoopBackOff,
			ReleaseID:             "release-old",
		}})

		Expect(failure).To(BeNil())
	})

	It("keeps details stamped for the resolved release", func() {
		failure := controllers.MapLastFailureDetails(resourceName, "release-new", []corev1alpha1.LastFailureDetail{{
			ContainerName:         resourceName,
			LastTerminationReason: controllers.ReasonCrashLoopBackOff,
			ReleaseID:             "release-new",
		}})

		Expect(failure).NotTo(BeNil())
		Expect(failure.Container.FailureType).To(Equal(controllers.FailureTypeCrashLoop))
	})

	It("keeps unstamped details, for agents that predate the stamp", func() {
		failure := controllers.MapLastFailureDetails(resourceName, "release-new", []corev1alpha1.LastFailureDetail{{
			ContainerName:         resourceName,
			LastTerminationReason: controllers.ReasonCrashLoopBackOff,
		}})

		Expect(failure).NotTo(BeNil())
	})

	It("keeps stamped details when the CR carries no release annotation", func() {
		failure := controllers.MapLastFailureDetails(resourceName, "", []corev1alpha1.LastFailureDetail{{
			ContainerName:         resourceName,
			LastTerminationReason: controllers.ReasonCrashLoopBackOff,
			ReleaseID:             "release-old",
		}})

		Expect(failure).NotTo(BeNil())
	})

	It("filters stale details out of a mixed slice", func() {
		failure := controllers.MapLastFailureDetails(resourceName, "release-new", []corev1alpha1.LastFailureDetail{
			{Type: corev1alpha1.FailureTypeReadinessFailure, ContainerName: resourceName, ReleaseID: "release-old"},
			{ContainerName: resourceName, LastTerminationReason: controllers.ReasonCrashLoopBackOff, ReleaseID: "release-new"},
		})

		Expect(failure.Type).To(Equal(models.FailureTypeRuntimeCrash))
		Expect(failure.Container.FailureType).To(Equal(controllers.FailureTypeCrashLoop))
	})

	It("types a mixed slice as a readiness failure", func() {
		failure := controllers.MapLastFailureDetails(resourceName, "", []corev1alpha1.LastFailureDetail{
			{Type: corev1alpha1.FailureTypeRuntimeCrash, ContainerName: resourceName + "-init"},
			{Type: corev1alpha1.FailureTypeReadinessFailure, ContainerName: resourceName},
		})

		Expect(failure.Type).To(Equal(models.FailureTypeReadinessFailure))
	})
})
