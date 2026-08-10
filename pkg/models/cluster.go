package models

import (
	"time"
)

const DefaultClusterIssuerName = "letsencrypt-prod"

// Fixed names for the platform wildcard TLS resources. The DNS Issuer and its
// supporting Secret are namespace-scoped, so all of these resources reside in
// the namespace selected by the bootstrap configuration.
const (
	CloudflareAPITokenSecretName  = "cloudflare-api-token"
	CloudflareAPITokenSecretKey   = "api-token"
	DNSIssuerName                 = "letsencrypt-dns"
	DNSIssuerPrivateKeySecretName = "letsencrypt-dns-key"
	PlatformWildcardTLSName       = "platform-wildcard-tls"
)

const SharedComputeClusterName = "shared-compute"

type Cluster struct {
	ID                     string `gorm:"primary_key;default:gen_random_uuid()"`
	OrganisationID         string `gorm:"not null"`
	Name                   string `gorm:"not null;check:name <> ''"`
	CreatedAt              time.Time
	UpdatedAt              time.Time
	SharedCompute          bool
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
