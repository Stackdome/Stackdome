package models

import (
	"database/sql/driver"
	"encoding/json"
)

const (
	ImageRegistryIDLabel = "registry.stackdome.io/id"
)

type RegistryState string

const (
	RegistryStatePending RegistryState = "Pending"
	RegistryStateRunning RegistryState = "Running"
	RegistryStateError   RegistryState = "Error"
)

const (
	// DefaultRegistryStorageSize applies to registries created via the API
	// on an org's own cluster.
	DefaultRegistryStorageSize = "50Gi"
	// DefaultPlatformOrgRegistryStorageSize applies to the registry seeded
	// for each org at signup on the shared-compute cluster.
	DefaultPlatformOrgRegistryStorageSize = "10Gi"
)

// OrgRegistryDefaults are install-level storage defaults for the registry
// seeded for each org at signup.
type OrgRegistryDefaults struct {
	StorageSize  string
	StorageClass string
}

type ClusterImageRegistry struct {
	ID                  string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	ClusterID           string `gorm:"not null;index:idx_cluster_image_registry_cluster_id" json:"cluster_id"`
	OrganisationID      string `gorm:"not null;index:idx_cluster_image_registry_organisation_id" json:"organisation_id"`
	Name                string `gorm:"not null;check:name <> ''" json:"name"`
	BackendStorageSize  string `gorm:"not null;check:backend_storage_size <> ''" json:"backend_storage_size"`
	BackendStorageClass string
	Status              *ClusterImageRegistryStatus `gorm:"type:jsonb" json:"status"`
}

type ClusterImageRegistryStatus struct {
	State       RegistryState `json:"state"`
	Conditions  []Condition   `json:"conditions"`
	RegistryUrl string        `json:"registry_url"`
}

// Scan and value methods for ClusterImageRegistryStatus
func (c *ClusterImageRegistryStatus) Scan(value interface{}) error {
	if value == nil {
		*c = ClusterImageRegistryStatus{}
		return nil
	}
	return json.Unmarshal(value.([]byte), c)
}
func (c ClusterImageRegistryStatus) Value() (driver.Value, error) {
	if c.State == "" {
		return nil, nil
	}
	return json.Marshal(c)
}
