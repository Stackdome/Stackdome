package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type ReleaseEventType string

const (
	ReleaseEventTypeReleaseCreated       ReleaseEventType = "release_created"
	ReleaseEventTypeReleaseChecksStarted ReleaseEventType = "release_checks_started"
	ReleaseEventTypeReleaseCheckFailed   ReleaseEventType = "release_check_failed"
	ReleaseEventTypeReleaseChecksPassed  ReleaseEventType = "release_checks_passed"
	ReleaseEventTypeReleaseStarted       ReleaseEventType = "release_started"
	ReleaseEventTypeBuildQueued          ReleaseEventType = "build_queued" // reserved; not emitted in v1
	ReleaseEventTypeBuildStarted         ReleaseEventType = "build_started"
	ReleaseEventTypeBuildSucceeded       ReleaseEventType = "build_succeeded"
	// ReleaseEventTypeBuildAttemptFailed: a build attempt errored but the build
	// job is still retrying (ImageBuild phase stays Pending until the job's
	// backoff limit is exhausted).
	ReleaseEventTypeBuildAttemptFailed ReleaseEventType = "build_attempt_failed"
	ReleaseEventTypeBuildFailed        ReleaseEventType = "build_failed"
	ReleaseEventTypeResourceWaiting    ReleaseEventType = "resource_waiting"
	ReleaseEventTypeResourceDeploying  ReleaseEventType = "resource_deploying"
	ReleaseEventTypeResourceReady      ReleaseEventType = "resource_ready"
	ReleaseEventTypeResourceFailed     ReleaseEventType = "resource_failed"
	// ReleaseEventTypeResourcePortsClosed: the rollout landed but a declared
	// port is not accepting connections.
	ReleaseEventTypeResourcePortsClosed ReleaseEventType = "resource_ports_closed"
	ReleaseEventTypeResourceTLSIssuing ReleaseEventType = "resource_tls_issuing"
	ReleaseEventTypeResourceTLSReady   ReleaseEventType = "resource_tls_ready"
	ReleaseEventTypeResourceTLSFailed  ReleaseEventType = "resource_tls_failed"
	ReleaseEventTypeReleaseReleased    ReleaseEventType = "release_released"
	ReleaseEventTypeReleaseFailed      ReleaseEventType = "release_failed"
	ReleaseEventTypeReleaseSuperseded  ReleaseEventType = "release_superseded"
	ReleaseEventTypeReleaseCancelled   ReleaseEventType = "release_cancelled"
)

type ReleaseEventScope string

const (
	ReleaseEventScopeRelease  ReleaseEventScope = "release"
	ReleaseEventScopeResource ReleaseEventScope = "resource"
)

type ReleaseEventLevel string

const (
	ReleaseEventLevelInfo    ReleaseEventLevel = "info"
	ReleaseEventLevelSuccess ReleaseEventLevel = "success"
	ReleaseEventLevelWarning ReleaseEventLevel = "warning"
	ReleaseEventLevelError   ReleaseEventLevel = "error"
)

type ReleaseEventSource string

const (
	ReleaseEventSourceHub     ReleaseEventSource = "hub"
	ReleaseEventSourceCluster ReleaseEventSource = "cluster"
)

// Well-known metadata keys.
const (
	ReleaseEventMetaReason  = "reason"
	ReleaseEventMetaCheck   = "check"
	ReleaseEventMetaBuildID = "build_id"
)

// Link kinds.
const ReleaseEventLinkKindBuildLogs = "build_logs"

type ReleaseEventLink struct {
	Kind   string            `json:"kind"`
	Label  string            `json:"label"`
	Target map[string]string `json:"target"`
}

type ReleaseEventLinks []ReleaseEventLink

func (l ReleaseEventLinks) Value() (driver.Value, error) {
	if l == nil {
		return nil, nil
	}
	return json.Marshal(l)
}

func (l *ReleaseEventLinks) Scan(value interface{}) error {
	if value == nil {
		*l = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed for ReleaseEventLinks")
	}
	return json.Unmarshal(b, l)
}

type ReleaseEventMetadata map[string]string

func (m ReleaseEventMetadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *ReleaseEventMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed for ReleaseEventMetadata")
	}
	return json.Unmarshal(b, m)
}

type ReleaseEvent struct {
	ID           string             `gorm:"primaryKey;default:gen_random_uuid()"`
	ReleaseID    string             `gorm:"not null"`
	StackID      string             `gorm:"not null"`
	Sequence     int                `gorm:"not null"`
	OccurredAt   time.Time          `gorm:"not null;default:now()"`
	Source       ReleaseEventSource `gorm:"not null"`
	Scope        ReleaseEventScope  `gorm:"not null"`
	ResourceName *string
	Type         ReleaseEventType     `gorm:"not null"`
	Level        ReleaseEventLevel    `gorm:"not null"`
	Message      string               `gorm:"not null"`
	DedupeKey    string               `gorm:"not null"`
	Links        ReleaseEventLinks    `gorm:"type:jsonb"`
	Metadata     ReleaseEventMetadata `gorm:"type:jsonb"`
	CreatedAt    time.Time
}

func (ReleaseEvent) TableName() string { return "release_events" }
