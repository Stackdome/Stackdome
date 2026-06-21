package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// StackReleaseState tracks the lifecycle of a release.
type StackReleaseState string

const (
	ReleaseStatePending    StackReleaseState = "Pending"
	ReleaseStateInProgress StackReleaseState = "InProgress"
	ReleaseStateReleased   StackReleaseState = "Released"
	ReleaseStateFailed     StackReleaseState = "Failed"
	ReleaseStateSuperseded StackReleaseState = "Superseded"
	ReleaseStateCancelled  StackReleaseState = "Cancelled"
)

func TerminalStackReleaseStates() []StackReleaseState {
	return []StackReleaseState{
		ReleaseStateReleased,
		ReleaseStateFailed,
		ReleaseStateSuperseded,
		ReleaseStateCancelled,
	}
}

// NonGrippingReleaseStates lists release states that should NOT hold
// delete-protection ("grip") on referenced resources.
//
// A release grips its referenced secrets/volumes/addons so that rolling back to
// it remains safe — the resources it needs must still exist. Only Released
// releases are valid rollback targets; Pending and InProgress releases may still
// converge and become the live deployment. Failed, Superseded, and Cancelled
// releases can never be redeployed, so their grip serves no purpose and would
// only block resource deletion until GC ages them out of the retention window.
//
// Grip is derived at query time (LEFT JOIN on stack_releases.state) rather than
// toggled by a write on state transition, so it tracks release lifecycle
// automatically and cannot drift.
var NonGrippingReleaseStates = []StackReleaseState{
	ReleaseStateFailed,
	ReleaseStateSuperseded,
	ReleaseStateCancelled,
}

// Active returns true if the release is still in-flight (not terminal).
func (s StackReleaseState) Active() bool {
	return !s.Terminal()
}

// Terminal returns true if the release has reached a final state.
func (s StackReleaseState) Terminal() bool {
	switch s {
	case ReleaseStateReleased, ReleaseStateFailed, ReleaseStateSuperseded, ReleaseStateCancelled:
		return true
	default:
		return false
	}
}

// StackRelease represents a single atomic deployment attempt for a stack.
type StackRelease struct {
	ID               string            `gorm:"primary_key;default:gen_random_uuid()"`
	StackID          string            `gorm:"not null"`
	Sequence         int               `gorm:"not null"`
	State            StackReleaseState `gorm:"not null"`
	Message          string
	Cause            ReleaseCause     `gorm:"type:jsonb"`
	Snapshot         StackSnapshot    `gorm:"type:jsonb;not null"`
	SnapshotRevision string           `gorm:"not null"`
	Manifest         *ReleaseManifest `gorm:"type:jsonb"`
	ManifestRevision string
	Pins             ReleasePins `gorm:"type:jsonb"`
	RendererVersion  string
	Outcome          *ReleaseOutcome `gorm:"type:jsonb"`
	CreatedBy        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	RenderedAt       *time.Time
	CompletedAt      *time.Time
}

func (StackRelease) TableName() string {
	return "stack_releases"
}

// --- ReleaseCause ---

type ReleaseCauseKind string

const (
	ReleaseCauseManual      ReleaseCauseKind = "manual"
	ReleaseCauseRollback    ReleaseCauseKind = "rollback"
	ReleaseCauseWebhookPush ReleaseCauseKind = "webhook_push"
)

type ReleaseCause struct {
	Kind   ReleaseCauseKind `json:"kind"`
	Detail string           `json:"detail,omitempty"`
}

