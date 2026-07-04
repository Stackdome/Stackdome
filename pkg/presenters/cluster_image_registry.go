package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"k8s.io/utils/ptr"
)

func PresentClusterImageRegistryList(clusterImageRegistries []*models.ClusterImageRegistry) *openapi.ClusterImageRegistryList {
	result := make([]openapi.ClusterImageRegistry, 0, len(clusterImageRegistries))
	for _, c := range clusterImageRegistries {
		curr := PresentClusterImageRegistry(c)
		if curr != nil {
			result = append(result, *curr)
		}
	}
	return &openapi.ClusterImageRegistryList{
		Items: result,
		Total: ptr.To(int32(len(result))),
	}
}

func PresentClusterImageRegistry(clusterImageRegistry *models.ClusterImageRegistry) *openapi.ClusterImageRegistry {
	if clusterImageRegistry == nil {
		return nil
	}
	res := &openapi.ClusterImageRegistry{
		Id:             &clusterImageRegistry.ID,
		Name:           clusterImageRegistry.Name,
		OrganisationId: &clusterImageRegistry.OrganisationID,
		Spec:           PresentClusterImageRegistrySpec(clusterImageRegistry),
		Status:         PresentClusterImageRegistryStatus(clusterImageRegistry.Status),
	}

	return res
}

func PresentClusterImageRegistrySpec(clusterImageRegistry *models.ClusterImageRegistry) *openapi.ClusterImageRegistrySpec {
	if clusterImageRegistry == nil {
		return nil
	}
	res := &openapi.ClusterImageRegistrySpec{
		BackendStorageSize: &clusterImageRegistry.BackendStorageSize,
	}

	if clusterImageRegistry.BackendStorageClass != "" {
		res.BackendStorageClass = &clusterImageRegistry.BackendStorageClass
	}
	return res
}

func PresentClusterImageRegistryStatus(status *models.ClusterImageRegistryStatus) *openapi.ClusterImageRegistryStatus {
	if status == nil {
		return nil
	}
	return &openapi.ClusterImageRegistryStatus{
		State:      PresentClusterImageRegistryState(status.State),
		Conditions: presentConditions(status.Conditions),
	}
}

func PresentClusterImageRegistryState(state models.RegistryState) *openapi.ClusterImageRegistryState {
	switch state {
	case models.RegistryStatePending:
		return openapi.IMAGE_REGISTRY_PENDING.Ptr()
	case models.RegistryStateRunning:
		return openapi.IMAGE_REGISTRY_RUNNING.Ptr()
	case models.RegistryStateError:
		return openapi.IMAGE_REGISTRY_ERROR.Ptr()
	default:
		return openapi.IMAGE_REGISTRY_PENDING.Ptr()
	}
}

func ConvertClusterImageRegistry(in *openapi.ClusterImageRegistry) *models.ClusterImageRegistry {
	if in == nil {
		return nil
	}
	res := &models.ClusterImageRegistry{
		Name: in.GetName(),
	}
	if in.HasSpec() {
		if in.Spec.HasBackendStorageSize() {
			res.BackendStorageSize = in.Spec.GetBackendStorageSize()
		}
		if in.Spec.HasBackendStorageClass() {
			res.BackendStorageClass = in.Spec.GetBackendStorageClass()
		}
	}
	return res
}
