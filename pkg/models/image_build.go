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

const (
	DefaultDockerfilePath = "Dockerfile"
	DefaultBuildContext   = "."
)

type BuildConfigSpec struct {
	SourceContext           BuildContextSource   `json:"source_context"`
	ContextPathWithinSource string               `json:"context_path_within_source"`
	DockerfilePath          string               `json:"dockerfile_path"`
	SourceRevision          BuildSourceRevision  `json:"source_revision"`
	BuildImageRepository    BuildImageRepository `json:"build_image_repository"`
	// PushRegistryCredentialID pins an org-level registry credential for the
	// push target, overriding host auto-attach.
	PushRegistryCredentialID string `json:"push_registry_credential_id,omitempty"`
}

type BuildImageRepository struct {
	InsecureRegistry     bool   `json:"insecure_registry"`
	UseInClusterRegistry bool   `json:"use_in_cluster_registry"`
	ClusterRegistryName  string `json:"cluster_registry_name,omitempty"`
	ExternalImageRef     string `json:"external_image_ref,omitempty"`
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
		if rev.Git.Branch == "" && rev.Git.Tag == "" {
			return errors.New("source_revision.git: a branch or tag is required (the commit SHA is resolved at release time)")
		}
		if rev.Git.Branch != "" && rev.Git.Tag != "" {
			return errors.New("source_revision.git: branch and tag cannot both be set")
		}
	}

	if b.BuildImageRepository.ExternalImageRef != "" && b.BuildImageRepository.UseInClusterRegistry {
		return errors.New("external_image_ref cannot be set if use_in_cluster_registry is true")
	}
	if b.BuildImageRepository.ExternalImageRef == "" && !b.BuildImageRepository.UseInClusterRegistry {
		return errors.New("external_image_ref is required if use_in_cluster_registry is false")
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
	Branch string `json:"branch"`
	Tag    string `json:"tag"`
	Commit string `json:"commit"`
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
	// IntegrationID pins an org-level git integration for clone auth,
	// overriding host auto-attach. Resolution lands with git integrations.
	IntegrationID string `json:"integration_id,omitempty"`
}

type ImageBuildStatus struct {
	Conditions             []Condition         `json:"conditions"`
	State                  string              `json:"state"`
	BuildSourceHash        string              `json:"build_source_hash"`
	ImageURL               string              `json:"image_url"`
	BuildSourceRevision    string              `json:"build_source_revision"`
	LastObservedStatusHash string              `json:"last_observed_status_hash,omitempty"`
	LastBuildFailureDetail *BuildFailureDetail `json:"last_build_failure_detail,omitempty"`
}

type BuildFailureDetail struct {
	FailureType  string `json:"failure_type"`
	Reason       string `json:"reason,omitempty"`
	Message      string `json:"message,omitempty"`
	RestartCount int32  `json:"restart_count"`
	ExitCode     *int32 `json:"exit_code,omitempty"`
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
