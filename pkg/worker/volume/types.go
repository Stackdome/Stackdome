package volume

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type volumeService interface {
	InternalGet(ctx context.Context, ID string) (*models.Volume, *errors.ServiceError)
	ListVolumesUsedByStack(ctx context.Context, stackID string) ([]*models.Volume, *errors.ServiceError)
}

type stackService interface {
	InternalGetStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError)
}

type stackVolumeStore interface {
	GetByVolumeID(ctx context.Context, volumeID string) (*models.StackVolume, *errors.ServiceError)
}
