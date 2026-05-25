package models

type StackTopology struct {
	Nodes []TopologyNode `json:"nodes"`
	Edges []TopologyEdge `json:"edges"`
}

type TopologyNode struct {
	Ref     TopologyNodeRef    `json:"ref"`
	Label   string             `json:"label"`
	Outputs []OutputDescriptor `json:"outputs,omitempty"`
	State   string             `json:"state,omitempty"`
}

type TopologyEdge struct {
	Id            string                 `json:"id,omitempty"`
	Kind          string                 `json:"kind"`
	Source        TopologyNodeRef        `json:"source"`
	Target        TopologyNodeRef        `json:"target"`
	Mappings      []ConnectionMapping    `json:"mappings,omitempty"`
	Config        map[string]interface{} `json:"config,omitempty"`
	SourceOfTruth string                 `json:"source_of_truth"`
}
