package models

type SourceVolumeType string

const (
	EmptyVolume               SourceVolumeType = "EmptyVolume"
	RemoteDirSyncedVolume     SourceVolumeType = "RemoteDirSyncedVolume"
	BuildArtifactSyncedVolume SourceVolumeType = "BuildArtifactSyncedVolume"
	GitRepoVolume             SourceVolumeType = "GitRepoVolume"
)

type VolumeMount struct {
	StackID          string
	StackResourceID  string
	SourceVolumeID   string
	SourceVolumeType SourceVolumeType
	SourceSubPath    string
	TargetPath       string
}
