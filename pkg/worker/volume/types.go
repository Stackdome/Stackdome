//go:generate mockgen -source=types.go -destination=types_mock.go -package=volume
package volume

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

type volumeService interface {
	InternalGet(ctx context.Context, ID string) (*models.Volume, *errors.ServiceError)
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
}
