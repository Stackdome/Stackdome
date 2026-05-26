package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type ConnectionKind string

const (
	ConnectionKindEnv                 ConnectionKind = "env"
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
	Id       string              `json:"id,omitempty"`
	Kind     ConnectionKind      `json:"kind"`
	From     TopologyNodeRef     `json:"from"`
	To       TopologyNodeRef     `json:"to"`
	Mappings []ConnectionMapping `json:"mappings,omitempty"`
	Config   ConnectionConfig    `json:"config,omitempty"`
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
	// Used when Type is file, represents the path where the file should be mounted inside the container
	Path string `json:"path,omitempty"`
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

func (c StackConnection) IsKind(kind ConnectionKind) bool {
	return c.Kind == kind
}

func (c StackConnection) SourceIs(nodeType TopologyNodeType) bool {
	return c.From.Type == nodeType
}

func (c StackConnection) TargetIs(nodeType TopologyNodeType) bool {
	return c.To.Type == nodeType
}

func (c StackConnection) TargetsStackResource(resourceName string) bool {
	return c.TargetIs(TopologyNodeTypeStackResource) &&
		c.To.Name == resourceName
}

func (c StackConnection) ConfigString(key string) (string, bool, error) {
	value, ok := c.Config[key]
	if !ok {
		return "", false, nil
	}
	asString, ok := value.(string)
	if !ok {
		return "", false, fmt.Errorf("config.%s must be a string", key)
	}
	return asString, true, nil
}

func (c StackConnection) RequiredConfigString(key string) (string, error) {
	value, _, err := c.ConfigString(key)
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", fmt.Errorf("config.%s is required", key)
	}
	return value, nil
}

func (c StackConnection) ConfigBool(key string) (bool, bool, error) {
	value, ok := c.Config[key]
	if !ok {
		return false, false, nil
	}
	asBool, ok := value.(bool)
	if !ok {
		return false, false, fmt.Errorf("config.%s must be a boolean", key)
	}
	return asBool, true, nil
}

func (connections StackConnections) HasAnyKind(kinds ...ConnectionKind) bool {
	for _, connection := range connections {
		for _, kind := range kinds {
			if connection.IsKind(kind) {
				return true
			}
		}
	}
	return false
}

func (connections StackConnections) OfKind(kind ConnectionKind) StackConnections {
	result := make(StackConnections, 0, len(connections))
	for _, connection := range connections {
		if connection.IsKind(kind) {
			result = append(result, connection)
		}
	}
	return result
}

func (connections StackConnections) EnvForStackResource(resourceName string) StackConnections {
	result := make(StackConnections, 0, len(connections))
	for _, connection := range connections {
		if connection.IsKind(ConnectionKindEnv) && connection.TargetsStackResource(resourceName) {
			result = append(result, connection)
		}
	}
	return result
}

func (connections StackConnections) FromType(sourceType TopologyNodeType) StackConnections {
	result := make(StackConnections, 0, len(connections))
	for _, connection := range connections {
		if connection.SourceIs(sourceType) {
			result = append(result, connection)
		}
	}
	return result
}

func (r StackConnectionRecord) ToStackConnection() StackConnection {
	return StackConnection{
		Id:       r.ID,
		Kind:     r.Kind,
		From:     r.FromRef,
		To:       r.ToRef,
		Mappings: []ConnectionMapping(r.Mappings),
		Config:   r.Config,
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
		Config:   connection.Config,
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
