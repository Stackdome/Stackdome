package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
)

type WorkspaceUserState string

const (
	WorkspaceUserIDLabel = "workspaceuser.stackdome.io/id"
)

const (
	WorkspaceUserProvisionPending   WorkspaceUserState = "Pending"
	WorkspaceUserProvisionCompleted WorkspaceUserState = "Provisioned"
	WorkspaceUserProvisionError     WorkspaceUserState = "Error"
)

type StringArray []string

type WorkspaceUserStatus struct {
	ObservedVersion       int64              `json:"observed_version"`
	ServiceAccountName    string             `json:"workspace_service_account_name"`
	ServiceAccountToken   string             `json:"workspace_service_account_token"`
	ClusterCACert         string             `json:"cluster_ca_cert"`
	ClusterUrl            string             `json:"cluster_url"`
	ProvisionedNamespaces []string           `json:"provisioned_namespaces"`
	Conditions            []Condition        `json:"conditions"`
	ClusterStatusHash     string             `json:"cluster_status_hash"`
	State                 WorkspaceUserState `json:"state"`
	Message               string             `json:"message"`
}

type WorkspaceUser struct {
	ID                  string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	UserID              string
	ClusterID           string
	OrganisationID      string
	ProjectID           string                `gorm:"index" json:"project_id"`
	WorkspaceNamespaces []*WorkspaceNamespace `gorm:"foreignKey:WorkspaceUserID;references:ID"`
	SshPublicKey        string
	// Tracks the version of the object in the database.
	Version           int
	Status            *WorkspaceUserStatus `gorm:"type:jsonb"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletionTimeStamp *time.Time `json:"deletion_timestamp"`
}

func (w *WorkspaceUser) WorkspaceNamespaceMap() map[string]*WorkspaceNamespace {
	res := make(map[string]*WorkspaceNamespace)
	for i := range w.WorkspaceNamespaces {
		res[w.WorkspaceNamespaces[i].Workspace] = w.WorkspaceNamespaces[i]
	}
	return res
}

func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = StringArray{}
		return nil
	}
	return pq.Array(sa).Scan(value)
}

func (sa StringArray) Value() (driver.Value, error) {
	return pq.Array(sa).Value()
}

func (s *WorkspaceUserStatus) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &s)
}

func (s WorkspaceUserStatus) Value() (driver.Value, error) {
	return json.Marshal(s)
}
