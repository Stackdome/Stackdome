package models

import (
	"fmt"
	"time"
)

const PlatformOrganisationName = "PlatformOrganisation"

type Organisation struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	ID        string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	Name      string
	Platform  bool                  `gorm:"column:platform"`
	Domains   []*OrganisationDomain `gorm:"foreignKey:OrganisationID"`
}

func UserOrgNameFromOauth(name string) string {
	return fmt.Sprintf("org-%s", name)
}
