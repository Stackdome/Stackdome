package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/samber/lo"
)

const (
	StackIDLabel       = "stack.stackdome.io/id"
	StackIDAnnotation  = "stack.stackdome.io/id"
	StackRevisionLabel = "stack.stackdome.io/revision"
)

type StackState string

const (
	StackReady    StackState = "Ready"
	StackDeleting StackState = "Deleting"
	StackPending  StackState = "Pending"
	StackFailed   StackState = "Failed"
	StackError    StackState = "Error"
)

type Stack struct {
	ID                string      `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	OrganisationID    string      `gorm:"not null"`
	TeamID            string      `gorm:"index" json:"team_id"`
	ClusterID         string      `gorm:"not null"`
	UserID            string      `gorm:"not null"`
	Name              string      `gorm:"not null;<-:create"`
	NamespaceID       string      `gorm:"not null"`
	Namespace         string      `gorm:"unique;not null;<-:create"`
	Labels            Labels      `gorm:"type:jsonb"`
	Annotations       Annotations `gorm:"type:jsonb"`
	CrRevision        string
	Connections       StackConnections `gorm:"-"`
	StackResources    []*StackResource `gorm:"foreignKey:StackID"`
	Volumes           []*Volume        `gorm:"-"`
	Status            *StackStatus     `gorm:"type:jsonb"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletionTimestamp *time.Time `gorm:"default:NULL"`
}

// hasImageBuilds
func (s *Stack) HasImageBuilds() bool {
	for _, resource := range s.StackResources {
		if resource.BuildConfig != nil && (*resource.BuildConfig != BuildConfigSpec{}) {
			return true
		}
	}
	return false
}

type StackStatus struct {
	State                  StackState     `json:"state"`
	Message                string         `json:"message"`
	ObservedCrRevision     string         `json:"observed_cr_revision"`
	Conditions             []Condition    `json:"conditions"`
	LastObservedStatusHash string         `json:"last_observed_status_hash"`
	LastValidationRun      *ValidationRun `json:"last_validation_run"`
}

type ValidationRun struct {
	// List of validation types that were run.
	ValidationTypes []string `json:"validation_types"`
	StackRevision   string   `json:"stack_revision"`
	Passed          bool     `json:"passed"`
	Message         string   `json:"message"`
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

func (ws *Stack) SecretsInUse() []string {
	res := make([]string, 0)
	for _, resource := range ws.StackResources {
		if resource.HasGitCredentials() {
			if resource.BuildConfig != nil && resource.BuildConfig.SourceContext.Git != nil &&
				resource.BuildConfig.SourceContext.Git.GitSecretRef != nil {
				res = append(res, resource.BuildConfig.SourceContext.Git.GitSecretRef.SecretID)
			}
		}

		if resource.HasImagePullSecrets() {
			if resource.ImageConfig != nil && resource.ImageConfig.PullSecretRef != nil {
				res = append(res, resource.ImageConfig.PullSecretRef.SecretID)
			}
		}

		if resource.HasImagePushSecrets() {
			if resource.BuildConfig != nil && resource.BuildConfig.RegistrySecretRef != nil {
				res = append(res, resource.BuildConfig.RegistrySecretRef.SecretID)
			}
		}

		if resource.HasEnvVarsFromSecret() {
			if resource.ExecutionConfig != nil && len(resource.ExecutionConfig.EnvVarsFromSecrets) > 0 {
				for _, secretRef := range resource.ExecutionConfig.EnvVarsFromSecrets {
					res = append(res, secretRef.SecretID)
				}
			}
		}
	}
	for _, connection := range ws.Connections {
		if connection.From.Type == TopologyNodeTypeSecret && connection.From.Id != "" {
			res = append(res, connection.From.Id)
		}
	}
	return lo.Uniq(res)
}

func (ws *Stack) DirectConfigSecretsInUse() []string {
	var res []string
	for _, resource := range ws.StackResources {
		if resource.BuildConfig != nil && resource.BuildConfig.SourceContext.Git != nil &&
			resource.BuildConfig.SourceContext.Git.GitSecretRef != nil {
			res = append(res, resource.BuildConfig.SourceContext.Git.GitSecretRef.SecretID)
		}
		if resource.ImageConfig != nil && resource.ImageConfig.PullSecretRef != nil {
			res = append(res, resource.ImageConfig.PullSecretRef.SecretID)
		}
		if resource.BuildConfig != nil && resource.BuildConfig.RegistrySecretRef != nil {
			res = append(res, resource.BuildConfig.RegistrySecretRef.SecretID)
		}
	}
	return lo.Uniq(res)
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

func (s *Stack) HasImagePullSecrets() bool {
	for i := range s.StackResources {
		if s.StackResources[i].HasImagePullSecrets() {
			return true
		}
	}
	return false
}

// HasImagePushSecrets checks if any stack resource has an image push secret configured.
func (s *Stack) HasImagePushSecrets() bool {
	for i := range s.StackResources {
		if s.StackResources[i].HasImagePushSecrets() {
			return true
		}
	}
	return false
}

func (s *Stack) HasGitCredentials() bool {
	for i := range s.StackResources {
		if s.StackResources[i].HasGitCredentials() {
			return true
		}
	}
	return false
}

func (s *Stack) GetGitCredentialsMap() map[string]string {
	gitCredentialsMap := make(map[string]string)
	for i := range s.StackResources {
		if s.StackResources[i].BuildConfig != nil &&
			s.StackResources[i].BuildConfig.SourceContext.Git != nil &&
			s.StackResources[i].BuildConfig.SourceContext.Git.GitSecretRef != nil {
			gitCredentialsMap[s.StackResources[i].BuildConfig.SourceContext.Git.RepoURL] = s.StackResources[i].BuildConfig.SourceContext.Git.GitSecretRef.SecretID
		}
	}
	return gitCredentialsMap
}

// returns a map of image names to their pull secret IDs
func (s *Stack) GetImagePullSecretIDMap() map[string]string {
	imagePullSecretMap := make(map[string]string)
	for i := range s.StackResources {
		if s.StackResources[i].ImageConfig != nil && s.StackResources[i].ImageConfig.PullSecretRef != nil {
			imagePullSecretMap[s.StackResources[i].ImageConfig.Image] = s.StackResources[i].ImageConfig.PullSecretRef.SecretID
		}
	}
	return imagePullSecretMap
}

// GetImagePushSecretIDMap returns a map of image repository URLs to their push secret IDs.
func (s *Stack) GetImagePushSecretIDMap() map[string]string {
	imagePushSecretMap := make(map[string]string)
	for i := range s.StackResources {
		if s.StackResources[i].BuildConfig != nil && s.StackResources[i].BuildConfig.RegistrySecretRef != nil {
			imagePushSecretMap[s.StackResources[i].BuildConfig.ImageRepositoryUrl] = s.StackResources[i].BuildConfig.RegistrySecretRef.SecretID
		}
	}
	return imagePushSecretMap
}
