package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type WorkspaceResourceBuild struct {
	ID                    string
	Namespace             string
	WorkspaceID           string
	WorkspaceResourceID   string
	WorkspaceResourceName string
	BuildSourceHash       string
	ImageRegistry         string
	Status                *WorkspaceResourceBuildStatus `gorm:"type:jsonb"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type WorkspaceResourceBuildStatus struct {
	Conditions             []Condition `json:"conditions"`
	State                  string      `json:"state"`
	BuildSourceHash        string      `json:"build_source_hash"`
	ImageURL               string      `json:"image_url"`
	LastObservedStatusHash string      `json:"last_observed_status_hash,omitempty"`
}

// Unmarhsal and marshal JSONB column types
func (w *WorkspaceResourceBuildStatus) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &w)
}

func (w WorkspaceResourceBuildStatus) Value() (driver.Value, error) {
	return json.Marshal(w)
}
