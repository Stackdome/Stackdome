package models

import "time"

// GitInstallation records a GitHub App installation on a GitHub account,
// linked to the org's github_app integration.
type GitInstallation struct {
	ID                  string `gorm:"primary_key;default:gen_random_uuid()"`
	GitIntegrationID    string `gorm:"not null;index"`
	InstallationID      int64  `gorm:"not null"`
	AccountLogin        string `gorm:"not null"`
	AccountType         string
	RepositorySelection string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// GitHub account types as reported on installations.
const (
	GitAccountTypeUser         = "User"
	GitAccountTypeOrganization = "Organization"
)
