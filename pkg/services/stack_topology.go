package services

import "github.com/ashishmax31/stackdome-api-server/pkg/models"

func buildStackTopology(
	stack *models.Stack,
	addonsByID map[string]*models.PostgresAddon,
	secretsByID map[string]*models.Secret,
) models.StackTopology {
	nodes := make([]models.TopologyNode, 0, len(stack.StackResources)+len(stack.Volumes)+len(addonsByID)+len(secretsByID))
	edges := make([]models.TopologyEdge, 0, len(stack.Connections))

	for _, resource := range stack.StackResources {
		state := ""
		if resource.Status != nil {
			state = string(resource.Status.State)
		}
		nodes = append(nodes, models.TopologyNode{
			Ref:     models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: resource.Name},
			Label:   resource.Name,
			Outputs: resource.EnsureDeclaredOutputs(),
			State:   state,
		})

		for _, dependency := range resource.DependsOn {
			edges = append(edges, models.TopologyEdge{
				Kind:          "depends_on",
				Source:        models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: dependency},
				Target:        models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: resource.Name},
				SourceOfTruth: "depends_on",
			})
		}
	}

	for _, volume := range stack.Volumes {
		nodes = append(nodes, models.TopologyNode{
			Ref:   models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Name: volume.Name},
			Label: volume.Name,
		})
	}

	for _, addon := range addonsByID {
		state := ""
		if addon != nil {
			state = string(addon.Status.State)
		}
		nodes = append(nodes, models.TopologyNode{
			Ref:     models.TopologyNodeRef{Type: models.TopologyNodeTypePostgresAddon, Id: addon.ID},
			Label:   addon.Name,
			Outputs: addon.EnsureDeclaredOutputs(),
			State:   state,
		})
	}

	for _, secret := range secretsByID {
		nodes = append(nodes, models.TopologyNode{
			Ref:     models.TopologyNodeRef{Type: models.TopologyNodeTypeSecret, Id: secret.ID},
			Label:   secret.Name,
			Outputs: secret.EnsureDeclaredOutputs(),
		})
	}

	for _, connection := range stack.Connections {
		edges = append(edges, models.TopologyEdge{
			Id:            connection.Id,
			Kind:          connection.Kind.String(),
			Source:        connection.From,
			Target:        connection.To,
			Mappings:      connection.Mappings,
			Config:        connection.Config,
			SourceOfTruth: "connection",
		})
	}

	return models.StackTopology{
		Nodes: nodes,
		Edges: edges,
	}
}
