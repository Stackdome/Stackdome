package services

import (
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func TestNodesAndEdgesFromConnectionsDrivesTopology(t *testing.T) {
	resources := []*models.StackResource{
		{ID: "res-api", Name: "api"},
		{ID: "res-worker", Name: "worker"},
	}
	volumes := []*models.Volume{
		{ID: "vol-1", Name: "uploads"},
	}

	resourcesByName := map[string]*models.StackResource{
		"api":    resources[0],
		"worker": resources[1],
	}
	volumesByName := map[string]*models.Volume{
		"uploads": volumes[0],
	}

	stack := &models.Stack{
		ID:             "stack-1",
		StackResources: resources,
		Volumes:        volumes,
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
			{
				Id:   "vol-api",
				Kind: models.ConnectionKindVolumeMount,
				From: models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Name: "uploads"},
				To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "api"},
				Config: models.ConnectionConfig{"mount_path": "/uploads"},
			},
		},
	}

	nodes, edges := nodesAndEdgesFromConnections(stack, resourcesByName, volumesByName)

	if len(nodes) != 4 {
		t.Fatalf("expected 4 nodes from connections (pg, secret, volume, stack_resource), got %d", len(nodes))
	}
	if len(edges) != 3 {
		t.Fatalf("expected 3 edges from connections, got %d", len(edges))
	}

	pgKey := nodeKey(models.TopologyNodeTypePostgresAddon, "pg-1")
	if node, ok := nodes[pgKey]; !ok {
		t.Fatalf("expected postgres addon node from connection ref")
	} else if node.Ref.Id != "pg-1" {
		t.Fatalf("expected postgres node ref.Id 'pg-1', got '%s'", node.Ref.Id)
	}

	secretKey := nodeKey(models.TopologyNodeTypeSecret, "sec-1")
	if node, ok := nodes[secretKey]; !ok {
		t.Fatalf("expected secret node from connection ref")
	} else if node.Ref.Id != "sec-1" {
		t.Fatalf("expected secret node ref.Id 'sec-1', got '%s'", node.Ref.Id)
	}

	volumeKey := nodeKey(models.TopologyNodeTypeVolume, "vol-1")
	if node, ok := nodes[volumeKey]; !ok {
		t.Fatalf("expected volume node from connection ref")
	} else if node.Ref.Name != "uploads" || node.Ref.Id != "vol-1" {
		t.Fatalf("expected volume node with name 'uploads' and id 'vol-1', got name='%s' id='%s'", node.Ref.Name, node.Ref.Id)
	}

	apiKey := nodeKey(models.TopologyNodeTypeStackResource, "res-api")
	if node, ok := nodes[apiKey]; !ok {
		t.Fatalf("expected stack_resource node from connection ref")
	} else if node.Ref.Name != "api" || node.Ref.Id != "res-api" {
		t.Fatalf("expected api node with name 'api' and id 'res-api', got name='%s' id='%s'", node.Ref.Name, node.Ref.Id)
	}

	var foundVolumeMount bool
	for _, edge := range edges {
		if edge.Kind == "volume_mount" {
			foundVolumeMount = true
			if edge.Source.Id != "vol-1" || edge.Target.Id != "res-api" {
				t.Fatalf("expected volume_mount edge with populated IDs, got source.Id='%s' target.Id='%s'", edge.Source.Id, edge.Target.Id)
			}
		}
	}
	if !foundVolumeMount {
		t.Fatalf("expected volume_mount edge from connections")
	}
}
