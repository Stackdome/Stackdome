package models

import "time"

type InviteStatus string

const (
	InviteStatusPending  InviteStatus = "pending"
	InviteStatusAccepted InviteStatus = "accepted"
	InviteStatusRevoked  InviteStatus = "revoked"
	InviteStatusExpired  InviteStatus = "expired"
)

type OrgInvite struct {
	ID             string       `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	Email          string       `gorm:"not null" json:"email"`
	OrganisationID string       `gorm:"not null" json:"organisation_id"`
	ProjectID      string       `gorm:"not null" json:"project_id"`
	ProjectRole    ProjectRole  `gorm:"not null;column:project_role" json:"project_role"`
	TokenHash      string       `gorm:"not null" json:"-"`
	EncryptedToken string       `gorm:"not null" json:"-"`
	Status         InviteStatus `gorm:"not null;default:pending" json:"status"`
	ExpiresAt      time.Time    `gorm:"not null" json:"expires_at"`
	InvitedByID    string       `gorm:"not null" json:"invited_by_id"`
	AcceptedAt     *time.Time   `json:"accepted_at,omitempty"`
	EmailSent      bool         `gorm:"not null;default:false" json:"email_sent"`
	EmailSentAt    *time.Time   `json:"email_sent_at,omitempty"`
	EmailError     *string      `json:"email_error,omitempty"`
	CreatedAt      time.Time    `gorm:"not null" json:"created_at"`

	Organisation *Organisation `gorm:"foreignKey:OrganisationID" json:"organisation,omitempty"`
	Project      *Project      `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	InvitedBy    *User         `gorm:"foreignKey:InvitedByID" json:"invited_by,omitempty"`
}
