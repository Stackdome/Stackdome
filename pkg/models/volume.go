package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type GitRepoRevisionType string

const (
	Branch GitRepoRevisionType = "branch"
	Tag    GitRepoRevisionType = "tag"
	Commit GitRepoRevisionType = "commit"
)

type VolumeAccessMode string

const (
	READ_WRITE_ONCE VolumeAccessMode = "ReadWriteOnce"
	READ_WRITE_MANY VolumeAccessMode = "ReadWriteMany"
	READ_ONLY_MANY  VolumeAccessMode = "ReadOnlyMany"
)

const (
	VolumePhasePending = "Pending"
	VolumePhaseReady   = "Ready"
)

type Volume struct {
	ID             string           `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	OrganisationID string           `gorm:"not null" json:"organisation_id"`
	ProjectID      string           `gorm:"index" json:"project_id"`
	UserID         string           `gorm:"not null" json:"user_id"`
	Name           string           `gorm:"not null; <-:create" json:"name"`
	NamespaceID    string           `gorm:"not null" json:"namespace_id"`
	Namespace      string           `gorm:"not null;  <-:create" json:"namespace"`
	Labels         Labels           `gorm:"type:jsonb" json:"labels"`
	Annotations    Annotations      `gorm:"type:jsonb" json:"annotations"`
	Size           string           `gorm:"<-:create" json:"size"`
	StorageClass   string           `gorm:"<-:create" json:"storage_class,omitempty"`
	AccessMode     VolumeAccessMode `gorm:"not null; <-:create"`
	VolumeSource   *VolumeSource    `gorm:"type:jsonb" json:"volume_source,omitempty"`
	SyncBeforeUse  bool             `json:"sync_before_use"`
	Status         *VolumeStatus    `gorm:"type:jsonb" json:"volume_status,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

func (v *Volume) VolumeSourceType() SourceVolumeType {
	if v.VolumeSource != nil && v.VolumeSource.GitRepoSource != nil {
		return GitRepoVolume
	}
	return EmptyVolume
}

type VolumeSource struct {
	GitRepoSource *GitRepoSource `json:"git_repo_source,omitempty"`
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

type GitRepoSource struct {
	RepoUrl  string          `json:"repo_url"`
	Revision GitRepoRevision `json:"revision"`
}

func (g GitRepoSource) Value() (driver.Value, error) {
	return json.Marshal(g)
}

func (g *GitRepoSource) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte for GitRepoSource failed")
	}
	return json.Unmarshal(b, &g)
}

type GitRepoRevision struct {
	Branch string `json:"branch"`
	Tag    string `json:"tag"`
	Commit string `json:"commit"`
}

// Valid
func (g GitRepoRevision) IsValid() bool {
	switch g.Type() {
	case Branch:
		return g.Branch != ""
	case Tag:
		return g.Tag != ""
	case Commit:
		return g.Commit != ""
	default:
		return false
	}
}

func (g GitRepoRevision) Type() GitRepoRevisionType {
	switch {
	case g.Branch != "":
		return Branch
	case g.Tag != "":
		return Tag
	case g.Commit != "":
		return Commit
	default:
		return GitRepoRevisionType("")
	}
}

func (g GitRepoRevision) Value() (driver.Value, error) {
	return json.Marshal(g)
}

func (g *GitRepoRevision) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte for GitRepoRevision failed")
	}
	return json.Unmarshal(b, &g)
}

type VolumeStatus struct {
	ObservedGeneration     int64       `json:"observed_generation"`
	Conditions             []Condition `json:"conditions"`
	Phase                  string      `json:"phase"`
	LastObservedStatusHash string      `json:"last_observed_status_hash,omitempty"`
	InUse                  bool        `json:"in_use"`
	LastSyncedGitRevision  string      `json:"last_synced_git_revision,omitempty"`
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

type VolumeMountPath struct {
	VolumeID string `json:"volume_id"`
	Path     string `json:"path"`
}

// Scan and Value methods for VolumeMountPath
func (vmp *VolumeMountPath) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte for VolumeMountPath failed")
	}
	return json.Unmarshal(b, &vmp)
}

func (vmp VolumeMountPath) Value() (driver.Value, error) {
	return json.Marshal(vmp)
}
