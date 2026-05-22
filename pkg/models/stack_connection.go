package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type ConnectionKind string

const (
	ConnectionKindEnv                 ConnectionKind = "env"
	ConnectionKindSecretMount         ConnectionKind = "secret_mount"
	ConnectionKindVolumeMount         ConnectionKind = "volume_mount"
	ConnectionKindBuildArtifactSource ConnectionKind = "build_artifact_source"
)

type TopologyNodeType string

const (
	TopologyNodeTypeStackResource TopologyNodeType = "stack_resource"
	TopologyNodeTypePostgresAddon TopologyNodeType = "addon/postgres"
	TopologyNodeTypeSecret        TopologyNodeType = "secret"
	TopologyNodeTypeVolume        TopologyNodeType = "volume"
	TopologyNodeTypeObjectStore   TopologyNodeType = "object_store"
)

type ConnectionTargetType string

const (
	ConnectionTargetTypeEnv  ConnectionTargetType = "env"
	ConnectionTargetTypeFile ConnectionTargetType = "file"
)

type StackConnections []StackConnection

type StackConnection struct {
	Id       string                 `json:"id,omitempty"`
	Kind     ConnectionKind         `json:"kind"`
	From     TopologyNodeRef        `json:"from"`
	To       TopologyNodeRef        `json:"to"`
	Mappings []ConnectionMapping    `json:"mappings,omitempty"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

type TopologyNodeRef struct {
	Type TopologyNodeType `json:"type"`
	Id   string           `json:"id,omitempty"`
	Name string           `json:"name,omitempty"`
}

type ConnectionMapping struct {
	Target ConnectionTarget `json:"target"`
	Value  ValueRef         `json:"value"`
}

type ConnectionTarget struct {
	Type ConnectionTargetType `json:"type"`
	Name string               `json:"name,omitempty"`
	Path string               `json:"path,omitempty"`
}

type ValueRef struct {
	Output   string                    `json:"output,omitempty"`
	Template string                    `json:"template,omitempty"`
	Values   map[string]OutputValueRef `json:"values,omitempty"`
}

type OutputValueRef struct {
	Output string `json:"output"`
}

func (k ConnectionKind) String() string {
	return string(k)
}

func (c *StackConnections) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &c)
}

func (c StackConnections) Value() (driver.Value, error) {
	return json.Marshal(c)
}
