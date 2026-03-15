package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type ObjectStore struct {
	ID             string `gorm:"primary_key;default:gen_random_uuid()"`
	OrganisationID string `gorm:"not null;index"`
	Name           string `gorm:"not null"`

	// Spec fields
	Configuration   ObjectStoreConfiguration `gorm:"type:jsonb;not null"`
	DestinationPath string                   `gorm:"not null"`
	RetentionPolicy string

	// Status fields
	Status    ObjectStoreStatus `gorm:"type:jsonb"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ObjectStoreConfiguration struct {
	S3Credentials    *S3Credentials    `json:"s3Credentials,omitempty"`
	AzureCredentials *AzureCredentials `json:"azureCredentials,omitempty"`
	GCSCredentials   *GCSCredentials   `json:"gcsCredentials,omitempty"`
}

type S3Credentials struct {
	AccessKeyID     SecretReference `json:"accessKeyId"`
	SecretAccessKey SecretReference `json:"secretAccessKey"`
	Region          string          `json:"region"`
	Endpoint        string          `json:"endpoint,omitempty"`
}

type AzureCredentials struct {
	ConnectionString   SecretReference `json:"connectionString"`
	StorageAccountName string          `json:"storageAccountName"`
}

type GCSCredentials struct {
	ServiceAccountCredentials SecretReference `json:"serviceAccountCredentials"`
}

type ObjectStoreStatus struct {
	State      string      `json:"state,omitempty"`
	Message    string      `json:"message,omitempty"`
	Conditions []Condition `json:"conditions,omitempty"`
}

// Implement driver.Valuer and sql.Scanner for custom types
func (c *ObjectStoreConfiguration) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &c)
}

func (c ObjectStoreConfiguration) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (s *ObjectStoreStatus) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &s)
}

func (s ObjectStoreStatus) Value() (driver.Value, error) {
	return json.Marshal(s)
}
