package models

type ResourceUsage struct {
	ID           string `gorm:"primary_key;default:gen_random_uuid()"`
	ResourceType string `gorm:"not null"`
	ResourceID   string `gorm:"not null"`
	ConsumerType string `gorm:"not null"`
	ConsumerID   string `gorm:"not null"`
	StackID      string `gorm:"not null;index"`
}

func (ResourceUsage) TableName() string {
	return "resource_usages"
}