func (c ReleaseCause) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ReleaseCause) Scan(value interface{}) error {
	if value == nil {
		*c = ReleaseCause{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed for ReleaseCause")
	}
	return json.Unmarshal(b, c)
}

// --- StackSnapshot ---

// StackShellSnapshot captures the identifying fields of a Stack at a point in time.
type StackShellSnapshot struct {
	ID             string      `json:"id"`
	OrganisationID string      `json:"organisation_id"`
	TeamID         string      `json:"team_id"`
	ClusterID      string      `json:"cluster_id"`
	UserID         string      `json:"user_id"`
	Name           string      `json:"name"`
	NamespaceID    string      `json:"namespace_id"`
	Namespace      string      `json:"namespace"`
	Labels         Labels      `json:"labels"`
	Annotations    Annotations `json:"annotations"`
}

// StackSnapshot is an immutable copy of a stack's full spec at the time a release was created.
type StackSnapshot struct {
	Stack       StackShellSnapshot `json:"stack"`
	Resources   []*StackResource   `json:"resources"`
	Volumes     []*Volume          `json:"volumes"`
	Connections StackConnections   `json:"connections"`
	CapturedAt  time.Time          `json:"captured_at"`
}

// NewStackSnapshot deep-copies a Stack into a snapshot. Uses JSON round-trip
// to break pointer sharing so downstream mutations don't corrupt live objects.
func NewStackSnapshot(stack *Stack) (StackSnapshot, error) {
	labels, err := deepCopyJSON(stack.Labels)
	if err != nil {
		return StackSnapshot{}, fmt.Errorf("snapshot labels: %w", err)
	}

	annotations, err := deepCopyJSON(stack.Annotations)
	if err != nil {
		return StackSnapshot{}, fmt.Errorf("snapshot annotations: %w", err)
	}

	shell := StackShellSnapshot{
		ID:             stack.ID,
		OrganisationID: stack.OrganisationID,
		TeamID:         stack.TeamID,
		ClusterID:      stack.ClusterID,
		UserID:         stack.UserID,
		Name:           stack.Name,
		NamespaceID:    stack.NamespaceID,
		Namespace:      stack.Namespace,
		Labels:         labels,
		Annotations:    annotations,
	}

	resources, err := deepCopySlice(stack.StackResources)
	if err != nil {
		return StackSnapshot{}, fmt.Errorf("snapshot resources: %w", err)
	}

	volumes, err := deepCopySlice(stack.Volumes)
	if err != nil {
		return StackSnapshot{}, fmt.Errorf("snapshot volumes: %w", err)
	}

	connections, err := deepCopyJSON(stack.Connections)
	if err != nil {
		return StackSnapshot{}, fmt.Errorf("snapshot connections: %w", err)
	}

	return StackSnapshot{
		Stack:       shell,
		Resources:   resources,
		Volumes:     volumes,
		Connections: connections,
		CapturedAt:  time.Now().UTC(),
	}, nil
}

// ToStack reconstructs a Stack from the snapshot for rendering.
func (s *StackSnapshot) ToStack() *Stack {
	return &Stack{
		ID:             s.Stack.ID,
		OrganisationID: s.Stack.OrganisationID,
		TeamID:         s.Stack.TeamID,
		ClusterID:      s.Stack.ClusterID,
		UserID:         s.Stack.UserID,
		Name:           s.Stack.Name,
		NamespaceID:    s.Stack.NamespaceID,
		Namespace:      s.Stack.Namespace,
		Labels:         s.Stack.Labels,
		Annotations:    s.Stack.Annotations,
		StackResources: s.Resources,
		Volumes:        s.Volumes,
		Connections:    s.Connections,
	}
}

func (s StackSnapshot) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StackSnapshot) Scan(value interface{}) error {
	if value == nil {
		*s = StackSnapshot{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed for StackSnapshot")
	}
	return json.Unmarshal(b, s)
}

// --- ReleaseManifest ---

// ReleaseManifest holds the rendered Kubernetes CRs produced by the renderer.
// Once written it is immutable (enforced by a DB trigger).
type ReleaseManifest struct {
	StackCR            json.RawMessage                  `json:"stack_cr"`
	StackResourceCRs   map[string]json.RawMessage       `json:"stack_resource_crs"`
	ResourceNames      []string                         `json:"resource_names"`
	VolumeBuildSources map[string][]BuildArtifactSource `json:"volume_build_sources,omitempty"`
	ResourceRevisions  map[string]string                `json:"resource_revisions"`
}

func (m ReleaseManifest) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *ReleaseManifest) Scan(value interface{}) error {
	if value == nil {
		*m = ReleaseManifest{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed for ReleaseManifest")
	}
	return json.Unmarshal(b, m)
}

// --- ReleasePins ---

// ReleasePins captures the resolved artifact coordinates for each resource.
type ReleasePins struct {
	// Resource Name -> Resolved artifact src.
	Resources map[string]ResourcePins `json:"resources,omitempty"`
}

// ResourcePins holds the pinned artifact identifiers for a single resource.
type ResourcePins struct {
	GitSHA      string `json:"git_sha,omitempty"`
	VolumeHash  string `json:"volume_hash,omitempty"`
	ImageDigest string `json:"image_digest,omitempty"`
}

func (p ReleasePins) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *ReleasePins) Scan(value interface{}) error {
	if value == nil {
		*p = ReleasePins{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed for ReleasePins")
	}
	return json.Unmarshal(b, p)
}

// --- ReleaseOutcome ---

// ReleaseOutcome records per-resource convergence results.
type ReleaseOutcome struct {
	Resources map[string]ResourceOutcome `json:"resources,omitempty"`
	Duration  time.Duration              `json:"duration"`
}

// ResourceOutcome tracks the final state of a single resource in a release.
type ResourceOutcome struct {
	Phase         StackResourcePhase `json:"phase"`
	ReadyReplicas int32              `json:"ready_replicas"`
	Replicas      int32              `json:"replicas"`
	Message       string             `json:"message,omitempty"`
}

func (o ReleaseOutcome) Value() (driver.Value, error) {
	return json.Marshal(o)
}

func (o *ReleaseOutcome) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed for ReleaseOutcome")
	}
	return json.Unmarshal(b, o)
}

// ReleaseSummary is a lightweight projection used for release GC decisions.
type ReleaseSummary struct {
	ID       string
	Sequence int
	State    StackReleaseState
}

// --- generic deep-copy helpers ---

// deepCopySlice deep-copies a slice of pointers via JSON round-trip.
func deepCopySlice[T any](src []*T) ([]*T, error) {
	if src == nil {
		return nil, nil
	}
	data, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	var dst []*T
	if err := json.Unmarshal(data, &dst); err != nil {
		return nil, err
	}
	return dst, nil
}

// deepCopyJSON deep-copies a value via JSON round-trip.
func deepCopyJSON[T any](src T) (T, error) {
	data, err := json.Marshal(src)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("deep copy marshal: %w", err)
	}
	var dst T
	if err := json.Unmarshal(data, &dst); err != nil {
		var zero T
		return zero, fmt.Errorf("deep copy unmarshal: %w", err)
	}
	return dst, nil
}
