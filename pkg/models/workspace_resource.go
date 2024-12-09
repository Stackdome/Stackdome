package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type WorkspaceResource struct {
	ID          string `gorm:"default:gen_random_uuid()"`
	UserID      string `gorm:"not null"`
	WorkspaceID string
	Name        string      `gorm:"<-:create"`
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
	DependsOn       Dependencies     `gorm:"type:jsonb"`
	LifecycleConfig *LifecycleConfig `gorm:"type:jsonb"`
	Ports           Ports            `gorm:"type:jsonb"`
	StateFul        bool
	Status          *WorkspaceResourceStatus `gorm:"type:jsonb"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Dependencies []string

type Ports []Port

type SourceVolumeType string

const (
	EmptyVolume               SourceVolumeType = "EmptyVolume"
	LocalSyncedVolume         SourceVolumeType = "LocalSyncedVolume"
	BuildArtifactSyncedVolume SourceVolumeType = "BuildArtifactSyncedVolume"
)

type VolumeMount struct {
	WorkspaceStorageID  string
	WorkspaceResourceID string
	SourceVolumeID      string
	SourceVolumeType    SourceVolumeType
	SourceSubPath       string
	TargetPath          string
}

type WorkspaceResourceStatus struct {
	State                  string      `json:"state"`
	ObservedVersion        int64       `json:"observed_version"`
	Conditions             []Condition `json:"conditions"`
	PublicIngresses        []Ingress   `json:"public_ingresses"`
	InternalServiceName    *string     `json:"internal_service_name,omitempty"`
	LastObservedStatusHash string      `json:"last_observed_status_hash,omitempty"`
}

type Ingress struct {
	URL        string `json:"url"`
	TargetPort int    `json:"target_port"`
}

type Port struct {
	Number          int    `json:"number"`
	Protocol        string `json:"protocol"`
	ExposedToPublic bool   `json:"exposed_to_public"`
	SubdomainPrefix string `json:"subdomain_prefix"`
}

type LifecycleConfig struct {
	LastRestartRequestTime *time.Time `json:"last_restart_request_time"`
}

type BuildConfig struct {
	SourceVolumeID string `json:"source_volume_id"`
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

func (v *WorkspaceResource) VolumeMountMap() map[string]*VolumeMount {
	volumeMountMap := make(map[string]*VolumeMount)
	for i := range v.VolumeMounts {
		volumeMountMap[v.VolumeMounts[i].SourceVolumeID] = v.VolumeMounts[i]
	}
	return volumeMountMap
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

func (p *Ports) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &p)
}

func (p Ports) Value() (driver.Value, error) {
	return json.Marshal(p)
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

func (rs *WorkspaceResourceStatus) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &rs)
}

func (rs WorkspaceResourceStatus) Value() (driver.Value, error) {
	return json.Marshal(rs)
}
