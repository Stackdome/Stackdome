package release

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type releaseService interface {
	InternalGet(ctx context.Context, releaseID string) (*models.StackRelease, *errors.ServiceError)
	InternalListActive(ctx context.Context) ([]*models.StackRelease, *errors.ServiceError)
	MarkRendering(ctx context.Context, id string) (bool, *errors.ServiceError)
	MarkApplyingDirect(ctx context.Context, id string) (bool, *errors.ServiceError)
	SaveManifest(ctx context.Context, id string, m *models.ReleaseManifest, rev string, pins models.ReleasePins, rendererVersion string) (bool, *errors.ServiceError)
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
