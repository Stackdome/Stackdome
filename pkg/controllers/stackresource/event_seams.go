package workspaceresource

import (
	"context"

	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -source=event_seams.go -destination=event_seams_mock_test.go -package=workspaceresource

// releaseActiveChecker is a narrow seam to look up the active release for a
// stack without importing the full StackReleaseService (which would close an
// import cycle). Mirrors the pkg/controllers/stack seam of the same name.
type releaseActiveChecker interface {
	InternalGetActiveByStackID(ctx context.Context, stackID string) (*models.StackRelease, *apperrors.ServiceError)
}

// resourceEventRecorder is the narrow slice of services.ReleaseEventRecorder the
// controller needs. It is declared locally because the recorder's mock lives
// in-package in pkg/services (an import cycle keeps it out of pkg/mocks).
type resourceEventRecorder interface {
	RecordResourceEvent(ctx context.Context, release *models.StackRelease, resourceName string, eventType models.ReleaseEventType, reason string) *apperrors.ServiceError
}
