package services

import (
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func TestBuildStackTopologyIncludesExplicitAndDerivedEdges(t *testing.T) {
	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{
				Name:      "api",
				DependsOn: models.Dependencies{"postgres"},
				Ports: models.Ports{
					{Name: "http", Number: 8080, Protocol: "http", ExposedToPublic: true},
				},
				Status: &models.StackResourceStatus{State: models.StackResourcePhaseReady},
			},
			{
				Name:   "worker",
				Status: &models.StackResourceStatus{State: models.StackResourcePhasePending},
			},
		},
		Volumes: []*models.Volume{
			{Name: "uploads"},
		},
		Connections: models.StackConnections{
			{
				Id:   "pg-api",
				Kind: models.ConnectionKindEnv,
				From: models.TopologyNodeRef{Type: models.TopologyNodeTypePostgresAddon, Id: "pg-1"},
				To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "api"},
				Mappings: []models.ConnectionMapping{
					{
						Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "DATABASE_URL"},
						Value:  models.ValueRef{Output: "url"},
					},
				},
			},
			{
				Id:   "tls-api",
				Kind: models.ConnectionKindEnv,
				From: models.TopologyNodeRef{Type: models.TopologyNodeTypeSecret, Id: "sec-1"},
				To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "api"},
				Mappings: []models.ConnectionMapping{
					{
						Target: models.ConnectionTarget{Type: models.ConnectionTargetTypeEnv, Name: "TLS_CERT"},
						Value:  models.ValueRef{Output: "key['tls.crt']"},
					},
				},
			},
		},
	}

	topology := buildStackTopology(stack, map[string]*models.PostgresAddon{
		"pg-1": {ID: "pg-1", Name: "postgres"},
	}, map[string]*models.Secret{
		"sec-1": {ID: "sec-1", Name: "tls", Keys: []string{"tls.crt"}},
	})

	if len(topology.Nodes) != 5 {
		t.Fatalf("expected 5 nodes, got %d", len(topology.Nodes))
	}
	if len(topology.Edges) != 3 {
		t.Fatalf("expected 3 edges, got %d", len(topology.Edges))
	}

	nodesByLabel := make(map[string]models.TopologyNode)
	for _, node := range topology.Nodes {
		nodesByLabel[node.Label] = node
	}
	if _, ok := nodesByLabel["api"]; !ok {
		t.Fatalf("expected api node")
	}
	if _, ok := nodesByLabel["postgres"]; !ok {
		t.Fatalf("expected postgres node")
	}
	if _, ok := nodesByLabel["tls"]; !ok {
		t.Fatalf("expected tls node")
	}
	if _, ok := nodesByLabel["uploads"]; !ok {
		t.Fatalf("expected uploads volume node")
	}

	var foundDependsOn bool
	for _, edge := range topology.Edges {
		if edge.SourceOfTruth == "depends_on" {
			foundDependsOn = true
			if edge.Kind != "depends_on" {
				t.Fatalf("expected depends_on edge kind, got %q", edge.Kind)
			}
			if edge.Source.Name != "postgres" || edge.Target.Name != "api" {
				t.Fatalf("unexpected depends_on edge: %#v", edge)
			}
		}
	}
	if !foundDependsOn {
		t.Fatalf("expected derived depends_on edge")
	}
}
