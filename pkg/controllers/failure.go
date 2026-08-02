package controllers

import (
	"github.com/Stackdome/stackdome/pkg/models"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

// Termination reasons written by the kubelet, plus ReasonPortNotListening which
// the cluster agent writes when a declared port is proven closed. Neither
// exports these strings.
const (
	ReasonCrashLoopBackOff     = "CrashLoopBackOff"
	ReasonOOMKilled            = "OOMKilled"
	ReasonImagePullBackOff     = "ImagePullBackOff"
	ReasonErrImagePull         = "ErrImagePull"
	ReasonCreateContainerError = "CreateContainerError"
	ReasonPortNotListening     = "PortNotListening"
)

// Container-detail failure types exposed by the API.
const (
	FailureTypeCrashLoop            = "crash_loop"
	FailureTypeOutOfMemory          = "out_of_memory"
	FailureTypeImagePullFailed      = "image_pull_failed"
	FailureTypeCreateContainerError = "create_container_error"
	FailureTypeExitError            = "exit_error"
	FailureTypePortNotListening     = "port_not_listening"
)

func MapFailureType(reason string) string {
	switch reason {
	case ReasonCrashLoopBackOff:
		return FailureTypeCrashLoop
	case ReasonOOMKilled:
		return FailureTypeOutOfMemory
	case ReasonImagePullBackOff, ReasonErrImagePull:
		return FailureTypeImagePullFailed
	case ReasonCreateContainerError:
		return FailureTypeCreateContainerError
	case ReasonPortNotListening:
		return FailureTypePortNotListening
	default:
		return FailureTypeExitError
	}
}

func MapLastFailureDetails(resourceName string, details []corev1alpha1.LastFailureDetail) *models.StackResourceFailure {
	if len(details) == 0 {
		return nil
	}
	failure := &models.StackResourceFailure{Type: failureTypeForDetails(details)}
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

// failureTypeForDetails picks the server-side failure type from the agent's
// per-detail classification. An empty Type means runtime_crash: older agents do
// not set the field.
func failureTypeForDetails(details []corev1alpha1.LastFailureDetail) models.StackResourceFailureType {
	for _, d := range details {
		if d.Type == corev1alpha1.FailureTypeReadinessFailure {
			return models.FailureTypeReadinessFailure
		}
	}
	return models.FailureTypeRuntimeCrash
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
