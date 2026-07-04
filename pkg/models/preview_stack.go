package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// PreviewStackSource indicates how a preview stack was created.
type PreviewStackSource string

const (
	PreviewStackSourceManual  PreviewStackSource = "manual"
	PreviewStackSourceWebhook PreviewStackSource = "webhook"
)

const (
	PreviewStackLabel    = "preview.stackdome.io/preview-stack"
	PreviewConfigIDLabel = "preview.stackdome.io/config-id"
	PreviewPRNumberLabel = "preview.stackdome.io/pr-number"
	PreviewStackIDLabel  = "preview.stackdome.io/preview-id"
)

// PreviewStackPhase represents the lifecycle phase of a preview stack.
type PreviewStackPhase string

const (
	PreviewStackPhaseProvisioning PreviewStackPhase = "Provisioning"
	PreviewStackPhaseDeploying    PreviewStackPhase = "Deploying"
	PreviewStackPhaseReady        PreviewStackPhase = "Ready"
	PreviewStackPhaseFailed       PreviewStackPhase = "Failed"
	PreviewStackPhaseDeleting     PreviewStackPhase = "Deleting"
)

type PreviewReconcilerStatus struct {
	LastAppliedCommitSHA     string    `json:"last_applied_commit_sha,omitempty"`
	LastAppliedStackfileHash string    `json:"last_applied_stackfile_hash,omitempty"`
	LastAppliedOverridesHash string    `json:"last_applied_overrides_hash,omitempty"`
	LastProcessedForceSyncAt time.Time `json:"last_processed_force_sync_at,omitempty"`
}

func (s PreviewReconcilerStatus) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *PreviewReconcilerStatus) Scan(value any) error {
	if value == nil {
		*s = PreviewReconcilerStatus{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return errors.New("unsupported type for PreviewReconcilerStatus")
	}
}

// PreviewStackStatus holds the current status of a preview stack.
type PreviewStackStatus struct {
	Phase   PreviewStackPhase    `json:"phase"`
	Reason  string               `json:"reason,omitempty"`
	Message string               `json:"message,omitempty"`
	Outputs *PreviewStackOutputs `json:"outputs,omitempty"`
}

func (s PreviewStackStatus) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *PreviewStackStatus) Scan(value any) error {
	if value == nil {
		*s = PreviewStackStatus{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return errors.New("unsupported type for PreviewStackStatus")
	}
}

// PreviewStackOutputs holds deployment outputs for a preview stack.
type PreviewStackOutputs struct {
	CommitSHA string       `json:"commit_sha,omitempty"`
	URLs      []PreviewURL `json:"urls,omitempty"`
}

// PreviewURL maps a resource name to its public URL.
type PreviewURL struct {
	Resource string `json:"resource"`
	URL      string `json:"url"`
}

// ImageOverrides maps resource names to image references for preview stacks.
type ImageOverrides map[string]string

func (o ImageOverrides) Value() (driver.Value, error) {
	return json.Marshal(o)
}

func (o *ImageOverrides) Scan(value any) error {
	if value == nil {
		*o = ImageOverrides{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, o)
	case string:
		return json.Unmarshal([]byte(v), o)
	default:
		return errors.New("unsupported type for ImageOverrides")
	}
}

// PreviewStack represents a preview environment spawned from a StackPreviewConfig.
type PreviewStack struct {
	ID                   string `gorm:"primary_key;default:gen_random_uuid()"`
	OrganisationID       string `gorm:"not null"`
	TeamID               string `gorm:"not null"`
	UserID               string `gorm:"not null"`
	StackPreviewConfigID string `gorm:"not null"`
	StackID              *string
	ActiveReleaseID      *string
	Name                 string                  `gorm:"not null"`
	PRNumber             string                  `gorm:"not null"`
	Branch               string                  `gorm:"not null"`
	CommitSHA            string                  `gorm:"not null"`
	Source               PreviewStackSource      `gorm:"not null"`
	ImageOverrides       ImageOverrides          `gorm:"type:jsonb"`
	Labels               Labels                  `gorm:"type:jsonb"`
	Annotations          Annotations             `gorm:"type:jsonb"`
	Status               PreviewStackStatus      `gorm:"type:jsonb;not null"`
	ReconcilerStatus     PreviewReconcilerStatus `gorm:"type:jsonb;not null;default:'{}'"`
	ForceSyncRequestedAt *time.Time
	StackfileContent     *string
	DeletionTimestamp    *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (PreviewStack) TableName() string {
	return "preview_stacks"
}
