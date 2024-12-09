package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

const (
	WorkspaceStorageIDLabel = "workspacestorage.stackdome.io/id"
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
	ID             string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	OrganisationID string `gorm:"not null"`
	UserID         string `gorm:"not null"`
	Name           string `gorm:"not null; <-:create"`
	Namespace      string `gorm:"unique;not null;  <-:create"`
	WorkspaceName  string `gorm:"<-:create"`
	// Tracks the version of the object in the database.
	Version           int
	Labels            Labels                  `gorm:"type:jsonb"`
	Annotations       Annotations             `gorm:"type:jsonb"`
	SSHConfig         *SSHConfig              `gorm:"type:jsonb"`
	Volumes           []*Volume               `gorm:"foreignKey:WorkspaceStorageID"`
	Status            *WorkspaceStorageStatus `gorm:"type:jsonb"`
	DeletionTimeStamp *time.Time              `json:"deletion_timestamp"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (w *WorkspaceStorage) SetState(state WorkspaceStorageState) bool {
	if w.Status.State == state {
		return false
	}
	w.Status.State = state
	return true
}

func (w *WorkspaceStorage) VolumeMap() map[string]*Volume {
	volumeMap := make(map[string]*Volume)
	for i := range w.Volumes {
		volumeMap[w.Volumes[i].ID] = w.Volumes[i]
	}
	return volumeMap
}

func (w *WorkspaceStorage) VolumeExists(volumeID string) bool {
	_, exists := w.VolumeMap()[volumeID]
	return exists
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
	ID                 string        `gorm:"primaryKey; <-:create" json:"id"`
	WorkspaceStorageID string        `gorm:"primaryKey" json:"workspace_storage_id"`
	Name               string        `gorm:"not null; <-:create" json:"name"`
	Labels             Labels        `gorm:"type:jsonb" json:"labels"`
	Annotations        Annotations   `gorm:"type:jsonb" json:"annotations"`
	Size               string        `gorm:"<-:create" json:"size"`
	StorageClass       string        `gorm:"<-:create" json:"storage_class,omitempty"`
	VolumeSource       *VolumeSource `gorm:"type:jsonb" json:"volume_source,omitempty"`
	SyncBeforeUse      bool          `json:"sync_before_use"`
	Status             *VolumeStatus `gorm:"type:jsonb" json:"volume_status,omitempty"`
}

func (v *Volume) VolumeSourceType() SourceVolumeType {
	switch {
	case v.VolumeSource == nil:
		return EmptyVolume
	case v.VolumeSource.BuildSource != nil:
		return BuildArtifactSyncedVolume
	case v.VolumeSource.LocalSource != nil:
		return LocalSyncedVolume
	default:
		return EmptyVolume
	}
}

type VolumeSource struct {
	LocalSource *LocalSource         `json:"local_source,omitempty"`
	BuildSource BuildArtifactSources `json:"build_source,omitempty"`
}

func (v VolumeSource) Value() (driver.Value, error) {
	return json.Marshal(v)
}

func (v *VolumeSource) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte for VolumeSource failed")
	}
	return json.Unmarshal(b, &v)
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
	ObservedGeneration     int64                   `json:"observed_generation"`
	Conditions             []Condition             `json:"conditions"`
	Phase                  string                  `json:"phase"`
	BuildArtifactSyncs     []BuildArtifactSyncInfo `json:"build_artifact_syncs,omitempty"`
	LastObservedStatusHash string                  `json:"last_observed_status_hash,omitempty"`
	InUse                  bool                    `json:"in_use"`
	LastSyncedAt           *time.Time              `json:"last_synced_at,omitempty"`
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
	State                    WorkspaceStorageState `json:"state"`
	Conditions               []Condition           `json:"conditions"`
	Phase                    string                `json:"phase"`
	StorageServerServiceName string                `json:"storage_server_service_name"`
	LastObservedStatusHash   string                `json:"last_observed_status_hash,omitempty"`
	ObservedVersion          int64                 `json:"observed_version"`
	// Subpath within the storage pod where different volumes are mounted.
	VolumeMountPaths []VolumeMountPath `json:"volume_mount_paths,omitempty"`
}

type VolumeMountPath struct {
	VolumeID string `json:"volume_id"`
	Path     string `json:"path"`
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
