package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type WorkspaceResource struct {
	ID          string      `gorm:"default:gen_random_uuid()" json:"id"`
	WorkspaceID string      `gorm:"primary_key; not null"`
	Name        string      `gorm:"primary_key; not null; <-:create"`
	Labels      Labels      `gorm:"type:jsonb"`
	Annotations Annotations `gorm:"type:jsonb"`
	// Tracks the version of the object in the database.
	Version         int
	ImageRegistry   *string
	Build           *BuildConfig     `gorm:"type:jsonb"`
	Prebuilt        *PrebuiltConfig  `gorm:"type:jsonb"`
	Init            *InitConfig      `gorm:"type:jsonb"`
	ExecutionConfig *ExecutionConfig `gorm:"type:jsonb"`
	VolumeMounts    []*VolumeMount   `gorm:"foreignKey:WorkspaceResourceID"`
	DependsOn       *Dependencies    `gorm:"type:jsonb"`
	LifecycleConfig *LifecycleConfig `gorm:"type:jsonb"`
	Ports           []*Port          `gorm:"foreignKey:WorkspaceResourceID"`
	StateFul        bool             `json:"stateFul"`
	Status          *ResourceStatus  `gorm:"type:jsonb"`
}

type Dependencies []string

type VolumeMount struct {
	WorkspaceResourceID string
	WorkspaceStorageID  string
	SourceVolumeID      string
	SourceSubPath       string
	TargetPath          string
}

type ResourceStatus struct {
	State           string      `json:"state"`
	ObservedVersion int64       `json:"observed_version"`
	Conditions      []Condition `json:"conditions"`
}

type Port struct {
	WorkspaceResourceID string
	Number              int
	Protocol            string
	ExposedToPublic     bool
	PublicURL           *string
}

type LifecycleConfig struct {
	LastRestartRequestTime *time.Time `json:"last_restart_request_time"`
}

type BuildConfig struct {
	ContextPath    string `json:"context_path"`
	DockerfilePath string `json:"dockerfile_path"`
	SourceHash     string `json:"source_hash"`
}

type PrebuiltConfig struct {
	ImageName string `json:"image_name"`
	Tag       string `json:"tag"`
}

type InitConfig struct {
	Command []string `json:"command"`
	Args    []string `json:"args"`
}

type ExecutionConfig struct {
	Command []string `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
	Env     []EnvVar `json:"env,omitempty"`
}

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (d *Dependencies) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &d)
}

func (d Dependencies) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (bc *BuildConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &bc)
}

func (bc BuildConfig) Value() (driver.Value, error) {
	return json.Marshal(bc)
}

func (pc *PrebuiltConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &pc)
}

func (pc PrebuiltConfig) Value() (driver.Value, error) {
	return json.Marshal(pc)
}

func (ic *InitConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &ic)
}

func (ic InitConfig) Value() (driver.Value, error) {
	return json.Marshal(ic)
}

func (ec *ExecutionConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &ec)
}

func (ec ExecutionConfig) Value() (driver.Value, error) {
	return json.Marshal(ec)
}

func (lc *LifecycleConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &lc)
}

func (lc LifecycleConfig) Value() (driver.Value, error) {
	return json.Marshal(lc)
}

func (rs *ResourceStatus) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &rs)
}

func (rs ResourceStatus) Value() (driver.Value, error) {
	return json.Marshal(rs)
}
