package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
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

// SecretReference points to a secret and optionally maps specific keys
type SecretReference struct {
	SecretID    string            `json:"secret_id"`              // UUID of the secret
	KeyMappings map[string]string `json:"key_mappings,omitempty"` // secret_key in db -> target_key mapping in k8s.
}
