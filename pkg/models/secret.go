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
	UserID         string `gorm:"not null;index"`
	Name           string `gorm:"not null"`
	Description    string
	Type           SecretType `gorm:"not null" json:"type"`
	EncryptedData  string     `gorm:"type:text;not null"`
	Keys           SecretKeys `gorm:"type:jsonb" json:"keys"`
	DataHash       string     `gorm:"not null" ` // For change detection
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Transient field for API responses (not stored in DB)
	Data map[string]string `gorm:"-" json:"data,omitempty"`
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
