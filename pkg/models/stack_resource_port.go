package models

type StackResourcePort struct {
	ID              string `gorm:"default:gen_random_uuid();primaryKey"`
	StackResourceID string `gorm:"not null;index"`
	OrganisationID  string `gorm:"not null;index"`
	Number          int
	Protocol        string
	ExposedToPublic bool
	SubdomainPrefix string
	DomainID        string `gorm:"not null"`
	Domain          Domain `gorm:"foreignKey:DomainID"`
}
