package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type ImageBuild struct {
	ID                string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	Name              string `gorm:"not null; <-:create" json:"name"`
	Namespace         string
	StackID           string
	StackResourceID   string
	StackResourceName string
	StackResource     *StackResource    `gorm:"foreignKey:StackResourceID"`
	Spec              BuildConfigSpec   `gorm:"type:jsonb"`
	Status            *ImageBuildStatus `gorm:"type:jsonb"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type BuildConfigSpec struct {
	SourceContext           BuildContextSource   `json:"source_context"`
	ContextPathWithinSource string               `json:"context_path_within_source"`
	DockerfilePath          string               `json:"dockerfile_path"`
	SourceRevision          BuildSourceRevision  `json:"source_revision"`
	BuildImageRepository    BuildImageRepository `json:"build_image_repository"`
	ImageRepositoryUrl      string               `json:"image_repository_url"`
}

type BuildImageRepository struct {
	InsecureRegistry     bool `json:"insecure_registry"`
	UseInClusterRegistry bool `json:"use_in_cluster_registry"`
}

func (b *BuildConfigSpec) Validate() error {
	ctx := b.SourceContext
	if (ctx.Volume == nil && ctx.Git == nil) || (ctx.Volume != nil && ctx.Git != nil) {
		return errors.New("exactly one of source_context.volume or source_context.git must be specified")
	}
	if ctx.Volume != nil {
		if ctx.Volume.SourceVolumeName == "" {
			return errors.New("source_context.volume: source_volume_name are required")
		}
	}
	if ctx.Git != nil {
		if ctx.Git.RepoURL == "" {
			return errors.New("source_context.git: repo_url is required")
		}
	}
	// SourceRevision is required, and only one of volume or git should be set
	rev := b.SourceRevision
	if (rev.Volume == nil && rev.Git == nil) || (rev.Volume != nil && rev.Git != nil) {
		return errors.New("exactly one of source_revision.volume or source_revision.git must be specified")
	}
	if rev.Volume != nil {
		if rev.Volume.CurrentVolumeHash == "" {
			return errors.New("source_revision.volume: current_volume_hash is required")
		}
	}
	if rev.Git != nil {
		if rev.Git.Branch == nil && rev.Git.Tag == "" && rev.Git.Commit == "" {
			return errors.New("source_revision.git: at least one of branch, tag, or commit is required")
		}
	}

	if b.ImageRepositoryUrl != "" && b.BuildImageRepository.UseInClusterRegistry {
		return errors.New("image_repository_url cannot be set if use_in_cluster_registry is true")
	}
	if b.ImageRepositoryUrl == "" && !b.BuildImageRepository.UseInClusterRegistry {
		// If the image repository URL is empty, we need to check if the in-cluster registry is set to true
		// If it is not, we need to return an error
		return errors.New("image_repository_url is required if use_in_cluster_registry is false")
	}
	return nil
}

type BuildSourceRevision struct {
	Volume *VolumeRevision `json:"volume"`
	Git    *GitRevision    `json:"git"`
}

type VolumeRevision struct {
	// Hash of the contents of the volume.
	CurrentVolumeHash string `json:"current_volume_hash"`
}

type GitRevision struct {
	Branch *GitBranch `json:"branch"`
	Tag    string     `json:"tag"`
	Commit string     `json:"commit"`
}

type BuildContextSource struct {
	Volume *VolumeBuildSource `json:"volume"`
	Git    *GitBuildSource    `json:"git"`
}

type VolumeBuildSource struct {
	SourceVolumeID   string `json:"source_volume_id"`
	SourceVolumeName string `json:"source_volume_name"`
}

type GitBuildSource struct {
	RepoURL string `json:"repo_url"`
}

type ImageBuildStatus struct {
	Conditions             []Condition `json:"conditions"`
	State                  string      `json:"state"`
	BuildSourceHash        string      `json:"build_source_hash"`
	ImageURL               string      `json:"image_url"`
	BuildSourceRevision    string      `json:"build_source_revision"`
	LastObservedStatusHash string      `json:"last_observed_status_hash,omitempty"`
}

// Unmarhsal and marshal JSONB column types
func (w *ImageBuildStatus) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &w)
}

func (w ImageBuildStatus) Value() (driver.Value, error) {
	return json.Marshal(w)
}

func (bc *BuildConfigSpec) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &bc)
}

func (bc BuildConfigSpec) Value() (driver.Value, error) {
	return json.Marshal(bc)
}
