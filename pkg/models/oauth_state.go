package models

import "time"

type OAuthState struct {
	ID        string    `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	State     string    `gorm:"not null;uniqueIndex" json:"state"`
	Provider  string    `gorm:"not null" json:"provider"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
}

const (
	OAuthProviderGitHub = "github"
)

func (OAuthState) TableName() string {
	return "oauth_states"
}
