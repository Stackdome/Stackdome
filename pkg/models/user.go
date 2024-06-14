package models

import (
	"time"
)

type Role string

const (
	UserRole              Role = "User"
	OrganisationAdminRole Role = "OrganisationAdmin"
	PlatformAdminRole     Role = "PlatformAdmin"
)

type User struct {
	ID             string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	InternalID     int    `gorm:"autoIncrement;<-:false"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Name           string
	Email          string `gorm:"unique"`
	Password       string
	Organisation   string
	Role           Role
	OrganisationID int
}

func (u *User) GetInternalID() int {
	return u.InternalID
}
