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

type ConnectionConfigKey string

const (
	// Postgres addon config keys
	ConnectionConfigKeyDatabase        ConnectionConfigKey = "database"
	ConnectionConfigKeyCredentialScope ConnectionConfigKey = "credential_scope"
	ConnectionConfigKeySuperuser       ConnectionConfigKey = "superuser"

	// Volume mount config keys
	ConnectionConfigKeyMountPath ConnectionConfigKey = "mount_path"
	ConnectionConfigKeySubPath   ConnectionConfigKey = "sub_path"
	ConnectionConfigKeyReadOnly  ConnectionConfigKey = "read_only"

	// Build artifact source config keys
	ConnectionConfigKeySourcePath      ConnectionConfigKey = "source_path"
	ConnectionConfigKeyDestinationPath ConnectionConfigKey = "destination_path"
)

// Postgres credential_scope values
const (
	CredentialScopeOwner     = "owner"
	CredentialScopeSuperuser = "superuser"
)

type ConnectionTargetType string

const (
	ConnectionTargetTypeEnv  ConnectionTargetType = "env"
	ConnectionTargetTypeFile ConnectionTargetType = "file"
)

type StackConnections []StackConnection
type ConnectionMappings []ConnectionMapping
type ConnectionConfig map[string]interface{}

// StackConnection represents a directed edge between two topology nodes in a stack.
//
// The unique constraint on (stack_id, kind, from_ref, to_ref, discriminator) prevents
// duplicate edges. The Discriminator field is a config-derived scoping key that allows
// multiple connections between the same (from, to) pair when they target different
// logical sub-resources — e.g., two env connections from the same PostgresAddon to the
// same stack resource, each targeting a different database. For connection types that
// don't need sub-resource scoping, the discriminator is an empty string and the
// constraint behaves like a simple (stack_id, kind, from_ref, to_ref) unique index.
//
// To extend discriminator support for new connection types, add a case to
// ComputeDiscriminator(). The store layer calls it automatically before persistence.
type StackConnection struct {
	ID            string             `json:"id" gorm:"column:id;primary_key;default:gen_random_uuid()"`
	StackID       string             `json:"-" gorm:"not null;index"`
	Kind          ConnectionKind     `json:"kind" gorm:"not null"`
	From          TopologyNodeRef    `json:"from" gorm:"column:from_ref;type:jsonb;not null"`
	To            TopologyNodeRef    `json:"to" gorm:"column:to_ref;type:jsonb;not null"`
	Mappings      ConnectionMappings `json:"mappings,omitempty" gorm:"type:jsonb"`
	Config        ConnectionConfig   `json:"config,omitempty" gorm:"type:jsonb"`
	Discriminator string             `json:"-" gorm:"not null;default:''"`
	CreatedAt     time.Time          `json:"-"`
	UpdatedAt     time.Time          `json:"-"`
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

// ComputeDiscriminator derives a scoping key from the connection's config based on
// the source type and kind. This allows the unique index to permit multiple connections
// between the same (from, to) pair when they address different sub-resources.
//
// Currently supported:
//   - addon/postgres + env → config.database (allows wiring multiple databases from one addon)
//
// Returns "" for all other connection types, preserving standard edge uniqueness.
func (c *StackConnection) ComputeDiscriminator() string {
	if c.Kind == ConnectionKindEnv && c.From.Type == TopologyNodeTypePostgresAddon {
		if db, ok := c.Config[string(ConnectionConfigKeyDatabase)].(string); ok && db != "" {
			return db
		}
	}
	return ""
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

func (StackConnection) TableName() string {
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
