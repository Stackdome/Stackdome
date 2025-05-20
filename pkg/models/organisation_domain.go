package models

import "time"

type OrganisationDomain struct {
	ID             string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	OrganisationID string `gorm:"not null"`
	Domain         string `gorm:"unique;not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
