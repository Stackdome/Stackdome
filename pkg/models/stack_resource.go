package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type StackResource struct {
	ID          string `gorm:"default:gen_random_uuid()"`
	UserID      string `gorm:"not null"`
	StackID     string
	Name        string      `gorm:"<-:create"`
	Namespace   string      `gorm:"<-:create"`
	Labels      Labels      `gorm:"type:jsonb"`
	Annotations Annotations `gorm:"type:jsonb"`
	// Tracks the version of the object in the database.
	Version         int
	BuildConfig     *BuildConfigSpec `gorm:"type:jsonb"`
	ImageConfig     *ImageConfigSpec `gorm:"type:jsonb"`
	Init            *InitConfig      `gorm:"type:jsonb"`
	ExecutionConfig *ExecutionConfig `gorm:"type:jsonb"`
	VolumeMounts    []*VolumeMount   `gorm:"foreignKey:StackResourceID"`
	DependsOn       Dependencies     `gorm:"type:jsonb"`
	LifecycleConfig *LifecycleConfig `gorm:"type:jsonb"`
	Ports           Ports            `gorm:"type:jsonb"`
	StateFul        bool
	Status          *StackResourceStatus `gorm:"type:jsonb"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Dependencies []string

type Ports []Port

type StackResourceStatus struct {
	State                         string      `json:"state"`
	ObservedVersion               int64       `json:"observed_version"`
	Conditions                    []Condition `json:"conditions"`
	PublicIngresses               []Ingress   `json:"public_ingresses"`
	InternalServiceName           *string     `json:"internal_service_name,omitempty"`
	LastObservedStatusHash        string      `json:"last_observed_status_hash,omitempty"`
	LastRestartRequestProcessedAt *time.Time  `json:"last_restart_request_processed_at,omitempty"`
}

type Ingress struct {
	URL        string `json:"url"`
	TargetPort int    `json:"target_port"`
}

type Port struct {
	Number                   int    `json:"number"`
	Protocol                 string `json:"protocol"`
	ExposedToPublic          bool   `json:"exposed_to_public"`
	SubdomainPrefix          string `json:"subdomain_prefix"`
	GeneratedSubdomainPrefix string `json:"generated_subdomain_prefix"`
	ExposedFqdn              string `json:"fqdn"`
}

type LifecycleConfig struct {
	RestartRequestTime *time.Time `json:"restart_request_time"`
}

type ImageConfigSpec struct {
	Image string `json:"image"`
}

func (i ImageConfigSpec) Validate() error {
	if i.Image == "" {
		return errors.New("image cannot be empty")
	}
	return nil
}

type InitConfig struct {
	ImageConfig *ImageConfigSpec `json:"image_config"`
	Command     []string         `json:"command"`
	Args        []string         `json:"args"`
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

func (v *StackResource) VolumeMountMap() map[string]*VolumeMount {
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

func (pc *ImageConfigSpec) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &pc)
}

func (pc ImageConfigSpec) Value() (driver.Value, error) {
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

func (rs *StackResourceStatus) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &rs)
}

func (rs StackResourceStatus) Value() (driver.Value, error) {
	return json.Marshal(rs)
}
