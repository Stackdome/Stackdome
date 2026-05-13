package models

import "time"

type RefreshToken struct {
	ID        string     `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	UserID    string     `gorm:"not null" json:"userId"`
	TokenHash string     `gorm:"not null" json:"-"`
	ExpiresAt time.Time  `gorm:"not null" json:"expiresAt"`
	CreatedAt time.Time  `gorm:"not null" json:"createdAt"`
	RevokedAt *time.Time `json:"revokedAt,omitempty"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
