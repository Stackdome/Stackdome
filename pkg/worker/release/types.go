package release

import (
	"context"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type subReconcilerResult struct {
	resultNil          bool
	resultStop         bool
	resultRequeue      bool
	resultRequeueAfter *time.Duration
}

var (
	resultNil          = subReconcilerResult{resultNil: true}
	resultStop         = subReconcilerResult{resultStop: true}
	resultRequeue      = subReconcilerResult{resultRequeue: true}
	resultRequeueAfter = func(t time.Duration) subReconcilerResult {
		return subReconcilerResult{resultRequeueAfter: &t}
	}
)

type subReconciler interface {
	Reconcile(ctx context.Context, release *models.StackRelease) (subReconcilerResult, error)
	Name() string
}

type releaseService interface {
	InternalGet(ctx context.Context, releaseID string) (*models.StackRelease, *errors.ServiceError)
	InternalGetActiveByStackID(ctx context.Context, stackID string) (*models.StackRelease, *errors.ServiceError)
	InternalListActive(ctx context.Context) ([]*models.StackRelease, *errors.ServiceError)
	MarkInProgress(ctx context.Context, id string) (bool, *errors.ServiceError)
	SaveManifest(ctx context.Context, id string, m *models.ReleaseManifest, rev string, pins models.ReleasePins, rendererVersion string) (bool, *errors.ServiceError)
	MarkSuperseded(ctx context.Context, id string, reason string) (bool, *errors.ServiceError)
	MarkReleased(ctx context.Context, id string, outcome models.ReleaseOutcome) (bool, *errors.ServiceError)
	MarkFailed(ctx context.Context, id string, message string, outcome *models.ReleaseOutcome) (bool, *errors.ServiceError)
	AppendImageDigests(ctx context.Context, id string, digests map[string]string) *errors.ServiceError
}

type stackService interface {
	InternalGetStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError)
	UpdateStackCrRevision(ctx context.Context, ID string, revision string) *errors.ServiceError
}

type secretService interface {
	InternalGetByID(ctx context.Context, id string) (*models.Secret, *errors.ServiceError)
}

type postgresAddonService interface {
	InternalGetCredentials(ctx context.Context, addonID, database string, superuser bool) (*models.PostgresCredentials, *errors.ServiceError)
}

type volumeService interface {
	ListVolumesUsedByStack(ctx context.Context, stackID string) ([]*models.Volume, *errors.ServiceError)
}
