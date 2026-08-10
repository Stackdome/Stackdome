package models

import "time"

type ComputeEntitlementSource string

const ComputeEntitlementSourceTrial ComputeEntitlementSource = "trial"

type ComputeEntitlementStatus string

const ComputeEntitlementStatusActive ComputeEntitlementStatus = "active"

// ComputeEntitlement records why an organisation may use compute. The alpha
// grants only trial entitlements; future subscription and licence sources can
// be added without changing runtime admission callers.
type ComputeEntitlement struct {
	ID             string                   `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	OrganisationID string                   `gorm:"not null;uniqueIndex:idx_compute_entitlement_org_source" json:"organisation_id"`
	Source         ComputeEntitlementSource `gorm:"not null;uniqueIndex:idx_compute_entitlement_org_source" json:"source"`
	Status         ComputeEntitlementStatus `gorm:"not null" json:"status"`
	StartsAt       time.Time                `gorm:"not null" json:"starts_at"`
	ExpiresAt      *time.Time               `gorm:"index" json:"expires_at,omitempty"`
	CreatedAt      time.Time                `json:"created_at"`
	UpdatedAt      time.Time                `json:"updated_at"`
}

type SharedComputeLeaseState string

const (
	SharedComputeLeaseStateActive         SharedComputeLeaseState = "active"
	SharedComputeLeaseStateCleanupPending SharedComputeLeaseState = "cleanup_pending"
	SharedComputeLeaseStateCleaning       SharedComputeLeaseState = "cleaning"
	SharedComputeLeaseStateCleaned        SharedComputeLeaseState = "cleaned"
	SharedComputeLeaseStateError          SharedComputeLeaseState = "error"
)

// SharedComputeLease records platform capacity reserved for an organisation.
// A lease continues consuming capacity after its entitlement expires until
// runtime cleanup reaches the cleaned state.
type SharedComputeLease struct {
	ID               string                  `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	OrganisationID   string                  `gorm:"not null;index" json:"organisation_id"`
	EntitlementID    string                  `gorm:"not null;uniqueIndex" json:"entitlement_id"`
	State            SharedComputeLeaseState `gorm:"not null" json:"state"`
	ActivatedAt      time.Time               `gorm:"not null" json:"activated_at"`
	CleanupStartedAt *time.Time              `json:"cleanup_started_at,omitempty"`
	CleanedUpAt      *time.Time              `json:"cleaned_up_at,omitempty"`
	ErrorAt          *time.Time              `json:"error_at,omitempty"`
	ErrorMessage     string                  `json:"error_message,omitempty"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
}

// ComputeAccess is the entitlement and shared-capacity lease observed by one
// admission decision. It is not persisted as its own record.
type ComputeAccess struct {
	Entitlement *ComputeEntitlement
	Lease       *SharedComputeLease
}

// ComputeLimits are the effective organisation limits enforced by runtime
// admission. They are configuration today and may later be resolved from a
// plan or licence without changing service callers.
type ComputeLimits struct {
	MaxStacksPerOrganization         int64 `yaml:"maxStacksPerOrganization" json:"max_stacks_per_organization"`
	MaxStackResourcesPerOrganization int64 `yaml:"maxStackResourcesPerOrganization" json:"max_stack_resources_per_organization"`
	ReplicasPerStackResource         int32 `yaml:"replicasPerStackResource" json:"replicas_per_stack_resource"`
	ConcurrentBuilds                 int   `yaml:"concurrentBuilds" json:"concurrent_builds"`
}
