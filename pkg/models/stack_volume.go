package models

type StackVolume struct {
	StackID  string `gorm:"primary_key"`
	VolumeID string `gorm:"primary_key"`
}
