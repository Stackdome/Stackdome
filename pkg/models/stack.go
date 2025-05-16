package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

const (
	StackIDLabel = "stack.stackdome.io/id"
)

type Stack struct {
	ID             string      `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	OrganisationID string      `gorm:"not null"`
	UserID         string      `gorm:"not null"`
	Name           string      `gorm:"not null;<-:create"`
	WorkspaceName  string      `gorm:"not null;<-:create"`
	Namespace      string      `gorm:"unique;not null;<-:create"`
	Labels         Labels      `gorm:"type:jsonb"`
	Annotations    Annotations `gorm:"type:jsonb"`
	Version        int
	StackResources []*StackResource `gorm:"foreignKey:StackID"`
	Status         *StackStatus     `gorm:"type:jsonb"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type StackStatus struct {
	State                  string      `json:"state"`
	ObservedVersion        int64       `json:"observed_version"`
	Conditions             []Condition `json:"conditions"`
	LastObservedStatusHash string      `json:"last_observed_status_hash,omitempty"`
}

func (ws *StackStatus) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &ws)
}

func (ws StackStatus) Value() (driver.Value, error) {
	return json.Marshal(ws)
}

func (ws *Stack) ResourcesMap() map[string]*StackResource {
	resourceMap := make(map[string]*StackResource)
	for i := range ws.StackResources {
		resourceMap[ws.StackResources[i].Name] = ws.StackResources[i]
	}
	return resourceMap
}

func (ws *Stack) HasVolumeMounts() bool {
	for i := range ws.StackResources {
		if len(ws.StackResources[i].VolumeMounts) > 0 {
			return true
		}
	}
	return false
}

func (ws *Stack) VolumeMountIds() []string {
	volumeMountIds := make([]string, 0)
	for i := range ws.StackResources {
		for j := range ws.StackResources[i].VolumeMounts {
			volumeMountIds = append(volumeMountIds, ws.StackResources[i].VolumeMounts[j].SourceVolumeID)
		}
	}
	return volumeMountIds
}

func (s *Stack) ExposedPortFqdnMap() map[string]map[int]string {
	exposedPortFqdnMap := make(map[string]map[int]string)
	for i := range s.StackResources {
		exposedPortFqdnMap[s.StackResources[i].Name] = make(map[int]string)
		for j := range s.StackResources[i].Ports {
			if s.StackResources[i].Ports[j].ExposedToPublic {
				exposedPortFqdnMap[s.StackResources[i].Name][s.StackResources[i].Ports[j].Number] = s.StackResources[i].Ports[j].ExposedFqdn
			}
		}
	}
	return exposedPortFqdnMap
}
