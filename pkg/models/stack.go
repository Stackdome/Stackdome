package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/samber/lo"
)

const (
	StackIDLabel       = "core.stackdome.io/stack-id"
	StackIDAnnotation  = "stack.stackdome.io/id"
	StackRevisionLabel = "stack.stackdome.io/revision"
)

const (
	SkipClusterProvisioningAnnotation = "stack.stackdome.io/skip-cluster-provisioning"
	SimulateReleaseStateAnnotation    = "stack.stackdome.io/simulate-release-state"
)

type StackState string

const (
	StackReady       StackState = "Ready"
	StackDeleting    StackState = "Deleting"
	StackPending     StackState = "Pending"
	StackFailed      StackState = "Failed"
	StackError       StackState = "Error"
	StackDegraded    StackState = "Degraded"
	StackProgressing StackState = "Progressing"
)

// StackPhase mirrors the cluster-agent's phase vocabulary for API-facing typed enums.
type StackPhase string

const (
	StackPhaseReady       StackPhase = "Ready"
	StackPhasePending     StackPhase = "Pending"
	StackPhaseFailed      StackPhase = "Failed"
	StackPhaseDegraded    StackPhase = "Degraded"
	StackPhaseProgressing StackPhase = "Progressing"
)

type StackConditionType string

const (
	StackConditionAvailable      StackConditionType = "Available"
	StackConditionResourcesReady StackConditionType = "ResourcesReady"
	StackConditionStalled        StackConditionType = "Stalled"
	StackConditionDegraded       StackConditionType = "Degraded"
	StackConditionProgressing    StackConditionType = "Progressing"
)

type StackConvergenceRecord struct {
	Revision  string    `json:"revision"`
	ReleaseID string    `json:"release_id,omitempty"`
	At        time.Time `json:"at"`
}

type StackResourceSummary struct {
	Name              string                          `json:"name"`
	Phase             StackResourcePhase              `json:"phase,omitempty"`
	ObservedRevision  string                          `json:"observed_revision,omitempty"`
	ConvergedRevision string                          `json:"converged_revision,omitempty"`
	LastConverged     *StackResourceConvergenceRecord `json:"last_converged,omitempty"`
	AvailableReplicas int32                           `json:"available_replicas,omitempty"`
	UpdatedReplicas   int32                           `json:"updated_replicas,omitempty"`
	Replicas          int32                           `json:"replicas,omitempty"`
	Missing           bool                            `json:"missing,omitempty"`
	Message           string                          `json:"message,omitempty"`
}

type StackResourceConvergenceRecord struct {
	Revision string    `json:"revision"`
	At       time.Time `json:"at"`
}

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
	State                  StackState              `json:"state"`
	Message                string                  `json:"message"`
	ObservedCrRevision     string                  `json:"observed_cr_revision"`
	Conditions             []Condition             `json:"conditions"`
	LastObservedStatusHash string                  `json:"last_observed_status_hash"`
	LastValidationRun      *ValidationRun          `json:"last_validation_run"`
	TargetRevision         string                  `json:"target_revision,omitempty"`
	LastConverged          *StackConvergenceRecord `json:"last_converged,omitempty"`
	Resources              []StackResourceSummary  `json:"resources,omitempty"`
}

func (s *StackStatus) GetLastConverged() *StackConvergenceRecord {
	if s == nil {
		return nil
	}
	return s.LastConverged
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
