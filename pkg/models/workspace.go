package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Workspace struct {
	ID                 string      `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	OrganisationID     string      `gorm:"not null"`
	UserID             string      `gorm:"not null"`
	Name               string      `gorm:"not null; <-:create"`
	Namespace          string      `gorm:"unique;not null;  <-:create"`
	Labels             Labels      `gorm:"type:jsonb"`
	Annotations        Annotations `gorm:"type:jsonb"`
	Version            int
	WorkspaceResources []*WorkspaceResource `gorm:"foreignKey:WorkspaceID"`
	Status             *WorkspaceStatus     `gorm:"type:jsonb"`
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
