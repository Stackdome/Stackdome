package computeaccess

import "time"

type ComputeEntitlementSource string

const ComputeEntitlementSourceTrial ComputeEntitlementSource = "trial"

type ComputeEntitlementStatus string

const ComputeEntitlementStatusActive ComputeEntitlementStatus = "active"

// ComputeEntitlement records why an organisation may use compute.
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
// Capacity remains reserved until runtime cleanup reaches the cleaned state.
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

// ComputeAccess groups an entitlement and its shared-capacity lease. It is not
// persisted as its own record.
type ComputeAccess struct {
	Entitlement *ComputeEntitlement
	Lease       *SharedComputeLease
}
