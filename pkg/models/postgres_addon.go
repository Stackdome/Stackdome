package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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

// ImportPasswordSecretName returns the predictable K8s Secret name used to
// store the password for importing from an external database.
func (p *PostgresAddon) ImportPasswordSecretName() string {
	return fmt.Sprintf("%s-import-password", p.Name)
}

// ExternalClusterRefName returns the identifier used in the PostgresCluster CR
// bootstrap import spec to reference the external database connection definition.
func (p *PostgresAddon) ExternalClusterRefName() string {
	return fmt.Sprintf("%s-external-ref", p.Name)
}

func (p *PostgresAddon) HasDatabase(name string) bool {
	for _, db := range p.Databases {
		if db.Name == name {
			return true
		}
	}
	return false
}

type PostgresAddonDatabase struct {
	ID              string             `gorm:"primary_key;default:gen_random_uuid()"`
	PostgresAddonID string             `gorm:"not null;index"`
	Name            string             `gorm:"not null"`
	Extensions      PostgresExtensions `gorm:"type:jsonb"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type PostgresBackupType string

const (
	PostgresBackupTypeManual    PostgresBackupType = "Manual"
	PostgresBackupTypeScheduled PostgresBackupType = "Scheduled"
)

type PostgresBackup struct {
	ID              string `gorm:"primary_key;default:gen_random_uuid()"`
	PostgresAddonID string `gorm:"not null;index"`
	Name            string
	Description     string
	Type            PostgresBackupType
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
	State                  string                       `json:"state,omitempty"`
	Message                string                       `json:"message,omitempty"`
	Conditions             []Condition                  `json:"conditions,omitempty"`
	ConnectionInfo         *PostgresAddonConnectionInfo `json:"connectionInfo,omitempty"`
	Databases              []PostgresDatabaseInfo       `json:"databases,omitempty"`
	LastObservedStatusHash string                       `json:"lastObservedStatusHash,omitempty"`
}

type PostgresAddonConnectionInfo struct {
	Host           string                       `json:"host,omitempty"`
	Port           int32                        `json:"port,omitempty"`
	SSLMode        string                       `json:"sslMode,omitempty"`
	WriteService   string                       `json:"writeService,omitempty"`
	ReadService    string                       `json:"readService,omitempty"`
	ClusterSecrets *PostgresAddonClusterSecrets `json:"clusterSecrets,omitempty"`
}

type PostgresAddonClusterSecrets struct {
	SuperuserSecret     *string           `json:"superuserSecret,omitempty"`
	UserSecrets         map[string]string `json:"userSecrets,omitempty"`
	CACertificateSecret string            `json:"caCertificateSecret,omitempty"`
}

type PostgresDatabaseInfo struct {
	Name  string `json:"name"`
	Owner string `json:"owner,omitempty"`
}

// PostgresCredentials is the JIT-fetched credential response (not stored in DB)
type PostgresCredentials struct {
	Database         string `json:"database"`
	Host             string `json:"host"`
	Port             int32  `json:"port"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	SSLMode          string `json:"sslMode"`
	ConnectionString string `json:"connectionString"`
	CACertificate    string `json:"caCertificate,omitempty"`
}

// ToFieldMap returns credential values keyed by PostgresAddonEnvFields names.
func (c *PostgresCredentials) ToFieldMap() map[string]string {
	m := make(map[string]string, len(PostgresAddonEnvFields))
	for _, field := range PostgresAddonEnvFields {
		switch field {
		case "host":
			m[field] = c.Host
		case "port":
			m[field] = strconv.Itoa(int(c.Port))
		case "username":
			m[field] = c.Username
		case "password":
			m[field] = c.Password
		case "database":
			m[field] = c.Database
		case "sslmode":
			m[field] = c.SSLMode
		case "connectionString":
			m[field] = c.ConnectionString
		case "caCertificate":
			m[field] = c.CACertificate
		}
	}
	return m
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
