package controllers

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

func MapFailureType(reason string) string {
	switch reason {
	case "CrashLoopBackOff":
		return "crash_loop"
	case "OOMKilled":
		return "out_of_memory"
	case "ImagePullBackOff", "ErrImagePull":
		return "image_pull_failed"
	case "CreateContainerError":
		return "create_container_error"
	default:
		return "exit_error"
	}
}

func MapLastFailureDetails(resourceName string, details []corev1alpha1.LastFailureDetail) *models.StackResourceFailure {
	if len(details) == 0 {
		return nil
	}
	failure := &models.StackResourceFailure{Type: models.FailureTypeRuntimeCrash}
	initName := resourceName + "-init"
	for _, d := range details {
		fd := mapContainerFailureDetail(d)
		switch d.ContainerName {
		case resourceName:
			failure.Container = fd
		case initName:
			failure.InitContainer = fd
		}
	}
	return failure
}

func MapBuildFailureDetail(d *corev1alpha1.LastFailureDetail) *models.BuildFailureDetail {
	if d == nil {
		return nil
	}
	return &models.BuildFailureDetail{
		FailureType:  MapFailureType(d.LastTerminationReason),
		Reason:       d.LastTerminationReason,
		Message:      d.LastTerminationMessage,
		RestartCount: d.RestartCount,
		ExitCode:     d.LastTerminationExitCode,
	}
}

func mapContainerFailureDetail(d corev1alpha1.LastFailureDetail) *models.ContainerFailureDetail {
	return &models.ContainerFailureDetail{
		FailureType:  MapFailureType(d.LastTerminationReason),
		Reason:       d.LastTerminationReason,
		Message:      d.LastTerminationMessage,
		RestartCount: d.RestartCount,
		ExitCode:     d.LastTerminationExitCode,
	}
}
