package models

type SecretUsage struct {
	SecretID string `gorm:"not null"`
	StackID  string `gorm:"not null"`
}
