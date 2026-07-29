package models

import (
	"time"
)

const DefaultClusterIssuerName = "letsencrypt-prod"

// PlatformClusterName is the fixed name of the bootstrap-provisioned platform
// cluster row; it is not configurable.
const PlatformClusterName = "platform"

type Cluster struct {
	ID                     string `gorm:"primary_key;default:gen_random_uuid()"`
	OrganisationID         string `gorm:"unique;not null"`
	Name                   string `gorm:"not null;check:name <> ''"`
	CreatedAt              time.Time
	UpdatedAt              time.Time
	Platform               bool
	ClusterURL             string `gorm:"not null;check:cluster_url <> ''"`
	EncryptedClusterCAData string `gorm:"not null"`
	EncryptedToken         string `gorm:"not null"`
	ManagerRunning         bool
	ClusterInfo            *ClusterInfo            `gorm:"type:jsonb"`
	ImageRegistries        []*ClusterImageRegistry `gorm:"foreignKey:ClusterID"`

	// Transient fields — populated by service layer after decryption, never persisted
	ClusterCAData string `gorm:"-"`
	Token         string `gorm:"-"`
}
