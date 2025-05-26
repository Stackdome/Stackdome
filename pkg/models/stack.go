package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	StackIDLabel = "stack.stackdome.io/id"
)

type StackState string

const (
	StackReady   StackState = "Ready"
	StackPending StackState = "Pending"
	StackFailed  StackState = "Failed"
)

type Stack struct {
	ID             string      `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	OrganisationID string      `gorm:"not null"`
	UserID         string      `gorm:"not null"`
	Name           string      `gorm:"not null;<-:create"`
	NamespaceID    string      `gorm:"not null"`
	Namespace      string      `gorm:"unique;not null;<-:create"`
	Labels         Labels      `gorm:"type:jsonb"`
	Annotations    Annotations `gorm:"type:jsonb"`
	Version        int
	StackResources []*StackResource `gorm:"foreignKey:StackID"`
	Volumes        []*Volume        `gorm:"-"`
	Status         *StackStatus     `gorm:"type:jsonb"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type StackStatus struct {
	State                  StackState  `json:"state"`
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

func (ws *Stack) VolumesMap() map[string]*Volume {
	volumeMap := make(map[string]*Volume)
	for i := range ws.Volumes {
		volumeMap[ws.Volumes[i].Name] = ws.Volumes[i]
	}
	return volumeMap
}

func (ws *Stack) HasVolumeMounts() bool {
	for i := range ws.StackResources {
		if len(ws.StackResources[i].VolumeMounts) > 0 {
			return true
		}
	}
	return false
}

func (ws *Stack) UsesInClusterRegistry() bool {
	for _, resource := range ws.StackResources {
		if resource.BuildConfig != nil && resource.BuildConfig.BuildImageRepository.UseInClusterRegistry {
			return true
		}
	}
	return false
}

func (ws *Stack) PopulateInternalImageRegistryUrlsForResources(registryUrl string) {
	for i := range ws.StackResources {
		curr := ws.StackResources[i]
		if curr.BuildConfig != nil && curr.BuildConfig.BuildImageRepository.UseInClusterRegistry {
			curr.BuildConfig.ImageRepositoryUrl = fmt.Sprintf(
				"%s/%s/%s/%s", registryUrl, ws.OrganisationID, ws.Name, curr.Name)
		}
	}
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

func (s *Stack) HasExposedPorts() bool {
	for i := range s.StackResources {
		for j := range s.StackResources[i].Ports {
			if s.StackResources[i].Ports[j].ExposedToPublic {
				return true
			}
		}
	}
	return false
}
