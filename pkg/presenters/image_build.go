package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"k8s.io/utils/ptr"
)

func PresentImageBuildList(in []*models.ImageBuild) []openapi.ImageBuild {
	result := make([]openapi.ImageBuild, len(in))
	for i, w := range in {
		result[i] = PresentImageBuild(w)
	}
	return result
}

func PresentImageBuild(in *models.ImageBuild) openapi.ImageBuild {
	return openapi.ImageBuild{
		Id:                &in.ID,
		Namespace:         &in.Namespace,
		StackId:           &in.StackID,
		StackResourceId:   in.StackResourceID,
		StackResourceName: in.StackResourceName,
		BuildContext:      presentSourceContext(in.Spec.SourceContext),
		SourceRevision:    presentSourceRevision(in.Spec.SourceRevision),
		ImageRepo:         presentImageRepo(in.Spec),
		Status:            presentImageBuildStatus(in.Status),
	}
}

func presentImageRepo(in models.BuildConfigSpec) string {
	if in.BuildImageRepository.UseInClusterRegistry {
		return "InClusterRegistry"
	}
	return in.BuildImageRepository.ExternalImageRef
}

func presentImageBuildStatus(status *models.ImageBuildStatus) *openapi.ImageBuildStatus {
	if status == nil {
		return nil
	}
	return &openapi.ImageBuildStatus{
		State:                  &status.State,
		Conditions:             presentConditions(status.Conditions),
		ImageUrl:               &status.ImageURL,
		BuildSourceRevision:    &status.BuildSourceRevision,
		LastBuildFailureDetail: presentBuildFailureDetail(status.LastBuildFailureDetail),
	}
}

func presentBuildFailureDetail(d *models.BuildFailureDetail) *openapi.BuildFailureDetail {
	if d == nil {
		return nil
	}
	return &openapi.BuildFailureDetail{
		FailureType:  ptr.To(d.FailureType),
		Reason:       ptr.To(d.Reason),
		Message:      ptr.To(d.Message),
		RestartCount: ptr.To(d.RestartCount),
		ExitCode:     d.ExitCode,
	}
}
