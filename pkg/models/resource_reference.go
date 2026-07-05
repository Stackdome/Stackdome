package models

import "time"

type ReferentType string

const (
	ReferentSecret             ReferentType = "secret"
	ReferentVolume             ReferentType = "volume"
	ReferentPostgresAddon      ReferentType = "postgres_addon"
	ReferentRegistryCredential ReferentType = "registry_credential"
	ReferentGitIntegration     ReferentType = "git_integration"
)

type RelationKind string

const (
	RelationEnv           RelationKind = "env"
	RelationVolumeMount   RelationKind = "volume_mount"
	RelationBuildArtifact RelationKind = "build_artifact_source"
	RelationGitCredential RelationKind = "git_credential"
	RelationImagePull     RelationKind = "image_pull"
	RelationImagePush     RelationKind = "image_push"
)

// ResourceReference is a derived index row: a scope (the live stack spec when
// ReleaseID is nil, or a specific release) references a deletable referent.
type ResourceReference struct {
	ID           string       `gorm:"primary_key;default:gen_random_uuid()"`
	StackID      string       `gorm:"not null;index"`
	ReleaseID    *string      `gorm:"index"` // nil => live spec
	ReferentType ReferentType `gorm:"not null"`
	ReferentID   string       `gorm:"not null"`
	RelationKind RelationKind `gorm:"not null"`
	CreatedAt    time.Time
}

func (ResourceReference) TableName() string {
	return "resource_references"
}
