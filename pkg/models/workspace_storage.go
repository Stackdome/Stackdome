package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type WorkspaceStorageState string

const (
	WorkspaceStorageStateCreating WorkspaceStorageState = "Creating"
	WorkspaceStorageStateReady    WorkspaceStorageState = "Ready"
	WorkspaceStorageStateDeleting WorkspaceStorageState = "Deleting"
	WorkspaceStorageStateFailed   WorkspaceStorageState = "Failed"
	WorkspaceStorageStateCreated  WorkspaceStorageState = "Created"
	WorkspaceStorageStatePending  WorkspaceStorageState = "Pending"
)

type WorkspaceStorage struct {
	ID                string                  `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	OrganisationID    string                  `gorm:"not null"`
	UserID            string                  `gorm:"not null"`
	Name              string                  `gorm:"not null"`
	Namespace         string                  `gorm:"unique;not null"`
	Labels            Labels                  `gorm:"type:jsonb"`
	Annotations       Annotations             `gorm:"type:jsonb"`
	SSHConfig         *SSHConfig              `gorm:"type:jsonb"`
	Volumes           []Volume                `gorm:"foreignKey:WorkspaceStorageID"`
	Status            *WorkspaceStorageStatus `gorm:"type:jsonb"`
	State             WorkspaceStorageState   `gorm:"not null"`
	DeletionTimeStamp *time.Time              `json:"deletion_timestamp"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type SSHConfig struct {
	// Users public key.
	PublicKey string `json:"public_key"`
}

func (c SSHConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *SSHConfig) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte for SSHConfig failed")
	}
	return json.Unmarshal(b, &c)
}

type Volume struct {
	ID                 string               `gorm:"primaryKey" json:"id"`
	WorkspaceStorageID string               `gorm:"primaryKey" json:"workspace_storage_id"`
	Name               string               `gorm:"not null" json:"name"`
	Labels             Labels               `gorm:"type:jsonb" json:"labels"`
	Annotations        Annotations          `gorm:"type:jsonb" json:"annotations"`
	Size               string               `json:"size"`
	StorageClass       string               `json:"storage_class,omitempty"`
	LocalSource        *LocalSource         `gorm:"type:jsonb" json:"local_source,omitempty"`
	BuildSource        BuildArtifactSources `gorm:"type:jsonb" json:"build_source,omitempty"`
	SyncBeforeUse      bool                 `json:"sync_before_use"`
	VolumeStatus       *VolumeStatus        `gorm:"type:jsonb" json:"volume_status,omitempty"`
}

type BuildArtifactSources []BuildArtifactSource

func (b BuildArtifactSources) Value() (driver.Value, error) {
	return json.Marshal(b)
}

func (b *BuildArtifactSources) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte for BuildArtifactSources failed")
	}
	return json.Unmarshal(v, &b)
}

type VolumeStatus struct {
	ObservedGeneration int64                   `json:"observed_generation"`
	Conditions         []Condition             `json:"conditions"`
	Phase              string                  `json:"phase"`
	BuildArtifactSyncs []BuildArtifactSyncInfo `json:"build_artifact_syncs,omitempty"`
}

func (v VolumeStatus) Value() (driver.Value, error) {
	return json.Marshal(v)
}

func (v *VolumeStatus) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte for VolumeStatus failed")
	}
	return json.Unmarshal(b, &v)
}

type BuildArtifactSyncInfo struct {
	ResourceName string `json:"resource_name"`
	BuildID      string `json:"build_id"`
	Status       string `json:"status"`
}

type LocalSource struct {
	Path string `json:"path"`
	Sync bool   `json:"sync"`
}

func (l LocalSource) Value() (driver.Value, error) {
	return json.Marshal(l)
}

func (l *LocalSource) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte for LocalSource failed")
	}
	return json.Unmarshal(b, &l)
}

type BuildArtifactSource struct {
	ResourceName    string `json:"resource_name"`
	SourcePath      string `json:"source_path"`
	DestinationPath string `json:"destination_path"`
}

type WorkspaceStorageStatus struct {
	ObservedGeneration       int64       `json:"observed_generation"`
	Conditions               []Condition `json:"conditions"`
	Phase                    string      `json:"phase"`
	StorageServerServiceName string      `json:"storage_server_service_name"`
}

func (s WorkspaceStorageStatus) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *WorkspaceStorageStatus) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte for WorkspaceStorageStatus failed")
	}
	return json.Unmarshal(b, &s)
}
