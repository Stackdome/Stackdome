package presenters_test

import (
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
)

func TestPresentStackTopology(t *testing.T) {
	topology := &models.StackTopology{
		Nodes: []models.TopologyNode{
			{
				Ref:     models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "api"},
				Label:   "api",
				State:   "Ready",
				Outputs: []models.OutputDescriptor{{Name: "url.http", Type: models.OutputValueTypeString}},
			},
		},
		Edges: []models.TopologyEdge{
			{
				Id:            "api-worker",
				Kind:          "env",
				Source:        models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "api"},
				Target:        models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "worker"},
				SourceOfTruth: "connection",
				Mappings: []models.ConnectionMapping{
					{
						Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "API_URL"},
						Value:  models.ValueRef{Output: "url.http"},
					},
				},
			},
		},
	}

	out := presenters.PresentStackTopology(topology)

	if len(out.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(out.Nodes))
	}
	if out.Nodes[0].Label != "api" {
		t.Fatalf("expected api label, got %q", out.Nodes[0].Label)
	}
	if out.Nodes[0].GetState() != "Ready" {
		t.Fatalf("expected Ready state, got %q", out.Nodes[0].GetState())
	}
	if len(out.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(out.Edges))
	}
	if out.Edges[0].GetSourceOfTruth() != "connection" {
		t.Fatalf("expected connection source_of_truth, got %q", out.Edges[0].GetSourceOfTruth())
	}
	if out.Edges[0].Mappings[0].Value.GetOutput() != "url.http" {
		t.Fatalf("expected mapping output url.http, got %q", out.Edges[0].Mappings[0].Value.GetOutput())
	}
}
