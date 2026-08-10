package models

import "time"

type TrialAllocationState string

const (
	TrialAllocationStateActive         TrialAllocationState = "active"
	TrialAllocationStateCleanupPending TrialAllocationState = "cleanup_pending"
	TrialAllocationStateCleaning       TrialAllocationState = "cleaning"
	TrialAllocationStateCleaned        TrialAllocationState = "cleaned"
	TrialAllocationStateError          TrialAllocationState = "error"
)

type TrialAllocation struct {
	ID               string               `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	OrganisationID   string               `gorm:"not null;uniqueIndex" json:"organisation_id"`
	State            TrialAllocationState `gorm:"not null" json:"state"`
	ActivatedAt      time.Time            `gorm:"not null" json:"activated_at"`
	ExpiresAt        time.Time            `gorm:"not null;index" json:"expires_at"`
	CleanupStartedAt *time.Time           `json:"cleanup_started_at,omitempty"`
	CleanedUpAt      *time.Time           `json:"cleaned_up_at,omitempty"`
	ErrorAt          *time.Time           `json:"error_at,omitempty"`
	ErrorMessage     string               `json:"error_message,omitempty"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}
