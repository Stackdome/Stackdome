package models

import (
	"time"
)

type Cluster struct {
	ID             string `gorm:"primary_key;default:gen_random_uuid()"`
	OrganisationID int    `gorm:"unique;not null"`
	Name           string `gorm:"not null;check:name <> ''"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Default        bool
	ClusterURL     string `gorm:"not null;check:cluster_url <> ''"`
	ClusterCAData  string `gorm:"not null;check:cluster_ca_data <> ''"`
	ClientCertData string `gorm:"not null;check:client_cert_data <> ''"`
	ClientKeyData  string `gorm:"not null;check:client_key_data <> ''"`
}
