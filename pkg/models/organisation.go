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
	ID         int `gorm:"primary_key;autoIncrement"`
	Name       string
	DomainName string // TLD + SLD. Ex: example.test
	Default    bool
}
