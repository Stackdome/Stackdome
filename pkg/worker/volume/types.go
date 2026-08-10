//go:generate mockgen -source=types.go -destination=types_mock.go -package=volume
package volume

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

type volumeService interface {
	InternalGet(ctx context.Context, ID string) (*models.Volume, *errors.ServiceError)
	InternalList(ctx context.Context, ids []string) ([]*models.Volume, *errors.ServiceError)
	InternalListNotReady(ctx context.Context) ([]*models.Volume, *errors.ServiceError)
	ListVolumesUsedByStack(ctx context.Context, stackID string) ([]*models.Volume, *errors.ServiceError)
}

type stackService interface {
	InternalGetStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError)
}

type stackVolumeStore interface {
	GetByVolumeID(ctx context.Context, volumeID string) (*models.StackVolume, *errors.ServiceError)
}

type releaseService interface {
	InternalGet(ctx context.Context, releaseID string) (*models.StackRelease, *errors.ServiceError)
	InternalResolveAuthoritativeWorkloadRelease(ctx context.Context, stack *models.Stack) (*models.StackRelease, *errors.ServiceError)
	InternalListAuthoritativeWorkload(ctx context.Context) (*models.WorkloadAuthorityScan, *errors.ServiceError)
}

type referenceService interface {
	InternalListReleaseReferents(ctx context.Context, releaseIDs []string, referentType models.ReferentType) ([]models.ResourceReference, *errors.ServiceError)
}
