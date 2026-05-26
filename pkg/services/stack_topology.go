package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func (s *stackService) GetStackTopology(ctx context.Context, ID string) (*models.StackTopology, *errors.ServiceError) {
	stack, err := s.GetStack(ctx, ID)
	if err != nil {
		return nil, err
	}

	resourcesByName := make(map[string]*models.StackResource, len(stack.StackResources))
	for _, r := range stack.StackResources {
		resourcesByName[r.Name] = r
	}
	volumesByName := make(map[string]*models.Volume, len(stack.Volumes))
	for _, v := range stack.Volumes {
		volumesByName[v.Name] = v
	}

	nodes, edges := nodesAndEdgesFromConnections(stack, resourcesByName, volumesByName)

	if enrichErr := s.enrichExternalNodes(ctx, nodes); enrichErr != nil {
		return nil, enrichErr
	}

	for _, resource := range stack.StackResources {
		key := nodeKey(models.TopologyNodeTypeStackResource, resource.ID)
		if node, ok := nodes[key]; ok {
			node.Outputs = resource.EnsureDeclaredOutputs()
			if resource.Status != nil {
				node.State = string(resource.Status.State)
			}
			nodes[key] = node
		} else {
			n := models.TopologyNode{
				Ref:     models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Id: resource.ID, Name: resource.Name},
				Label:   resource.Name,
				Outputs: resource.EnsureDeclaredOutputs(),
			}
			if resource.Status != nil {
				n.State = string(resource.Status.State)
			}
			nodes[key] = n
		}

		for _, dep := range resource.DependsOn {
			depRef := models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: dep}
			if depResource, ok := resourcesByName[dep]; ok {
				depRef.Id = depResource.ID
			}
			edges = append(edges, models.TopologyEdge{
				Kind:          models.DependsOnTopologyEdgeKind,
				Source:        depRef,
				Target:        models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Id: resource.ID, Name: resource.Name},
				SourceOfTruth: "derived",
			})
		}
	}

	for _, volume := range stack.Volumes {
		key := nodeKey(models.TopologyNodeTypeVolume, volume.ID)
		if _, ok := nodes[key]; !ok {
			nodes[key] = models.TopologyNode{
				Ref:   models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Id: volume.ID, Name: volume.Name},
				Label: volume.Name,
			}
		}
	}

	nodeSlice := make([]models.TopologyNode, 0, len(nodes))
	for _, node := range nodes {
		nodeSlice = append(nodeSlice, node)
	}

	return &models.StackTopology{Nodes: nodeSlice, Edges: edges}, nil
}

func nodesAndEdgesFromConnections(
	stack *models.Stack,
	resourcesByName map[string]*models.StackResource,
	volumesByName map[string]*models.Volume,
) (map[string]models.TopologyNode, []models.TopologyEdge) {
	nodes := make(map[string]models.TopologyNode)
	edges := make([]models.TopologyEdge, 0, len(stack.Connections))

	for _, connection := range stack.Connections {
		fromRef := populateRef(connection.From, resourcesByName, volumesByName)
		toRef := populateRef(connection.To, resourcesByName, volumesByName)

		ensureNode(nodes, fromRef)
		ensureNode(nodes, toRef)

		edges = append(edges, models.TopologyEdge{
			Id:            connection.ID,
			Kind:          connection.Kind.String(),
			Source:        fromRef,
			Target:        toRef,
			Mappings:      connection.Mappings,
			Config:        connection.Config,
			SourceOfTruth: "connection",
		})
	}

	return nodes, edges
}

func populateRef(
	ref models.TopologyNodeRef,
	resourcesByName map[string]*models.StackResource,
	volumesByName map[string]*models.Volume,
) models.TopologyNodeRef {
	switch ref.Type {
	case models.TopologyNodeTypeStackResource:
		if r, ok := resourcesByName[ref.Name]; ok {
			ref.Id = r.ID
		}
	case models.TopologyNodeTypeVolume:
		if v, ok := volumesByName[ref.Name]; ok {
			ref.Id = v.ID
		}
	}
	return ref
}

func ensureNode(nodes map[string]models.TopologyNode, ref models.TopologyNodeRef) {
	key := nodeKey(ref.Type, ref.Id)
	if _, ok := nodes[key]; ok {
		return
	}
	label := ref.Name
	if label == "" {
		label = ref.Id
	}
	nodes[key] = models.TopologyNode{Ref: ref, Label: label}
}

func (s *stackService) enrichExternalNodes(ctx context.Context, nodes map[string]models.TopologyNode) *errors.ServiceError {
	for key, node := range nodes {
		switch node.Ref.Type {
		case models.TopologyNodeTypePostgresAddon:
			addon, err := s.postgresAddonService.InternalGetPostgresAddon(ctx, node.Ref.Id)
			if err != nil {
				if err.Is404() {
					s.logger.Warnf("postgres addon '%s' not found for topology, marking as missing", node.Ref.Id)
					node.State = "missing"
					nodes[key] = node
					continue
				}
				return errors.GeneralError("failed to fetch postgres addon '%s' for topology: %s", node.Ref.Id, err.Error())
			}
			node.Ref.Name = addon.Name
			node.Label = addon.Name
			node.Outputs = addon.EnsureDeclaredOutputs()
			if addon.Status.State != "" {
				node.State = string(addon.Status.State)
			}
			nodes[key] = node
		case models.TopologyNodeTypeSecret:
			secret, err := s.secretService.InternalGetByID(ctx, node.Ref.Id)
			if err != nil {
				if err.Is404() {
					s.logger.Warnf("secret '%s' not found for topology, marking as missing", node.Ref.Id)
					node.State = "missing"
					nodes[key] = node
					continue
				}
				return errors.GeneralError("failed to fetch secret '%s' for topology: %s", node.Ref.Id, err.Error())
			}
			node.Ref.Name = secret.Name
			node.Label = secret.Name
			node.Outputs = secret.EnsureDeclaredOutputs()
			nodes[key] = node
		}
	}
	return nil
}

func nodeKey(nodeType models.TopologyNodeType, id string) string {
	return string(nodeType) + ":" + id
}
