package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
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

type Volume struct {
	ID             string           `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	OrganisationID string           `gorm:"not null" json:"organisation_id"`
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
}

func (v *Volume) VolumeSourceType() SourceVolumeType {
	switch {
	case v.VolumeSource == nil:
		return EmptyVolume
	case len(v.VolumeSource.BuildSource) != 0:
		return BuildArtifactSyncedVolume
	case v.VolumeSource.RemoteDirSource != nil:
		return RemoteDirSyncedVolume
	case v.VolumeSource.GitRepoSource != nil:
		return GitRepoVolume
	default:
		return EmptyVolume
	}
}

type VolumeSource struct {
	RemoteDirSource *RemoteDirSource     `json:"remote_dir_source,omitempty"`
	BuildSource     BuildArtifactSources `json:"build_source,omitempty"`
	GitRepoSource   *GitRepoSource       `json:"git_repo_source,omitempty"`
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
	Branch GitBranch `json:"branch"`
	Tag    string    `json:"tag"`
	Commit string    `json:"commit"`
}

// Valid
func (g GitRepoRevision) IsValid() bool {
	switch g.Type() {
	case Branch:
		return g.Branch.Name != ""
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
	case g.Branch.Name != "":
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

type GitBranch struct {
	Name    string `json:"name"`
	HeadSha string `json:"head_sha"`
}

func (g GitBranch) Value() (driver.Value, error) {
	return json.Marshal(g)
}

func (g *GitBranch) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte for GitBranch failed")
	}
	return json.Unmarshal(b, &g)
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
	LastSyncedGitRevision  string                  `json:"last_synced_git_revision,omitempty"`
	LastRemoteDirSyncHash  string                  `json:"last_remote_dir_sync_hash,omitempty"`
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

func (b BuildArtifactSyncInfo) Value() (driver.Value, error) {
	return json.Marshal(b)
}

func (b *BuildArtifactSyncInfo) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte for BuildArtifactSyncInfo failed")
	}
	return json.Unmarshal(v, &b)
}

type RemoteDirSource struct {
	// Path within the client where the directory to be synced is located.
	Path string `json:"path"`
	// We use this to track the current state of the remote directory.
	CurrentDirectoryHash string `json:"current_directory_hash"`
}

func (l RemoteDirSource) Value() (driver.Value, error) {
	return json.Marshal(l)
}

func (l *RemoteDirSource) Scan(value interface{}) error {
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

func (b BuildArtifactSource) Value() (driver.Value, error) {
	return json.Marshal(b)
}
func (b *BuildArtifactSource) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte for BuildArtifactSource failed")
	}
	return json.Unmarshal(v, &b)
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
