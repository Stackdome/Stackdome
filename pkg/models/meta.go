package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	ObjectServerGeneration = "managedobject.stackdome.io/generation"
)

type Meta struct {
	ID        string `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
