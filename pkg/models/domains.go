package models

import "time"

type OwnerType string

const (
	OwnerTypeStackResource OwnerType = "stackresource"
	OwnerTypeOrganisation  OwnerType = "organisation"
)

type Domain struct {
	ID        string    `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	Fqdn      string    `gorm:"unique;not null"`
	OwnerID   string    `gorm:"not null;index"`
	OwnerType OwnerType `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DomainList []*Domain

func (d *DomainList) FqdnPresent(fqdn string) bool {
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

func (d *DomainList) FindByFqdn(fqdn string) *Domain {
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
