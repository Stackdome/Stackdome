package services

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
)

// ReleaseEventRecorder's mock lives in-package (not pkg/mocks): the interface's
// Record method references ReleaseEventInput, so a pkg/mocks mock would import
// services and close an import cycle (services -> auth -> mocks -> services).
// This mirrors the NamespaceService pattern above.
//go:generate mockgen -source=release_event_recorder.go -destination=release_event_recorder_mock_test.go -package=services -self_package=github.com/Stackdome/stackdome/pkg/services

type ReleaseEventInput struct {
	Release      *models.StackRelease
	ResourceName string // required for resource-scoped types
	Type         models.ReleaseEventType
	Message      string
	DedupeKey    string // defaulted from type/resource when empty
	Links        models.ReleaseEventLinks
	Metadata     models.ReleaseEventMetadata
}

type ReleaseEventRecorder interface {
	Record(ctx context.Context, input ReleaseEventInput) (*models.ReleaseEvent, *errors.ServiceError)
	RecordReleaseCreated(ctx context.Context, release *models.StackRelease) *errors.ServiceError
	RecordReleaseChecksStarted(ctx context.Context, release *models.StackRelease) *errors.ServiceError
	RecordReleaseCheckFailed(ctx context.Context, release *models.StackRelease, resourceName, check, reason string) *errors.ServiceError
	RecordReleaseChecksPassed(ctx context.Context, release *models.StackRelease) *errors.ServiceError
	RecordReleaseStarted(ctx context.Context, release *models.StackRelease) *errors.ServiceError
	RecordBuildEvent(ctx context.Context, release *models.StackRelease, resourceName string, eventType models.ReleaseEventType, buildID string, attribution string, reason string) *errors.ServiceError
	RecordResourceEvent(ctx context.Context, release *models.StackRelease, resourceName string, eventType models.ReleaseEventType, reason string) *errors.ServiceError
	RecordReleaseTerminal(ctx context.Context, release *models.StackRelease, state models.StackReleaseState, message string) *errors.ServiceError
}

// eventTypeMeta is the single source of truth for scope + level per type.
var eventTypeMeta = map[models.ReleaseEventType]struct {
	Scope models.ReleaseEventScope
	Level models.ReleaseEventLevel
}{
	models.ReleaseEventTypeReleaseCreated:       {models.ReleaseEventScopeRelease, models.ReleaseEventLevelInfo},
	models.ReleaseEventTypeReleaseChecksStarted: {models.ReleaseEventScopeRelease, models.ReleaseEventLevelInfo},
	models.ReleaseEventTypeReleaseCheckFailed:   {models.ReleaseEventScopeResource, models.ReleaseEventLevelError},
	models.ReleaseEventTypeReleaseChecksPassed:  {models.ReleaseEventScopeRelease, models.ReleaseEventLevelSuccess},
	models.ReleaseEventTypeReleaseStarted:       {models.ReleaseEventScopeRelease, models.ReleaseEventLevelInfo},
	models.ReleaseEventTypeBuildQueued:          {models.ReleaseEventScopeResource, models.ReleaseEventLevelInfo},
	models.ReleaseEventTypeBuildStarted:         {models.ReleaseEventScopeResource, models.ReleaseEventLevelInfo},
	models.ReleaseEventTypeBuildSucceeded:       {models.ReleaseEventScopeResource, models.ReleaseEventLevelSuccess},
	models.ReleaseEventTypeBuildFailed:          {models.ReleaseEventScopeResource, models.ReleaseEventLevelError},
	models.ReleaseEventTypeResourceWaiting:      {models.ReleaseEventScopeResource, models.ReleaseEventLevelWarning},
	models.ReleaseEventTypeResourceDeploying:    {models.ReleaseEventScopeResource, models.ReleaseEventLevelInfo},
	models.ReleaseEventTypeResourceReady:        {models.ReleaseEventScopeResource, models.ReleaseEventLevelSuccess},
	models.ReleaseEventTypeResourceFailed:       {models.ReleaseEventScopeResource, models.ReleaseEventLevelError},
	models.ReleaseEventTypeReleaseReleased:      {models.ReleaseEventScopeRelease, models.ReleaseEventLevelSuccess},
	models.ReleaseEventTypeReleaseFailed:        {models.ReleaseEventScopeRelease, models.ReleaseEventLevelError},
	models.ReleaseEventTypeReleaseSuperseded:    {models.ReleaseEventScopeRelease, models.ReleaseEventLevelWarning},
	models.ReleaseEventTypeReleaseCancelled:     {models.ReleaseEventScopeRelease, models.ReleaseEventLevelWarning},
}

