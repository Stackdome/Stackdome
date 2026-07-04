package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	RegistrySecretKey = "registry"
	UsernameSecretKey = "username"
	PasswordSecretKey = "password"
	SshSecretKey      = "ssh_private_key"
	TokenSecretKey    = "token"
)

type SecretType string

const (
	SecretTypeGeneric          SecretType = "Generic"
	SecretTypeDockerRegistry   SecretType = "DockerRegistry"
	SecretTypeGitCredentials   SecretType = "GitCredentials"
	SecretTypeUsernamePassword SecretType = "UsernamePassword"
	SecretTypeToken            SecretType = "Token"
	SecretTypeSSHKey           SecretType = "SSHKey"
)

const (
	SecretDataHashAnnotation = "stackdome.io/secret-data-hash"
	SecretIDAnnotation       = "stackdome.io/secret-id"
)

// ManagedSecretSlot identifies which credential slot of an owning resource a
// managed secret materializes.
type ManagedSecretSlot string

const (
	ManagedSecretSlotGit  ManagedSecretSlot = "git"
	ManagedSecretSlotPull ManagedSecretSlot = "pull"
	ManagedSecretSlotPush ManagedSecretSlot = "push"
)

// Kinds of resources that can own managed secrets.
const (
	ManagedByKindStackResource = "stack_resource"
	ManagedByKindPreviewConfig = "preview_config"
)

// ManagedSecretName returns the deterministic name for a managed secret
// materialized from inline credentials: managed-<stack>-<resource>-<slot>.
func ManagedSecretName(stackName, resourceName string, slot ManagedSecretSlot) string {
	return fmt.Sprintf("managed-%s-%s-%s", stackName, resourceName, slot)
}

// StackResourceManagedOwnerID is the managed_by_id for stack-resource-owned
// managed secrets. It uses the stack ID plus resource name so it is stable
// across resource re-creation within a stack.
func StackResourceManagedOwnerID(stackID, resourceName string) string {
	return fmt.Sprintf("%s/%s", stackID, resourceName)
}

type SecretKeys []string

func (s *SecretKeys) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &s)
}

func (s SecretKeys) Value() (driver.Value, error) {
	return json.Marshal(s)
}

type Secret struct {
	ID             string `gorm:"primary_key;default:gen_random_uuid()"`
	OrganisationID string `gorm:"not null;index"`
	TeamID         string `gorm:"index" json:"team_id"`
	UserID         string `gorm:"not null;index"`
	Name           string `gorm:"not null"`
	Description    string
	Type           SecretType `gorm:"not null" json:"type"`
	EncryptedData  string     `gorm:"type:text;not null"`
	Keys           SecretKeys `gorm:"type:jsonb" json:"keys"`
	DataHash       string     `gorm:"not null" ` // For change detection

	// Managed secrets are materialized from inline credentials and owned by
	// another resource; they are hidden from default secret listings.
	Managed       bool              `gorm:"not null;default:false"`
	ManagedByKind string            `gorm:"column:managed_by_kind"`
	ManagedByID   string            `gorm:"column:managed_by_id"`
	ManagedSlot   ManagedSecretSlot `gorm:"column:managed_slot"`

	CreatedAt time.Time
	UpdatedAt time.Time

	// Transient field for API responses (not stored in DB)
	Data    map[string]string  `gorm:"-" json:"data,omitempty"`
	Outputs []OutputDescriptor `gorm:"-" json:"outputs,omitempty"`
}

// SecretReference points to a secret and an optional key within that secret.
type SecretReference struct {
	SecretID string `json:"secret_id"` // UUID of the secret
	Key      string `json:"key"`       // Key within the secret
}

func (s *Secret) ClusterSecretName() string {
	switch s.Type {
	case SecretTypeDockerRegistry:
		return fmt.Sprintf("docker-registry-%s", s.Name)
	case SecretTypeGitCredentials:
		return fmt.Sprintf("git-credentials-%s", s.Name)
	case SecretTypeUsernamePassword:
		return fmt.Sprintf("username-password-%s", s.Name)
	case SecretTypeToken:
		return fmt.Sprintf("token-%s", s.Name)
	case SecretTypeSSHKey:
		return fmt.Sprintf("ssh-key-%s", s.Name)
	default:
		return fmt.Sprintf("generic-secret-%s", s.Name)
	}
}
