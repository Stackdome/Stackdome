package models

import (
	"time"
)

const (
	DefaultOrgName = "Default"
)

type Organisation struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	ID        string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	Name      string
	Default   bool
	// This doesnt get persisted or preloaded by gorm since we use polymorphic associations in the
	// Domain model. We have to manually load it when we need it.
	Domains []*Domain `gorm:"-"`
}