var terminalEventForState = map[models.StackReleaseState]models.ReleaseEventType{
	models.ReleaseStateReleased:   models.ReleaseEventTypeReleaseReleased,
	models.ReleaseStateFailed:     models.ReleaseEventTypeReleaseFailed,
	models.ReleaseStateSuperseded: models.ReleaseEventTypeReleaseSuperseded,
	models.ReleaseStateCancelled:  models.ReleaseEventTypeReleaseCancelled,
}

type ReleaseEventRecorderSpec struct {
	Store stores.ReleaseEventStore
}

type releaseEventRecorder struct {
	store stores.ReleaseEventStore
}

func NewReleaseEventRecorder(spec ReleaseEventRecorderSpec) ReleaseEventRecorder {
	if spec.Store == nil {
		panic("ReleaseEventRecorder requires a ReleaseEventStore")
	}
	return &releaseEventRecorder{store: spec.Store}
}

func (r *releaseEventRecorder) buildEvent(input ReleaseEventInput) (*models.ReleaseEvent, *errors.ServiceError) {
	if input.Release == nil {
		return nil, errors.GeneralError("release event input requires a release")
	}
	meta, ok := eventTypeMeta[input.Type]
	if !ok {
		return nil, errors.GeneralError("unknown release event type %q", input.Type)
	}
	if meta.Scope == models.ReleaseEventScopeResource && input.ResourceName == "" {
		return nil, errors.GeneralError("release event type %q requires a resource name", input.Type)
	}
	event := &models.ReleaseEvent{
		ReleaseID: input.Release.ID,
		StackID:   input.Release.StackID,
		Source:    models.ReleaseEventSourceHub,
		Scope:     meta.Scope,
		Type:      input.Type,
		Level:     meta.Level,
		Message:   input.Message,
		DedupeKey: input.DedupeKey,
		Links:     input.Links,
		Metadata:  input.Metadata,
	}
	if input.ResourceName != "" {
		event.ResourceName = &input.ResourceName
	}
	if event.DedupeKey == "" {
		event.DedupeKey = string(input.Type)
		if input.ResourceName != "" {
			event.DedupeKey = fmt.Sprintf("%s:%s", input.Type, input.ResourceName)
		}
	}
	return event, nil
}

func (r *releaseEventRecorder) Record(ctx context.Context, input ReleaseEventInput) (*models.ReleaseEvent, *errors.ServiceError) {
	event, serr := r.buildEvent(input)
	if serr != nil {
		return nil, serr
	}
	return r.store.Insert(ctx, event)
}

// RecordReleaseCreated must run inside the release-creation transaction.
func (r *releaseEventRecorder) RecordReleaseCreated(ctx context.Context, release *models.StackRelease) *errors.ServiceError {
	event, serr := r.buildEvent(ReleaseEventInput{
		Release:   release,
		Type:      models.ReleaseEventTypeReleaseCreated,
		Message:   "Release created",
		DedupeKey: "release:created",
	})
	if serr != nil {
		return serr
	}
	_, serr = r.store.InsertWithTx(ctx, event)
	return serr
}

func (r *releaseEventRecorder) RecordReleaseChecksStarted(ctx context.Context, release *models.StackRelease) *errors.ServiceError {
	_, serr := r.Record(ctx, ReleaseEventInput{
		Release:   release,
		Type:      models.ReleaseEventTypeReleaseChecksStarted,
		Message:   "Validating release",
		DedupeKey: "release:checks_started",
	})
	return serr
}

func (r *releaseEventRecorder) RecordReleaseCheckFailed(ctx context.Context, release *models.StackRelease, resourceName, check, reason string) *errors.ServiceError {
	_, serr := r.Record(ctx, ReleaseEventInput{
		Release:      release,
		ResourceName: resourceName,
		Type:         models.ReleaseEventTypeReleaseCheckFailed,
		Message:      fmt.Sprintf("Check %s failed for %s: %s", check, resourceName, reason),
		DedupeKey:    fmt.Sprintf("release:check_failed:%s:%s", resourceName, check),
		Metadata: models.ReleaseEventMetadata{
			models.ReleaseEventMetaCheck:  check,
			models.ReleaseEventMetaReason: reason,
		},
	})
	return serr
}

