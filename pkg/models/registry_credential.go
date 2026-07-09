package models

import "time"

type RegistryCredentialPurpose string

const (
	RegistryCredentialPurposePull RegistryCredentialPurpose = "pull"
	RegistryCredentialPurposePush RegistryCredentialPurpose = "push"
	RegistryCredentialPurposeBoth RegistryCredentialPurpose = "both"
)

// RegistryCredentialIDAnnotation records which org-level registry credential a
// synced cluster secret was synthesized from.
const RegistryCredentialIDAnnotation = "stackdome.io/registry-credential-id"

func (p RegistryCredentialPurpose) Valid() bool {
	switch p {
	case RegistryCredentialPurposePull, RegistryCredentialPurposePush, RegistryCredentialPurposeBoth:
		return true
	default:
		return false
	}
}

// Covers reports whether a credential with purpose p can be used for the
// requested purpose.
func (p RegistryCredentialPurpose) Covers(requested RegistryCredentialPurpose) bool {
	return p == RegistryCredentialPurposeBoth || p == requested
}

// ConflictsWith reports whether two credentials for the same (org, host) pair
// overlap: 'both' conflicts with everything, otherwise only exact duplicates
// conflict.
func (p RegistryCredentialPurpose) ConflictsWith(other RegistryCredentialPurpose) bool {
	return p == other || p == RegistryCredentialPurposeBoth || other == RegistryCredentialPurposeBoth
}

// RegistryCredential is an org-level credential for a container registry host,
// auto-attached by host match when no explicit per-resource credential is set.
type RegistryCredential struct {
	ID                string                    `gorm:"primary_key;default:gen_random_uuid()"`
	OrganisationID    string                    `gorm:"not null;index"`
	Host              string                    `gorm:"not null"` // normalized registry host
	Purpose           RegistryCredentialPurpose `gorm:"not null;default:'both'"`
	Username          string                    `gorm:"not null"`
	EncryptedPassword string                    `gorm:"type:text;not null"`
	DataHash          string                    `gorm:"not null"` // for change detection / release freezing
	CreatedAt         time.Time
	UpdatedAt         time.Time

	// Transient plaintext password; never stored or serialized.
	Password string `gorm:"-" json:"-"`
}
