package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

const (
	DefaultMaxActivePreviews = 10
	MaxMaxActivePreviews     = 50
	DefaultStackfilePath     = "stackfile.yaml"
	DefaultBaseBranch        = "main"
)

// PreviewGitRepository holds the git configuration for a preview environment.
type PreviewGitRepository struct {
	RepoURL     string  `json:"repo_url"`
	BaseBranch  string  `json:"base_branch"`
	GitSecretID *string `json:"git_secret_id,omitempty"`
}

func (r PreviewGitRepository) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *PreviewGitRepository) Scan(value any) error {
	if value == nil {
		*r = PreviewGitRepository{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed for PreviewGitRepository")
	}
	return json.Unmarshal(b, r)
}

// StackPreviewConfig defines how preview environments are created for a stack.
type StackPreviewConfig struct {
	ID                string `gorm:"primary_key;default:gen_random_uuid()"`
	OrganisationID    string `gorm:"not null"`
	TeamID            string `gorm:"not null"`
	UserID            string `gorm:"not null"`
	Name              string `gorm:"not null"`
	Description       string
	GitRepository     PreviewGitRepository `gorm:"type:jsonb;not null"`
	StackfilePath     string               `gorm:"not null;default:'stackfile.yaml'"`
	MaxActivePreviews int                  `gorm:"not null;default:10"`
	Labels            Labels               `gorm:"type:jsonb"`
	Annotations       Annotations          `gorm:"type:jsonb"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (s *StackPreviewConfig) GitRepoURL() string {
	return s.GitRepository.RepoURL
}

func (s *StackPreviewConfig) GitBaseBranch() string {
	return s.GitRepository.BaseBranch
}

func (s *StackPreviewConfig) UsesGitSecret() bool {
	return s.GitRepository.GitSecretID != nil && *s.GitRepository.GitSecretID != ""
}

func (s *StackPreviewConfig) GitSecretID() *string {
	return s.GitRepository.GitSecretID
}

func (StackPreviewConfig) TableName() string {
	return "stack_preview_configs"
}
