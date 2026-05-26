package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func PresentStackTopology(topology *models.StackTopology) *openapi.StackTopology {
	if topology == nil {
		return nil
	}

	return &openapi.StackTopology{
		Nodes: presentTopologyNodes(topology.Nodes),
		Edges: presentTopologyEdges(topology.Edges),
	}
}

func presentTopologyNodes(nodes []models.TopologyNode) []openapi.TopologyNode {
	result := make([]openapi.TopologyNode, len(nodes))
	for i, node := range nodes {
		result[i] = openapi.TopologyNode{
			Ref:   presentTopologyNodeRef(node.Ref),
			Label: node.Label,
		}
		if len(node.Outputs) > 0 {
			result[i].SetOutputs(presentOutputDescriptors(node.Outputs))
		}
		if node.State != "" {
			result[i].SetState(node.State)
		}
	}
	return result
}

func presentTopologyEdges(edges []models.TopologyEdge) []openapi.TopologyEdge {
	result := make([]openapi.TopologyEdge, len(edges))
	for i, edge := range edges {
		result[i] = openapi.TopologyEdge{
			Kind:          edge.Kind,
			Source:        presentTopologyNodeRef(edge.Source),
			Target:        presentTopologyNodeRef(edge.Target),
			SourceOfTruth: edge.SourceOfTruth,
		}
		if edge.Id != "" {
			result[i].SetId(edge.Id)
		}
		if len(edge.Mappings) > 0 {
			result[i].SetMappings(presentConnectionMappings(edge.Mappings))
		}
		if len(edge.Config) > 0 {
			result[i].SetConfig(edge.Config)
		}
	}
	return result
}

