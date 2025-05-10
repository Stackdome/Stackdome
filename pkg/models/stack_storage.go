package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

const (
	VolumeIDLabel = "volume.stackdome.io/id"
)

type StackStorageState string

const (
	StackStorageStateCreating StackStorageState = "Creating"
	StackStorageStateReady    StackStorageState = "Ready"
	StackStorageStateDeleting StackStorageState = "Deleting"
	StackStorageStateFailed   StackStorageState = "Failed"
	StackStorageStateCreated  StackStorageState = "Created"
	StackStorageStatePending  StackStorageState = "Pending"
)

type StackStorage struct {
	ID             string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	OrganisationID string `gorm:"not null"`
	UserID         string `gorm:"not null"`
	Name           string `gorm:"not null; <-:create"`
	Namespace      string `gorm:"unique;not null;  <-:create"`
	WorkspaceName  string `gorm:"<-:create"`
	// Tracks the version of the object in the database.
	Version           int
	Labels            Labels              `gorm:"type:jsonb"`
	Annotations       Annotations         `gorm:"type:jsonb"`
	SSHConfig         SSHConfig           `gorm:"type:jsonb"`
	Volumes           []*Volume           `gorm:"foreignKey:StorageID"`
	Status            *StackStorageStatus `gorm:"type:jsonb"`
	DeletionTimeStamp *time.Time          `json:"deletion_timestamp"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (w *StackStorage) SetState(state StackStorageState) bool {
	if w.Status.State == state {
		return false
	}
	w.Status.State = state
	return true
}

func (w *StackStorage) VolumeMap() map[string]*Volume {
	volumeMap := make(map[string]*Volume)
	for i := range w.Volumes {
		volumeMap[w.Volumes[i].ID] = w.Volumes[i]
	}
	return volumeMap
}

func (w *StackStorage) VolumeExists(volumeID string) bool {
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

type StackStorageStatus struct {
	State                    StackStorageState `json:"state"`
	Conditions               []Condition       `json:"conditions"`
	Phase                    string            `json:"phase"`
	StorageServerServiceName string            `json:"storage_server_service_name"`
	LastObservedStatusHash   string            `json:"last_observed_status_hash,omitempty"`
	ObservedVersion          int64             `json:"observed_version"`
	// Subpath within the storage pod where different volumes are mounted.
	VolumeMountPaths []VolumeMountPath `json:"volume_mount_paths,omitempty"`
}

func (s StackStorageStatus) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StackStorageStatus) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte for WorkspaceStorageStatus failed")
	}
	return json.Unmarshal(b, &s)
}
