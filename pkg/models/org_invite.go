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
	OrganisationID string       `gorm:"not null" json:"organisationId"`
	TeamID         string       `gorm:"not null" json:"teamId"`
	TeamRole       TeamRole     `gorm:"not null;column:team_role" json:"teamRole"`
	TokenHash      string       `gorm:"not null" json:"-"`
	EncryptedToken string       `gorm:"not null" json:"-"`
	Status         InviteStatus `gorm:"not null;default:pending" json:"status"`
	ExpiresAt      time.Time    `gorm:"not null" json:"expiresAt"`
	InvitedByID    string       `gorm:"not null" json:"invitedById"`
	AcceptedAt     *time.Time   `json:"acceptedAt,omitempty"`
	EmailSent      bool         `gorm:"not null;default:false" json:"emailSent"`
	EmailSentAt    *time.Time   `json:"emailSentAt,omitempty"`
	EmailError     *string      `json:"emailError,omitempty"`
	CreatedAt      time.Time    `gorm:"not null" json:"createdAt"`

	Organisation *Organisation `gorm:"foreignKey:OrganisationID" json:"organisation,omitempty"`
	Team         *Team         `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	InvitedBy    *User         `gorm:"foreignKey:InvitedByID" json:"invitedBy,omitempty"`
}
