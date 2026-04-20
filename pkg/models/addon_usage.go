package models

type AddonType string

const (
	AddonTypePostgres AddonType = "postgres"
)

type AddonUsage struct {
	ID              string    `gorm:"primary_key;default:gen_random_uuid()"`
	AddonType       AddonType `gorm:"not null"`
	AddonID         string    `gorm:"not null"`
	StackID         string    `gorm:"not null"`
	StackResourceID string    `gorm:"not null"`
}
