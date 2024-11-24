package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Workspace struct {
	ID                 string      `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	OrganisationID     string      `gorm:"not null"`
	UserID             string      `gorm:"not null"`
	Name               string      `gorm:"not null;<-:create"`
	Namespace          string      `gorm:"unique;not null;<-:create"`
	Labels             Labels      `gorm:"type:jsonb"`
	Annotations        Annotations `gorm:"type:jsonb"`
	Version            int
	WorkspaceResources []*WorkspaceResource `gorm:"foreignKey:WorkspaceID"`
	Status             *WorkspaceStatus     `gorm:"type:jsonb"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type WorkspaceStatus struct {
	State           string      `json:"state"`
	ObservedVersion int64       `json:"observed_version"`
	Conditions      []Condition `json:"conditions"`
}

func (ws *WorkspaceStatus) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &ws)
}

func (ws WorkspaceStatus) Value() (driver.Value, error) {
	return json.Marshal(ws)
}

func (ws *Workspace) ResourcesMap() map[string]*WorkspaceResource {
	resourceMap := make(map[string]*WorkspaceResource)
	for i := range ws.WorkspaceResources {
		resourceMap[ws.WorkspaceResources[i].Name] = ws.WorkspaceResources[i]
	}
	return resourceMap
}

func (ws *Workspace) HasVolumeMounts() bool {
	for i := range ws.WorkspaceResources {
		if len(ws.WorkspaceResources[i].VolumeMounts) > 0 {
			return true
		}
	}
	return false
}
