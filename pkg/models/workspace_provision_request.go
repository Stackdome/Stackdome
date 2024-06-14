package models

import "time"

type ProvisionRequestStatusCondition string

const (
	ProvisionRequestPending   ProvisionRequestStatusCondition = "Pending"
	ProvisionRequestCompleted ProvisionRequestStatusCondition = "Completed"
	ProvisionRequestError     ProvisionRequestStatusCondition = "Error"
)

type WorkspaceProvisionRequestStatus struct {
	WorkspaceProvisionRequestID  string `gorm:"primary_key"`
	WorkspaceNamespace           *string
	WorkspaceServiceAccountName  *string
	WorkspaceServiceAccountToken *string
	ClusterCACert                *string
	ClusterUrl                   *string
	Domain                       *string
	StatusCondition              string
	Message                      string
}

type WorkspaceProvisionRequest struct {
	ID             string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	UserID         string
	OrganisationID int
	SshPublicKey   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Status         *WorkspaceProvisionRequestStatus `gorm:"foreignkey:WorkspaceProvisionRequestID;references:ID"`
}
