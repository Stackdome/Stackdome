package models

import (
	"time"
)

type Cluster struct {
	ID             string `gorm:"primary_key;default:gen_random_uuid()"`
	OrganisationID string `gorm:"unique;not null"`
	Name           string `gorm:"not null;check:name <> ''"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Default        bool
	ClusterURL     string `gorm:"not null;check:cluster_url <> ''"`
	ClusterCAData  string `gorm:"not null;check:cluster_ca_data <> ''"`
	Token          string `gorm:"not null;check:token <> ''"`
	ManagerRunning bool
}
