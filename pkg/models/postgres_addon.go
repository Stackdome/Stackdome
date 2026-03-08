package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

const (
	PostgresAddonIDLabel = "postgres-addon.stackdome.io/id"
)

var SupportedPostgresExtensions = []string{
	"vector",
}

var _ Addon = &PostgresAddon{}

type PostgresAddon struct {
	ID             string      `gorm:"primary_key;default:gen_random_uuid()"`
	OrganisationID string      `gorm:"not null;index"`
	UserID         string      `gorm:"not null;index"`
	ClusterID      string      `gorm:"not null;index"`
	Name           string      `gorm:"not null"`
	NamespaceID    string      `gorm:"not null;index"`
	Namespace      string      `gorm:"not null"`
	Labels         Labels      `gorm:"type:jsonb"`
	Annotations    Annotations `gorm:"type:jsonb"`
	Revision       string

	// Spec fields
	PostgresVersion PostgresVersion        `gorm:"type:jsonb;not null"`
	Instances       PostgresInstances      `gorm:"type:jsonb;not null"`
	Resources       PostgresResources      `gorm:"type:jsonb;not null"`
	Storage         PostgresStorage        `gorm:"type:jsonb;not null"`
	Configuration   PostgresConfiguration  `gorm:"type:jsonb"`
	Initialization  PostgresInitialization `gorm:"type:jsonb"`
	BackupConfig    PostgresBackupConfig   `gorm:"type:jsonb"`

	// Lifecycle fields
	BackupRequestedAt *time.Time              `gorm:"default:NULL"`
	LifecycleConfig   PostgresLifecycleConfig `gorm:"type:jsonb"`

	// Status fields
	Status    PostgresAddonStatus `gorm:"type:jsonb"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// Relationships
	Databases []PostgresAddonDatabase `gorm:"foreignKey:PostgresAddonID"`
	Backups   []PostgresBackup        `gorm:"foreignKey:PostgresAddonID"`
}

func (p PostgresAddon) Type() string {
	return "postgres"
}

func (p PostgresAddon) AddonName() string {
	return p.Name
}

type PostgresAddonDatabase struct {
	ID              string             `gorm:"primary_key;default:gen_random_uuid()"`
	PostgresAddonID string             `gorm:"not null;index"`
	Name            string             `gorm:"not null"`
	Extensions      PostgresExtensions `gorm:"type:jsonb"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type PostgresBackup struct {
	ID              string `gorm:"primary_key;default:gen_random_uuid()"`
	PostgresAddonID string `gorm:"not null;index"`
	Name            string
	Description     string
	Type            string
	Phase           string
	StartedAt       *time.Time
	CompletedAt     *time.Time
	Error           string
	SizeBytes       *int32
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Domain types for PostgreSQL addon
type PostgresExtensions []string

type PostgresVersion struct {
	Major                     int  `json:"major"`
	Minor                     int  `json:"minor"`
	EnableMajorVersionUpgrade bool `json:"enableMajorVersionUpgrade,omitempty"`
	EnableMinorVersionUpgrade bool `json:"enableMinorVersionUpgrade,omitempty"`
}

type PostgresInstances struct {
	Count     int                        `json:"count"`
	Placement *PostgresInstancePlacement `json:"placement,omitempty"`
}

type PostgresInstancePlacement struct {
	TopologyKey  string               `json:"topology_key,omitempty"`
	Policy       string               `json:"policy,omitempty"` // preferred or required
	NodeSelector map[string]string    `json:"node_selector,omitempty"`
	Tolerations  []PostgresToleration `json:"tolerations,omitempty"`
}

type PostgresToleration struct {
	Key      string `json:"key"`
	Operator string `json:"operator"`
	Value    string `json:"value,omitempty"`
	Effect   string `json:"effect"`
}

type PostgresResources struct {
	CPU    PostgresCPUResource    `json:"cpu"`
	Memory PostgresMemoryResource `json:"memory"`
}

type PostgresCPUResource struct {
	Request string `json:"request"`
	Limit   string `json:"limit"`
}

type PostgresMemoryResource struct {
	Request string `json:"request"`
	Limit   string `json:"limit"`
}

type PostgresStorage struct {
	Size         string `json:"size"`
	StorageClass string `json:"storageClass,omitempty"`
}

type PostgresConfiguration struct {
	EnableSuperuserAccess bool              `json:"enableSuperuserAccess,omitempty"`
	Parameters            map[string]string `json:"parameters,omitempty"`
}

type PostgresInitialization struct {
	RestoreFromBackup      *PostgresRestoreFromBackup      `json:"restoreFromBackup,omitempty"`
	RestoreFromObjectStore *PostgresRestoreFromObjectStore `json:"restoreFromObjectStore,omitempty"`
	ImportFromExternal     *PostgresImportFromExternal     `json:"importFromExternal,omitempty"`
}

type PostgresRestoreFromBackup struct {
	BackupID string `json:"backupId"`
}

type PostgresRestoreFromObjectStore struct {
	ObjectStoreID         string     `json:"objectStoreId"`
	SourcePostgresAddonID string     `json:"sourcePostgresAddonId"`
	RecoveryTargetTime    *time.Time `json:"recoveryTargetTime,omitempty"`
}

type PostgresImportFromExternal struct {
	Host              string   `json:"host"`
	Port              int      `json:"port"`
	Database          string   `json:"database"`
	Username          string   `json:"username"`
	PasswordSecretID  string   `json:"passwordSecretId"`
	SslMode           *string  `json:"sslMode,omitempty"`
	DatabasesToImport []string `json:"databasesToImport,omitempty"`
}

type PostgresBackupConfig struct {
	ObjectStoreID string `json:"objectStoreId,omitempty"`
	Schedule      string `json:"schedule,omitempty"`
	WALArchiving  bool   `json:"walArchiving,omitempty"`
}

type PostgresLifecycleConfig struct {
	HibernationEnabled bool `json:"hibernationEnabled,omitempty"`
	FencingEnabled     bool `json:"fencingEnabled,omitempty"`
}

type PostgresAddonStatus struct {
	State          string                  `json:"state,omitempty"`
	Message        string                  `json:"message,omitempty"`
	Conditions     []Condition             `json:"conditions,omitempty"`
	ClusterInfo    *PostgresClusterInfo    `json:"clusterInfo,omitempty"`
	ConnectionInfo *PostgresConnectionInfo `json:"connectionInfo,omitempty"`
}

type PostgresClusterInfo struct {
	PrimaryInstance string `json:"primaryInstance,omitempty"`
	ReadyInstances  int    `json:"readyInstances,omitempty"`
	TotalInstances  int    `json:"totalInstances,omitempty"`
}

type PostgresConnectionInfo struct {
	Host        string                 `json:"host,omitempty"`
	Port        int                    `json:"port,omitempty"`
	Credentials PostgresCredentials    `json:"credentials,omitempty"`
	Databases   []PostgresDatabaseInfo `json:"databases,omitempty"`
}

type PostgresCredentials struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type PostgresDatabaseInfo struct {
	Name string `json:"name"`
	Size string `json:"size,omitempty"`
}

// Implement driver.Valuer and sql.Scanner for custom types
func (pv *PostgresVersion) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &pv)
}

func (pv PostgresVersion) Value() (driver.Value, error) {
	return json.Marshal(pv)
}

func (i *PostgresInstances) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &i)
}

func (i PostgresInstances) Value() (driver.Value, error) {
	return json.Marshal(i)
}

func (r *PostgresResources) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &r)
}

func (r PostgresResources) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (s *PostgresStorage) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &s)
}

func (s PostgresStorage) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (c *PostgresConfiguration) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &c)
}

func (c PostgresConfiguration) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (i *PostgresInitialization) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &i)
}

func (i PostgresInitialization) Value() (driver.Value, error) {
	return json.Marshal(i)
}

func (bc *PostgresBackupConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &bc)
}

func (bc PostgresBackupConfig) Value() (driver.Value, error) {
	return json.Marshal(bc)
}

func (s *PostgresAddonStatus) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &s)
}

func (s PostgresAddonStatus) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (e *PostgresExtensions) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &e)
}

func (e PostgresExtensions) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func (lc *PostgresLifecycleConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &lc)
}

func (lc PostgresLifecycleConfig) Value() (driver.Value, error) {
	return json.Marshal(lc)
}
