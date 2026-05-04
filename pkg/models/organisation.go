package models

import (
	"time"
)

type Organisation struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	ID        string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	Name      string
	Domains   []*OrganisationDomain `gorm:"foreignKey:OrganisationID"`
}
