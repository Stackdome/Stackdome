package models

type SourceVolumeType string

const (
	EmptyVolume   SourceVolumeType = "EmptyVolume"
	GitRepoVolume SourceVolumeType = "GitRepoVolume"
)

type VolumeMount struct {
	StackID          string
	StackResourceID  string
	SourceVolumeName string
	SourceVolumeID   string
	SourceVolumeType SourceVolumeType
	SourceSubPath    string
	TargetPath       string
}
