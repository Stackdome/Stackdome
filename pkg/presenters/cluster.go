package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"k8s.io/utils/ptr"
)

func PresentCluster(cluster *models.Cluster) *openapi.Cluster {
	if cluster == nil {
		return nil
	}
	return &openapi.Cluster{
		Id:             &cluster.ID,
		Name:           &cluster.Name,
		OrganisationId: &cluster.OrganisationID,
		Default:        &cluster.Default,
		ClusterUrl:     &cluster.ClusterURL,
	}
}

func PresentClusterList(clusters []*models.Cluster) *openapi.ClusterList {
	result := make([]openapi.Cluster, 0, len(clusters))
	for _, c := range clusters {
		result = append(result, *PresentCluster(c))
	}

	return &openapi.ClusterList{
		Items: result,
		Total: ptr.To(int32(len(result))),
	}
}

func ConvertCluster(in *openapi.Cluster) *models.Cluster {
	if in == nil {
		return nil
	}
	return &models.Cluster{
		ID:             in.GetId(),
		Name:           in.GetName(),
		OrganisationID: in.GetOrganisationId(),
		ClusterURL:     in.GetClusterUrl(),
		ClusterCAData:  in.GetClusterCaData(),
		Token:          in.GetClusterSaToken(),
	}
}