func (r *releaseEventRecorder) RecordReleaseChecksPassed(ctx context.Context, release *models.StackRelease) *errors.ServiceError {
	_, serr := r.Record(ctx, ReleaseEventInput{
		Release:   release,
		Type:      models.ReleaseEventTypeReleaseChecksPassed,
		Message:   "Release checks passed",
		DedupeKey: "release:checks_passed",
	})
	return serr
}

func (r *releaseEventRecorder) RecordReleaseStarted(ctx context.Context, release *models.StackRelease) *errors.ServiceError {
	_, serr := r.Record(ctx, ReleaseEventInput{
		Release:   release,
		Type:      models.ReleaseEventTypeReleaseStarted,
		Message:   "Deploying release",
		DedupeKey: "release:started",
	})
	return serr
}

func (r *releaseEventRecorder) RecordBuildEvent(ctx context.Context, release *models.StackRelease, resourceName string, eventType models.ReleaseEventType, buildID string, attribution string, reason string) *errors.ServiceError {
	var message string
	switch eventType {
	case models.ReleaseEventTypeBuildStarted:
		message = fmt.Sprintf("Building %s", resourceName)
	case models.ReleaseEventTypeBuildSucceeded:
		message = fmt.Sprintf("Build succeeded for %s", resourceName)
	case models.ReleaseEventTypeBuildFailed:
		message = fmt.Sprintf("Build failed for %s", resourceName)
		if reason != "" {
			message = fmt.Sprintf("%s: %s", message, reason)
		}
	default:
		return errors.GeneralError("unsupported build event type %q", eventType)
	}
	metadata := models.ReleaseEventMetadata{models.ReleaseEventMetaBuildID: buildID}
	if attribution != "" {
		metadata[models.ReleaseEventMetaAttribution] = attribution
	}
	if reason != "" {
		metadata[models.ReleaseEventMetaReason] = reason
	}
	var links models.ReleaseEventLinks
	if buildID != "" {
		links = models.ReleaseEventLinks{{
			Kind:  models.ReleaseEventLinkKindBuildLogs,
			Label: "View build logs",
			Target: map[string]string{
				"build_id":      buildID,
				"resource_name": resourceName,
			},
		}}
	}
	_, serr := r.Record(ctx, ReleaseEventInput{
		Release:      release,
		ResourceName: resourceName,
		Type:         eventType,
		Message:      message,
		DedupeKey:    fmt.Sprintf("build:%s:%s", resourceName, eventType),
		Links:        links,
		Metadata:     metadata,
	})
	return serr
}

func (r *releaseEventRecorder) RecordResourceEvent(ctx context.Context, release *models.StackRelease, resourceName string, eventType models.ReleaseEventType, reason string) *errors.ServiceError {
	var message string
	switch eventType {
	case models.ReleaseEventTypeResourceWaiting:
		message = fmt.Sprintf("%s is waiting on dependencies", resourceName)
		if reason != "" {
			message = fmt.Sprintf("%s is waiting: %s", resourceName, reason)
		}
	case models.ReleaseEventTypeResourceDeploying:
		message = fmt.Sprintf("Deploying %s", resourceName)
	case models.ReleaseEventTypeResourceReady:
		message = fmt.Sprintf("%s is ready", resourceName)
	case models.ReleaseEventTypeResourceFailed:
		message = fmt.Sprintf("%s failed to start", resourceName)
		if reason != "" {
			message = fmt.Sprintf("%s: %s", message, reason)
		}
	default:
		return errors.GeneralError("unsupported resource event type %q", eventType)
	}
	var metadata models.ReleaseEventMetadata
	if reason != "" {
		metadata = models.ReleaseEventMetadata{models.ReleaseEventMetaReason: reason}
	}
	_, serr := r.Record(ctx, ReleaseEventInput{
		Release:      release,
		ResourceName: resourceName,
		Type:         eventType,
		Message:      message,
		DedupeKey:    fmt.Sprintf("resource:%s:%s", resourceName, eventType),
		Metadata:     metadata,
	})
	return serr
}

func (r *releaseEventRecorder) RecordReleaseTerminal(ctx context.Context, release *models.StackRelease, state models.StackReleaseState, message string) *errors.ServiceError {
	eventType, ok := terminalEventForState[state]
	if !ok {
		return errors.GeneralError("state %q is not a terminal release state", state)
	}
	_, serr := r.Record(ctx, ReleaseEventInput{
		Release:   release,
		Type:      eventType,
		Message:   message,
		DedupeKey: fmt.Sprintf("release:terminal:%s", state),
	})
	return serr
}
