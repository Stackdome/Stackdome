package models

import (
	"time"
)

const (
	DefaultOrgName = "Default"
)

type Organisation struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	ID        string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	Name      string
	Default   bool
	// Foriegn key to OrganisationDomain
	Domains []*OrganisationDomain `gorm:"foreignKey:OrganisationID"`
}
