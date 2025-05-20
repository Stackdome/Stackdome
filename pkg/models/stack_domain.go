package models

import "time"

type StackDomain struct {
	ID                string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	OrganisationID    string `gorm:"not null"`
	Fqdn              string `gorm:"unique;not null"`
	StackID           string `gorm:"not null;index"`
	StackResourceID   string `gorm:"not null;index"`
	StackResourceName string `gorm:"not null"`
	TargetPort        int    `gorm:"not null"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type StackDomainList []*StackDomain

func (d *StackDomainList) FqdnPresent(fqdn string) bool {
	if d == nil {
		return false
	}
	for _, domain := range *d {
		if domain == nil {
			continue
		}
		if domain.Fqdn == fqdn {
			return true
		}
	}
	return false
}

func (d *StackDomainList) FindByFqdn(fqdn string) *StackDomain {
	if d == nil {
		return nil
	}
	for _, domain := range *d {
		if domain == nil {
			continue
		}
		if domain.Fqdn == fqdn {
			return domain
		}
	}
	return nil
}
