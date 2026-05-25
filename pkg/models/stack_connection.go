package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
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
type ConnectionMappings []ConnectionMapping
type ConnectionConfig map[string]interface{}

type StackConnection struct {
	Id       string                 `json:"id,omitempty"`
	Kind     ConnectionKind         `json:"kind"`
	From     TopologyNodeRef        `json:"from"`
	To       TopologyNodeRef        `json:"to"`
	Mappings []ConnectionMapping    `json:"mappings,omitempty"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

type StackConnectionRecord struct {
	ID        string             `gorm:"primary_key;default:gen_random_uuid()"`
	StackID   string             `gorm:"not null;index"`
	Kind      ConnectionKind     `gorm:"not null"`
	FromRef   TopologyNodeRef    `gorm:"column:from_ref;type:jsonb;not null"`
	ToRef     TopologyNodeRef    `gorm:"column:to_ref;type:jsonb;not null"`
	Mappings  ConnectionMappings `gorm:"type:jsonb"`
	Config    ConnectionConfig   `gorm:"type:jsonb"`
	CreatedAt time.Time
	UpdatedAt time.Time
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

func (r StackConnectionRecord) ToStackConnection() StackConnection {
	return StackConnection{
		Id:       r.ID,
		Kind:     r.Kind,
		From:     r.FromRef,
		To:       r.ToRef,
		Mappings: []ConnectionMapping(r.Mappings),
		Config:   map[string]interface{}(r.Config),
	}
}

func NewStackConnectionRecord(stackID string, connection StackConnection) StackConnectionRecord {
	return StackConnectionRecord{
		ID:       connection.Id,
		StackID:  stackID,
		Kind:     connection.Kind,
		FromRef:  connection.From,
		ToRef:    connection.To,
		Mappings: ConnectionMappings(connection.Mappings),
		Config:   ConnectionConfig(connection.Config),
	}
}

func (StackConnectionRecord) TableName() string {
	return "stack_connections"
}

func (c *StackConnections) Scan(value interface{}) error {
	if value == nil {
		*c = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, c)
}

func (c StackConnections) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (r *TopologyNodeRef) Scan(value interface{}) error {
	if value == nil {
		*r = TopologyNodeRef{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, r)
}

func (r TopologyNodeRef) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (m *ConnectionMappings) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, m)
}

func (m ConnectionMappings) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (c *ConnectionConfig) Scan(value interface{}) error {
	if value == nil {
		*c = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, c)
}

func (c ConnectionConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}
