package models

import "time"

// ResourceValidationCheckKind names an expensive release-time check.
type ResourceValidationCheckKind string

const (
	ValidationCheckImagePull  ResourceValidationCheckKind = "image_pull"
	ValidationCheckPushAccess ResourceValidationCheckKind = "push_access"
)

// ResourceValidationRecord remembers the last successful expensive validation
// for a resource so unchanged targets are not re-probed on every release.
// Failures are never recorded.
type ResourceValidationRecord struct {
	StackID      string                      `gorm:"primaryKey"`
	ResourceName string                      `gorm:"primaryKey"`
	CheckKind    ResourceValidationCheckKind `gorm:"primaryKey"`
	Fingerprint  string                      `gorm:"not null"`
	ValidatedAt  time.Time                   `gorm:"not null"`
}

func (ResourceValidationRecord) TableName() string {
	return "resource_validation_records"
}
