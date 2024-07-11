package models

import (
	"time"
)

const (
	DefaultOrgName = "Default"
)

type Organisation struct {
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ID         string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	Name       string
	DomainName string // TLD + SLD. Ex: example.test
	Default    bool
}
